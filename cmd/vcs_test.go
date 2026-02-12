package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdVCSListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/vcs") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVCSList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("vcs", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "github-main") {
		t.Errorf("expected output to contain VCS name 'github-main', got: %s", out)
	}
	if !strings.Contains(out, "gitlab-secondary") {
		t.Errorf("expected output to contain VCS name 'gitlab-secondary', got: %s", out)
	}
}

func TestCmdVCSGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/vcs/vcs-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVCS())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("vcs", "get", "--organization-id", "org-123", "--id", "vcs-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "github-main") {
		t.Errorf("expected output to contain VCS name 'github-main', got: %s", out)
	}
}

func TestCmdVCSCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVCS())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"vcs", "create",
		"--organization-id", "org-123",
		"--name", "github-main",
		"--description", "Main GitHub connection",
		"--vcs-type", "GITHUB",
		"--connection-type", "OAUTH",
		"--client-id", "client-id-123",
		"--client-secret", "client-secret-456",
		"--endpoint", "https://github.com",
		"--vcs-api-url", "https://api.github.com",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "github-main") {
		t.Errorf("expected output to contain 'github-main', got: %s", out)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body to contain data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "github-main" {
		t.Errorf("expected name 'github-main', got %v", attrs["name"])
	}
	if attrs["description"] != "Main GitHub connection" {
		t.Errorf("expected description 'Main GitHub connection', got %v", attrs["description"])
	}
	if attrs["vcsType"] != "GITHUB" {
		t.Errorf("expected vcsType 'GITHUB', got %v", attrs["vcsType"])
	}
	if attrs["connectionType"] != "OAUTH" {
		t.Errorf("expected connectionType 'OAUTH', got %v", attrs["connectionType"])
	}
	if attrs["clientId"] != "client-id-123" {
		t.Errorf("expected clientId 'client-id-123', got %v", attrs["clientId"])
	}
	if attrs["clientSecret"] != "client-secret-456" {
		t.Errorf("expected clientSecret 'client-secret-456', got %v", attrs["clientSecret"])
	}
	if attrs["endpoint"] != "https://github.com" {
		t.Errorf("expected endpoint 'https://github.com', got %v", attrs["endpoint"])
	}
	if attrs["apiUrl"] != "https://api.github.com" {
		t.Errorf("expected apiUrl 'https://api.github.com', got %v", attrs["apiUrl"])
	}
}

func TestCmdVCSDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "vcs/vcs-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("vcs", "delete", "--organization-id", "org-123", "--id", "vcs-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "vcs deleted") {
		t.Errorf("expected 'vcs deleted' in output, got: %s", out)
	}
}

func TestCmdVCSListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	// Need a valid server so newClient() doesn't os.Exit(1) on empty URL.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("vcs", "list")
	if err == nil {
		t.Fatal("expected error for vcs list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization flags, got: %v", err)
	}
}

func TestCmdVCSNameResolution(t *testing.T) {
	resetGlobalFlags()

	var resolverCalled bool
	var vcsListCalled bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")

		if strings.Contains(r.URL.Path, "/organization") && !strings.Contains(r.URL.Path, "/vcs") {
			resolverCalled = true
			if !strings.Contains(r.URL.RawQuery, "name%3D%3Dacme-corp") && !strings.Contains(r.URL.RawQuery, "name==acme-corp") {
				t.Errorf("expected org list filter with name==acme-corp, got query: %s", r.URL.RawQuery)
			}
			org := &terrakube.Organization{ID: "resolved-org-id", Name: "acme-corp"}
			_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{org})
			return
		}

		if strings.Contains(r.URL.Path, "/vcs") {
			vcsListCalled = true
			if !strings.Contains(r.URL.Path, "organization/resolved-org-id/vcs") {
				t.Errorf("expected resolved org ID in path, got: %s", r.URL.Path)
			}
			_ = jsonapi.MarshalPayload(w, testutil.FixtureVCSList())
			return
		}

		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("vcs", "list", "--organization-name", "acme-corp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resolverCalled {
		t.Error("expected organization name resolver to be called")
	}
	if !vcsListCalled {
		t.Error("expected VCS list endpoint to be called")
	}
	if !strings.Contains(out, "github-main") {
		t.Errorf("expected output to contain 'github-main', got: %s", out)
	}
}
