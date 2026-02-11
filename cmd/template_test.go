package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdTemplateListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/template") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTemplateList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("template", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "standard-plan") {
		t.Errorf("expected output to contain template name 'standard-plan', got: %s", out)
	}
	if !strings.Contains(out, "custom-apply") {
		t.Errorf("expected output to contain template name 'custom-apply', got: %s", out)
	}
}

func TestCmdTemplateGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/template/tpl-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTemplate())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("template", "get", "--organization-id", "org-123", "--id", "tpl-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "standard-plan") {
		t.Errorf("expected output to contain template name 'standard-plan', got: %s", out)
	}
}

func TestCmdTemplateCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTemplate())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"template", "create",
		"--organization-id", "org-123",
		"--name", "standard-plan",
		"--content", "flow:\n  - type: terraformPlan",
		"--description", "Standard Terraform plan template",
		"--version", "1.0.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "standard-plan") {
		t.Errorf("expected output to contain 'standard-plan', got: %s", out)
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
	if attrs["name"] != "standard-plan" {
		t.Errorf("expected name 'standard-plan', got %v", attrs["name"])
	}
	if attrs["tcl"] != "flow:\n  - type: terraformPlan" {
		t.Errorf("expected tcl content, got %v", attrs["tcl"])
	}
	if attrs["description"] != "Standard Terraform plan template" {
		t.Errorf("expected description, got %v", attrs["description"])
	}
	if attrs["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", attrs["version"])
	}
}

func TestCmdTemplateDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "template/tpl-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("template", "delete", "--organization-id", "org-123", "--id", "tpl-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "template deleted") {
		t.Errorf("expected 'template deleted' in output, got: %s", out)
	}
}

func TestCmdTemplateListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := executeCommand("template", "list")
	if err == nil {
		t.Fatal("expected error for template list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization flags, got: %v", err)
	}
}

func TestCmdTemplateNameResolution(t *testing.T) {
	resetGlobalFlags()

	var resolverCalled bool
	var templateListCalled bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")

		if strings.Contains(r.URL.Path, "/organization") && !strings.Contains(r.URL.Path, "/template") {
			resolverCalled = true
			if !strings.Contains(r.URL.RawQuery, "name%3D%3Dacme-corp") && !strings.Contains(r.URL.RawQuery, "name==acme-corp") {
				t.Errorf("expected org list filter with name==acme-corp, got query: %s", r.URL.RawQuery)
			}
			org := &terrakube.Organization{ID: "resolved-org-id", Name: "acme-corp"}
			_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{org})
			return
		}

		if strings.Contains(r.URL.Path, "/template") {
			templateListCalled = true
			if !strings.Contains(r.URL.Path, "organization/resolved-org-id/template") {
				t.Errorf("expected resolved org ID in path, got: %s", r.URL.Path)
			}
			_ = jsonapi.MarshalPayload(w, testutil.FixtureTemplateList())
			return
		}

		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("template", "list", "--organization-name", "acme-corp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resolverCalled {
		t.Error("expected organization name resolver to be called")
	}
	if !templateListCalled {
		t.Error("expected template list endpoint to be called")
	}
	if !strings.Contains(out, "standard-plan") {
		t.Errorf("expected output to contain 'standard-plan', got: %s", out)
	}
}

func TestCmdTemplateAliasTplRoutes(t *testing.T) {
	resetGlobalFlags()

	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := executeCommand("tpl", "list")
	if err == nil {
		t.Fatal("expected error for tpl list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected alias 'tpl' to route to template command, got error: %v", err)
	}
}
