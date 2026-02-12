package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdModuleVersionListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/module/mod-123/version") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		versions := []*terrakube.ModuleVersion{
			{ID: "mv-1", Version: "1.0.0"},
			{ID: "mv-2", Version: "0.9.0"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, versions)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module-version", "list",
		"--organization-id", "org-abc",
		"--module-id", "mod-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "mv-1") {
		t.Errorf("expected output to contain 'mv-1', got: %s", out)
	}
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected output to contain '1.0.0', got: %s", out)
	}
}

func TestCmdModuleVersionDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/module/mod-123/version/mv-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module-version", "delete",
		"--organization-id", "org-abc",
		"--module-id", "mod-123",
		"--id", "mv-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdModuleVersionListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("module-version", "list", "--module-id", "mod-123")
	if err == nil {
		t.Fatal("expected error for module-version list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
