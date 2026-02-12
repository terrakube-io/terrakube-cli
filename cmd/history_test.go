package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdHistoryListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/history") {
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
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/history/hist-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"history", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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

	_, err := executeCommand("history", "list", "--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901")
	if err == nil {
		t.Fatal("expected error for history list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
