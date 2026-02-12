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
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/webhook/b8c9d0e1-f2a3-4567-bcde-678901234567/events") {
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
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--webhook-id", "b8c9d0e1-f2a3-4567-bcde-678901234567",
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
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/webhook/b8c9d0e1-f2a3-4567-bcde-678901234567/event/we-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"webhook-event", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--webhook-id", "b8c9d0e1-f2a3-4567-bcde-678901234567",
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

	_, err := executeCommand("webhook-event", "list", "--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901", "--webhook-id", "b8c9d0e1-f2a3-4567-bcde-678901234567")
	if err == nil {
		t.Fatal("expected error for webhook-event list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
