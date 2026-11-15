package cmd

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// jobActionServer serves job 42 in status and records PATCH bodies.
func jobActionServer(t *testing.T, status string) func() []string {
	t.Helper()
	var mu sync.Mutex
	var patches []string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if !strings.HasSuffix(r.URL.Path, "/organization/"+runOrg+"/job/42") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		current := status
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			mu.Lock()
			patches = append(patches, string(b))
			mu.Unlock()
			current = "patched"
		}
		_ = jsonapi.MarshalPayload(w, &terrakube.Job{ID: "42", Status: current})
	}))
	t.Cleanup(ts.Close)
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), patches...)
	}
}

// AC-27, AC-28: each action sets its status when the current one allows it.
func TestCmdJobActions(t *testing.T) {
	for _, tt := range []struct{ action, from, sets string }{
		{"approve", "waitingApproval", "approved"},
		{"reject", "waitingApproval", "rejected"},
		{"reject", "running", "rejected"},
		{"cancel", "pending", "cancelled"},
		{"cancel", "running", "cancelled"},
	} {
		t.Run(tt.action+"/"+tt.from, func(t *testing.T) {
			resetGlobalFlags()
			patches := jobActionServer(t, tt.from)
			out, err := executeCommand("job", tt.action, "42", "-o", runOrg)
			if err != nil {
				t.Fatal(err)
			}
			got := patches()
			if len(got) != 1 || !strings.Contains(got[0], `"attributes":{"status":"`+tt.sets+`"}`) {
				t.Fatalf("expected one PATCH setting %s, got %v", tt.sets, got)
			}
			if !strings.Contains(out, `"id": "42"`) {
				t.Errorf("expected the job printed, got %q", out)
			}
		})
	}
}

// AC-27, AC-28: a status that does not allow the action sends no write.
func TestCmdJobActionsRefused(t *testing.T) {
	for _, tt := range []struct{ action, from, want string }{
		{"approve", "running", "job 42 is running: approve needs waitingApproval"},
		{"approve", "completed", "job 42 is completed: approve needs waitingApproval"},
		{"reject", "failed", "job 42 already ended failed"},
		{"cancel", "completed", "job 42 already ended completed"},
		{"cancel", "noChanges", "job 42 already ended noChanges"},
	} {
		t.Run(tt.action+"/"+tt.from, func(t *testing.T) {
			resetGlobalFlags()
			patches := jobActionServer(t, tt.from)
			_, err := executeCommand("job", tt.action, "42", "-o", runOrg)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
			if got := patches(); len(got) != 0 {
				t.Errorf("expected no write, got %v", got)
			}
		})
	}
}

// AC-29: job list --status filters on the server, combined with --filter.
func TestCmdJobListStatus(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"--status", "waitingApproval"}, `status=="waitingApproval"`},
		{[]string{"--status", "running", "--filter", `commitId=="abc"`}, `commitId=="abc";status=="running"`},
	} {
		resetGlobalFlags()
		var filter string
		ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			filter = r.URL.Query().Get("filter[job]")
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, []*terrakube.Job{{ID: "42", Status: "waitingApproval"}})
		}))
		_, err := executeCommand(append([]string{"job", "list", "-o", runOrg, "-w", runWs}, tt.args...)...)
		ts.Close()
		if err != nil {
			t.Fatal(err)
		}
		if filter != tt.want {
			t.Errorf("%v: filter = %q, want %q", tt.args, filter, tt.want)
		}
	}
}
