package cmd

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/jsonapi"
	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
)

// jobStatusServer answers job GETs with the given statuses in order (the last repeats).
func jobStatusServer(t *testing.T, statuses ...string) *int {
	t.Helper()
	var mu sync.Mutex
	calls := 0
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/42") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		mu.Lock()
		s := statuses[min(calls, len(statuses)-1)]
		calls++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, &terrakube.Job{ID: "42", Status: s})
	}))
	t.Cleanup(ts.Close)
	return &calls
}

func fastPolling(t *testing.T) {
	t.Helper()
	old := jobPollInterval
	jobPollInterval = time.Millisecond
	t.Cleanup(func() { jobPollInterval = old })
}

// AC-20: wait until a final status, print it, exit by the status.
func TestCmdJobWait(t *testing.T) {
	for _, tt := range []struct {
		statuses []string
		code     int
	}{
		{[]string{"queue", "running", "completed"}, exit.OK},
		{[]string{"running", "noChanges"}, exit.OK},
		{[]string{"running", "failed"}, exit.JobFailed},
		{[]string{"rejected"}, exit.JobFailed},
		{[]string{"cancelled"}, exit.JobFailed},
	} {
		final := tt.statuses[len(tt.statuses)-1]
		t.Run(final, func(t *testing.T) {
			resetGlobalFlags()
			fastPolling(t)
			calls := jobStatusServer(t, tt.statuses...)

			out, err := executeCommand("job", "wait", "42", "-o", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
			if got := exit.Code(err); got != tt.code {
				t.Fatalf("exit code %d, want %d (err %v)", got, tt.code, err)
			}
			if !strings.Contains(out, `"status": "`+final+`"`) {
				t.Errorf("expected the final job printed, got %q", out)
			}
			if *calls != len(tt.statuses) {
				t.Errorf("polled %d times, want %d", *calls, len(tt.statuses))
			}
		})
	}
}

// AC-21: still running after --timeout: exit 3.
func TestCmdJobWaitTimeout(t *testing.T) {
	resetGlobalFlags()
	fastPolling(t)
	jobStatusServer(t, "running")

	_, err := executeCommand("job", "wait", "42", "-o", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--timeout", "20ms")
	if exit.Code(err) != exit.Timeout || !strings.Contains(err.Error(), "timed out after 20ms, job still running") {
		t.Fatalf("expected a timeout, got %v (code %d)", err, exit.Code(err))
	}
}

// AC-22: a negative timeout fails before any request.
func TestCmdJobWaitNegativeTimeout(t *testing.T) {
	resetGlobalFlags()
	calls := jobStatusServer(t, "running")

	_, err := executeCommand("job", "wait", "42", "-o", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--timeout", "-1s")
	if err == nil || err.Error() != "--timeout must not be negative" || exit.Code(err) != exit.Failure {
		t.Fatalf("unexpected error: %v", err)
	}
	if *calls != 0 {
		t.Errorf("expected no request, got %d", *calls)
	}
}

// The process exit code comes from the error.
func TestRunReturnsTheExitCode(t *testing.T) {
	resetGlobalFlags()
	probe := &cobra.Command{Use: "exit-probe", RunE: func(*cobra.Command, []string) error {
		return &exit.Error{Code: exit.Timeout, Err: errors.New("slow")}
	}}
	rootCmd.AddCommand(probe)
	t.Cleanup(func() { rootCmd.RemoveCommand(probe) })
	rootCmd.SetArgs([]string{"exit-probe"})
	if code := run(); code != exit.Timeout {
		t.Errorf("run() = %d, want %d", code, exit.Timeout)
	}
}
