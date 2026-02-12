package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdImplementationListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/provider/prov-123/version/ver-456/implementation") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		impls := []*terrakube.Implementation{
			{ID: "im-1", Os: "linux", Arch: "amd64", Filename: "provider_linux_amd64.zip"},
			{ID: "im-2", Os: "darwin", Arch: "arm64", Filename: "provider_darwin_arm64.zip"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, impls)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"implementation", "list",
		"--organization-id", "org-abc",
		"--provider-id", "prov-123",
		"--provider-version-id", "ver-456",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "im-1") {
		t.Errorf("expected output to contain 'im-1', got: %s", out)
	}
	if !strings.Contains(out, "linux") {
		t.Errorf("expected output to contain 'linux', got: %s", out)
	}
}

func TestCmdImplementationDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/provider/prov-123/version/ver-456/implementation/im-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"implementation", "delete",
		"--organization-id", "org-abc",
		"--provider-id", "prov-123",
		"--provider-version-id", "ver-456",
		"--id", "im-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdImplementationListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("implementation", "list", "--provider-id", "prov-123", "--provider-version-id", "ver-456")
	if err == nil {
		t.Fatal("expected error for implementation list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
