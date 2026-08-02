package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdOrganizationListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/v1/organization") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganizationList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acme-corp") {
		t.Errorf("expected output to contain 'acme-corp', got: %s", out)
	}
	if !strings.Contains(out, "globex-corp") {
		t.Errorf("expected output to contain 'globex-corp', got: %s", out)
	}
}

func TestCmdOrganizationGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganization())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization", "get", "--id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acme-corp") {
		t.Errorf("expected output to contain 'acme-corp', got: %s", out)
	}
}

func TestCmdOrganizationCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganization())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"organization", "create",
		"--name", "acme-corp",
		"--description", "ACME Corporation infrastructure",
		"--execution-mode", "remote",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acme-corp") {
		t.Errorf("expected output to contain 'acme-corp', got: %s", out)
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
	if attrs["name"] != "acme-corp" {
		t.Errorf("expected name 'acme-corp', got %v", attrs["name"])
	}
	if attrs["description"] != "ACME Corporation infrastructure" {
		t.Errorf("expected description 'ACME Corporation infrastructure', got %v", attrs["description"])
	}
	if attrs["executionMode"] != "remote" {
		t.Errorf("expected executionMode 'remote', got %v", attrs["executionMode"])
	}
}

func TestCmdOrganizationDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization", "delete", "--id", "org-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "organization deleted") {
		t.Errorf("expected 'organization deleted' in output, got: %s", out)
	}
}
