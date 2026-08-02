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

func TestCmdVariableListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/38b6635a-d38e-46f2-a95e-d00a416de4fd/variable") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVariableList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("variable", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "AWS_REGION") {
		t.Errorf("expected output to contain variable key 'AWS_REGION', got: %s", out)
	}
	if !strings.Contains(out, "DB_PASSWORD") {
		t.Errorf("expected output to contain variable key 'DB_PASSWORD', got: %s", out)
	}
}

func TestCmdVariableGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/38b6635a-d38e-46f2-a95e-d00a416de4fd/variable/b8c9d0e1-f2a3-4567-bcde-678901234567") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVariable())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("variable", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd", "--id", "b8c9d0e1-f2a3-4567-bcde-678901234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "AWS_REGION") {
		t.Errorf("expected output to contain variable key 'AWS_REGION', got: %s", out)
	}
}

func TestCmdVariableCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/38b6635a-d38e-46f2-a95e-d00a416de4fd/variable") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVariable())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"variable", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd",
		"--key", "AWS_REGION",
		"--value", "us-east-1",
		"--description", "AWS region for deployments",
		"--category", "ENV",
		"--sensitive=false",
		"--hcl=false",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "AWS_REGION") {
		t.Errorf("expected output to contain 'AWS_REGION', got: %s", out)
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
	if attrs["key"] != "AWS_REGION" {
		t.Errorf("expected key 'AWS_REGION', got %v", attrs["key"])
	}
	if attrs["value"] != "us-east-1" {
		t.Errorf("expected value 'us-east-1', got %v", attrs["value"])
	}
	if attrs["description"] != "AWS region for deployments" {
		t.Errorf("expected description 'AWS region for deployments', got %v", attrs["description"])
	}
	if attrs["category"] != "ENV" {
		t.Errorf("expected category 'ENV', got %v", attrs["category"])
	}
}

func TestCmdVariableUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/38b6635a-d38e-46f2-a95e-d00a416de4fd/variable/b8c9d0e1-f2a3-4567-bcde-678901234567") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureVariable())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"variable", "update",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd",
		"--id", "b8c9d0e1-f2a3-4567-bcde-678901234567",
		"--value", "us-west-2",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "AWS_REGION") {
		t.Errorf("expected output to contain 'AWS_REGION', got: %s", out)
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
	if attrs["value"] != "us-west-2" {
		t.Errorf("expected value 'us-west-2', got %v", attrs["value"])
	}
}

func TestCmdVariableDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/38b6635a-d38e-46f2-a95e-d00a416de4fd/variable/b8c9d0e1-f2a3-4567-bcde-678901234567") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("variable", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd", "--id", "b8c9d0e1-f2a3-4567-bcde-678901234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "variable deleted") {
		t.Errorf("expected 'variable deleted' in output, got: %s", out)
	}
}

func TestCmdVariableListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("variable", "list", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd")
	if err == nil {
		t.Fatal("expected error for variable list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdVariableListMissingWorkspace(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("variable", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil {
		t.Fatal("expected error for variable list without workspace flags, got nil")
	}
	if !strings.Contains(err.Error(), "workspace") {
		t.Errorf("expected error to mention workspace, got: %v", err)
	}
}
