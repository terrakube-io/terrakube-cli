package cmd

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdWebhookListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/webhook") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		items := testutil.FixtureWebhookList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, items)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"webhook", "list",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wh1a2b3c") {
		t.Errorf("expected output to contain 'wh1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "PUSH") {
		t.Errorf("expected output to contain 'PUSH', got: %s", out)
	}
}

func TestCmdWebhookDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/webhook/wh-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"webhook", "delete",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--id", "wh-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdWebhookListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("webhook", "list", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for webhook list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
