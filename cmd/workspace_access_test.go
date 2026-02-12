package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdWorkspaceAccessListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/access") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		items := testutil.FixtureWorkspaceAccessList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, items)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-access", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wa1a2b3c") {
		t.Errorf("expected output to contain 'wa1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "platform-engineering") {
		t.Errorf("expected output to contain 'platform-engineering', got: %s", out)
	}
}

func TestCmdWorkspaceAccessGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/access/wa-001") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		item := testutil.FixtureWorkspaceAccess()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, item)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-access", "get",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--id", "wa-001",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wa1a2b3c") {
		t.Errorf("expected output to contain 'wa1a2b3c', got: %s", out)
	}
}

func TestCmdWorkspaceAccessCreateE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/access") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		item := &terrakube.WorkspaceAccess{ID: "wa-new", Name: "devs", ManageState: true}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, item)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-access", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--name", "devs",
		"--manage-state",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wa-new") {
		t.Errorf("expected output to contain 'wa-new', got: %s", out)
	}
}

func TestCmdWorkspaceAccessDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/b2c3d4e5-f6a7-8901-bcde-f12345678901/access/wa-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-access", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		"--id", "wa-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdWorkspaceAccessListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("workspace-access", "list", "--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901")
	if err == nil {
		t.Fatal("expected error for workspace-access list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
