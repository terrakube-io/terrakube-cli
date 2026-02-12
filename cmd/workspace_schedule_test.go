package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdWorkspaceScheduleListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/schedule") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		schedules := []*terrakube.WorkspaceSchedule{
			{ID: "sched-1", Schedule: "0 0 * * *", TemplateID: "tpl-abc-123"},
			{ID: "sched-2", Schedule: "0 12 * * MON-FRI", TemplateID: "tpl-def-456"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, schedules)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-schedule", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "sched-1") {
		t.Errorf("expected output to contain 'sched-1', got: %s", out)
	}
	if !strings.Contains(out, "0 0 * * *") {
		t.Errorf("expected output to contain '0 0 * * *', got: %s", out)
	}
}

func TestCmdWorkspaceScheduleDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/schedule/sched-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-schedule", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--id", "sched-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdWorkspaceScheduleListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("workspace-schedule", "list", "--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901")
	if err == nil {
		t.Fatal("expected error for workspace-schedule list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
