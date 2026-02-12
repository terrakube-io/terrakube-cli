package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdWebhookEventListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/webhook/wh-456/events") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		events := []*terrakube.WebhookEvent{
			{ID: "we-1", Branch: "main", Event: "PUSH", Path: "/", Priority: 1, TemplateID: "tpl-abc"},
			{ID: "we-2", Branch: "develop", Event: "TAG", Path: "/modules", Priority: 2, TemplateID: "tpl-def"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, events)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"webhook-event", "list",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--webhook-id", "wh-456",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "we-1") {
		t.Errorf("expected output to contain 'we-1', got: %s", out)
	}
	if !strings.Contains(out, "PUSH") {
		t.Errorf("expected output to contain 'PUSH', got: %s", out)
	}
}

func TestCmdWebhookEventDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/webhook/wh-456/event/we-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"webhook-event", "delete",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--webhook-id", "wh-456",
		"--id", "we-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdWebhookEventListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("webhook-event", "list", "--workspace-id", "ws-123", "--webhook-id", "wh-456")
	if err == nil {
		t.Fatal("expected error for webhook-event list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
