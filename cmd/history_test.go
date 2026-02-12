package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdHistoryListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/history") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		entries := []*terrakube.History{
			{ID: "hist-1", JobReference: "job-ref-001", Output: "state output data", Serial: 1},
			{ID: "hist-2", JobReference: "job-ref-002", Output: "updated state", Serial: 2},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, entries)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"history", "list",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "hist-1") {
		t.Errorf("expected output to contain 'hist-1', got: %s", out)
	}
	if !strings.Contains(out, "job-ref-001") {
		t.Errorf("expected output to contain 'job-ref-001', got: %s", out)
	}
}

func TestCmdHistoryDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/history/hist-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"history", "delete",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--id", "hist-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdHistoryListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("history", "list", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for history list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
