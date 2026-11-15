package cmd

import (
	"net/http"
	"testing"

	"terrakube/internal/exit"
)

// The hand-written commands surface API errors (exit 1) instead of hiding them.
func TestHandWrittenCommandsReportAPIErrors(t *testing.T) {
	commands := [][]string{
		{"job", "wait", "42", "-o", runOrg},
		{"job", "logs", "42", "-o", runOrg},
		{"job", "approve", "42", "-o", runOrg},
		{"job", "cancel", "42", "-o", runOrg},
		{"run", "apply", "-o", runOrg, "-w", runWs},
		{"run", "plan", "-o", runOrg, "-w", runWs, "-t", "Plan"},
		{"run", "plan", "-o", runOrg, "-w", runWs, "-t", runTpl},
		{"workspace", "lock", runWs, "-o", runOrg},
		{"workspace", "unlock", "prod", "-o", runOrg},
		{"job", "logs", "42", "-o", "my-org"},
	}
	for _, args := range commands {
		t.Run(args[0]+" "+args[1], func(t *testing.T) {
			resetGlobalFlags()
			ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"errors":[{"detail":"boom"}]}`))
			}))
			defer ts.Close()
			if _, err := executeCommand(args...); exit.Code(err) != exit.Failure {
				t.Errorf("%v: expected an API error (exit 1), got %v", args, err)
			}
		})
	}
}

// Without credentials they fail before any request.
func TestHandWrittenCommandsNeedLogin(t *testing.T) {
	for _, args := range [][]string{
		{"job", "wait", "42", "-o", runOrg},
		{"job", "logs", "42", "-o", runOrg},
		{"job", "reject", "42", "-o", runOrg},
		{"run", "destroy", "-o", runOrg, "-w", runWs},
		{"workspace", "lock", runWs, "-o", runOrg},
	} {
		resetGlobalFlags()
		if _, err := executeCommand(args...); err == nil {
			t.Errorf("%v: expected not logged in", args)
		}
	}
}
