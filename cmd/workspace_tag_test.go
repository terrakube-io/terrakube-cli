package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"
	"github.com/spf13/cobra"
)

func TestCmdWorkspaceTagListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/workspaceTag") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		tags := []*terrakube.WorkspaceTag{
			{ID: "wt-1", TagID: "tag-prod-001"},
			{ID: "wt-2", TagID: "tag-staging-002"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, tags)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-tag", "list",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wt-1") {
		t.Errorf("expected output to contain 'wt-1', got: %s", out)
	}
	if !strings.Contains(out, "tag-prod-001") {
		t.Errorf("expected output to contain 'tag-prod-001', got: %s", out)
	}
}

func TestCmdWorkspaceTagCreateE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/workspaceTag") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		tag := &terrakube.WorkspaceTag{ID: "wt-new", TagID: "tag-prod-001"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, tag)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-tag", "create",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--tag-id", "tag-prod-001",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wt-new") {
		t.Errorf("expected output to contain 'wt-new', got: %s", out)
	}
}

func TestCmdWorkspaceTagDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/workspace/ws-123/workspaceTag/wt-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-tag", "delete",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--id", "wt-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdWorkspaceTagListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	// Need a valid server so newClient() doesn't os.Exit(1) on empty URL.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("workspace-tag", "list", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for workspace-tag list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}

func TestCmdWorkspaceTagListMissingWorkspace(t *testing.T) {
	resetGlobalFlags()

	// Provide a server so the org-id flag passes, then expect workspace error.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("workspace-tag", "list", "--organization-id", "org-abc")
	if err == nil {
		t.Fatal("expected error for workspace-tag list without workspace flag, got nil")
	}
	if !strings.Contains(err.Error(), "workspace-id") && !strings.Contains(err.Error(), "workspace-name") {
		t.Errorf("expected error to mention workspace-id or workspace-name, got: %v", err)
	}
}

func TestCmdWorkspaceTagNameResolution(t *testing.T) {
	resetGlobalFlags()

	requestPaths := []string{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPaths = append(requestPaths, r.URL.RequestURI())

		// Request 1: org name resolution — GET /api/v1/organization?filter[organization]=name==acme
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/api/v1/organization") && !strings.Contains(r.URL.Path, "/workspace") {
			orgs := []*terrakube.Organization{{ID: "org-resolved-id", Name: "acme"}}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, orgs)
			return
		}

		// Request 2: workspace name resolution — GET /api/v1/organization/org-resolved-id/workspace?filter[workspace]=name==prod
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "organization/org-resolved-id/workspace") && !strings.Contains(r.URL.Path, "workspaceTag") {
			wss := []*terrakube.Workspace{{ID: "ws-resolved-id", Name: "prod"}}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, wss)
			return
		}

		// Request 3: actual workspace tag list — GET /api/v1/organization/org-resolved-id/workspace/ws-resolved-id/workspaceTag
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "organization/org-resolved-id/workspace/ws-resolved-id/workspaceTag") {
			tags := []*terrakube.WorkspaceTag{{ID: "wt-resolved", TagID: "tag-001"}}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, tags)
			return
		}

		t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
		w.WriteHeader(http.StatusNotFound)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace-tag", "list",
		"--organization-name", "acme",
		"--workspace-name", "prod",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "wt-resolved") {
		t.Errorf("expected output to contain 'wt-resolved', got: %s", out)
	}

	if len(requestPaths) != 3 {
		t.Fatalf("expected 3 requests (org resolve, workspace resolve, tag list), got %d: %v", len(requestPaths), requestPaths)
	}

	// Verify the cascade: org resolved first, then workspace uses resolved org ID
	if !strings.Contains(requestPaths[0], "/api/v1/organization") {
		t.Errorf("first request should be org resolution, got: %s", requestPaths[0])
	}
	if !strings.Contains(requestPaths[1], "organization/org-resolved-id/workspace") {
		t.Errorf("second request should use resolved org ID for workspace resolution, got: %s", requestPaths[1])
	}
	if !strings.Contains(requestPaths[2], "organization/org-resolved-id/workspace/ws-resolved-id/workspaceTag") {
		t.Errorf("third request should use both resolved IDs, got: %s", requestPaths[2])
	}
}

func TestCmdWorkspaceTagAlias(t *testing.T) {
	resetGlobalFlags()

	// Need a valid server so newClient() doesn't os.Exit(1).
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	// Verify the alias routes correctly by getting a required-flag error
	_, err := executeCommand("wstag", "list")
	if err == nil {
		t.Fatal("expected error for wstag list without required flags, got nil")
	}
	// Getting a required flag error (not "unknown command") confirms the alias works
	if strings.Contains(err.Error(), "unknown command") {
		t.Error("alias 'wstag' did not route to workspace-tag command")
	}
}

func TestCmdWorkspaceTagSubcommands(t *testing.T) {
	resetGlobalFlags()

	var wstagCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "workspace-tag" {
			wstagCmd = sub
			break
		}
	}
	if wstagCmd == nil {
		t.Fatal("expected workspace-tag command to be registered on root")
	}

	expected := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
	}

	for _, sub := range wstagCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("workspace-tag missing expected subcommand %q", name)
		}
	}
}

