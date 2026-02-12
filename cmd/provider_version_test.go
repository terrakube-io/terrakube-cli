package cmd

import (
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"
)

func TestCmdProviderVersionListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/provider/prov-123/version") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		proto := "5.0"
		versions := []*terrakube.ProviderVersion{
			{ID: "pv-1", VersionNumber: "5.0.0", Protocols: &proto},
			{ID: "pv-2", VersionNumber: "4.67.0"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, versions)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"provider-version", "list",
		"--organization-id", "org-abc",
		"--provider-id", "prov-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "pv-1") {
		t.Errorf("expected output to contain 'pv-1', got: %s", out)
	}
	if !strings.Contains(out, "5.0.0") {
		t.Errorf("expected output to contain '5.0.0', got: %s", out)
	}
}

func TestCmdProviderVersionDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/provider/prov-123/version/pv-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"provider-version", "delete",
		"--organization-id", "org-abc",
		"--provider-id", "prov-123",
		"--id", "pv-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdProviderVersionListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("provider-version", "list", "--provider-id", "prov-123")
	if err == nil {
		t.Fatal("expected error for provider-version list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
