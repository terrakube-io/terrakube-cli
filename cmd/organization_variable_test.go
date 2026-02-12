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

func TestCmdOrganizationVariableListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/globalvar") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganizationVariableList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization-variable", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "TF_LOG") {
		t.Errorf("expected output to contain variable key 'TF_LOG', got: %s", out)
	}
	if !strings.Contains(out, "AWS_DEFAULT_REGION") {
		t.Errorf("expected output to contain variable key 'AWS_DEFAULT_REGION', got: %s", out)
	}
}

func TestCmdOrganizationVariableGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/globalvar/ov-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganizationVariable())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization-variable", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "ov-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "TF_LOG") {
		t.Errorf("expected output to contain variable key 'TF_LOG', got: %s", out)
	}
}

func TestCmdOrganizationVariableCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureOrganizationVariable())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"organization-variable", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--key", "TF_LOG",
		"--value", "DEBUG",
		"--description", "Terraform log level",
		"--category", "ENV",
		"--sensitive",
		"--hcl",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "TF_LOG") {
		t.Errorf("expected output to contain 'TF_LOG', got: %s", out)
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
	if attrs["key"] != "TF_LOG" {
		t.Errorf("expected key 'TF_LOG', got %v", attrs["key"])
	}
	if attrs["value"] != "DEBUG" {
		t.Errorf("expected value 'DEBUG', got %v", attrs["value"])
	}
	if attrs["description"] != "Terraform log level" {
		t.Errorf("expected description 'Terraform log level', got %v", attrs["description"])
	}
	if attrs["category"] != "ENV" {
		t.Errorf("expected category 'ENV', got %v", attrs["category"])
	}
	if attrs["sensitive"] != true {
		t.Errorf("expected sensitive true, got %v", attrs["sensitive"])
	}
	if attrs["hcl"] != true {
		t.Errorf("expected hcl true, got %v", attrs["hcl"])
	}
}

func TestCmdOrganizationVariableDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "globalvar/ov-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization-variable", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "ov-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "organization-variable deleted") {
		t.Errorf("expected 'organization-variable deleted' in output, got: %s", out)
	}
}

func TestCmdOrganizationVariableListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("organization-variable", "list")
	if err == nil {
		t.Fatal("expected error for organization-variable list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
