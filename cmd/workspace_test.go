package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/testutil"
)

func TestCmdWorkspaceListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspaceList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("workspace", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production-vpc") {
		t.Errorf("expected output to contain 'production-vpc', got: %s", out)
	}
	if !strings.Contains(out, "staging-vpc") {
		t.Errorf("expected output to contain 'staging-vpc', got: %s", out)
	}
}

func TestCmdWorkspaceGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/d4e5f6a7-b8c9-0123-defa-234567890123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspace())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("workspace", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "d4e5f6a7-b8c9-0123-defa-234567890123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production-vpc") {
		t.Errorf("expected output to contain 'production-vpc', got: %s", out)
	}
}

func TestCmdWorkspaceCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspace())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--name", "production-vpc",
		"--description", "Production VPC infrastructure",
		"--source", "https://github.com/acme-corp/infra.git",
		"--branch", "main",
		"--folder", "/",
		"--iac-type", "terraform",
		"--iac-version", "1.5.7",
		"--execution-mode", "remote",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production-vpc") {
		t.Errorf("expected output to contain 'production-vpc', got: %s", out)
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
	if attrs["name"] != "production-vpc" {
		t.Errorf("expected name 'production-vpc', got %v", attrs["name"])
	}
	if attrs["branch"] != "main" {
		t.Errorf("expected branch 'main', got %v", attrs["branch"])
	}
	if attrs["executionMode"] != "remote" {
		t.Errorf("expected executionMode 'remote', got %v", attrs["executionMode"])
	}
}

func TestCmdWorkspaceListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("workspace", "list")
	if err == nil {
		t.Fatal("expected error for workspace list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdWorkspaceUpdateSendsOnlyChangedAttributes(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected only PATCH, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/d4e5f6a7-b8c9-0123-defa-234567890123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspace())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"workspace", "update",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--id", "d4e5f6a7-b8c9-0123-defa-234567890123",
		"--description", "d",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse PATCH body: %v", err)
	}
	data, _ := bodyMap["data"].(map[string]interface{})
	if data["type"] != "workspace" || data["id"] != "d4e5f6a7-b8c9-0123-defa-234567890123" {
		t.Errorf("expected workspace type and id, got: %s", capturedBody)
	}
	attrs, _ := data["attributes"].(map[string]interface{})
	if len(attrs) != 1 || attrs["description"] != "d" {
		t.Errorf("expected attributes {description: d}, got: %s", capturedBody)
	}
	if _, ok := data["relationships"]; ok {
		t.Errorf("PATCH body must not contain relationships, got: %s", capturedBody)
	}
}

// captureWorkspaceWrite runs a workspace create/update and returns the request's data object.
func captureWorkspaceWrite(t *testing.T, args ...string) map[string]interface{} {
	t.Helper()
	resetGlobalFlags()

	var body []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspace())
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	if _, err := executeCommand(append([]string{"workspace"}, args...)...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var payload struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse body %q: %v", body, err)
	}
	return payload.Data
}

const (
	testVcsID   = "e5f6a7b8-c9d0-1234-efab-345678901234"
	testAgentID = "f6a7b8c9-d0e1-2345-fabc-456789012345"
)

// AC-4: relationship flags go out as JSON:API relationships in the same request.
func TestCmdWorkspaceCreateSendsRelationships(t *testing.T) {
	data := captureWorkspaceWrite(t, "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--name", "n", "--vcs", testVcsID, "--agent", testAgentID)

	rels, _ := data["relationships"].(map[string]interface{})
	if len(rels) != 2 {
		t.Fatalf("expected relationships vcs and agent only, got %v", data["relationships"])
	}
	for name, want := range map[string]map[string]interface{}{
		"vcs":   {"type": "vcs", "id": testVcsID},
		"agent": {"type": "agent", "id": testAgentID},
	} {
		rel, _ := rels[name].(map[string]interface{})
		got, _ := rel["data"].(map[string]interface{})
		if got["type"] != want["type"] || got["id"] != want["id"] {
			t.Errorf("relationship %s: got %v, want %v", name, rel, want)
		}
	}
}

// AC-5: an empty value clears the link ({"data": null}).
func TestCmdWorkspaceUpdateClearsRelationship(t *testing.T) {
	data := captureWorkspaceWrite(t, "update", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--agent", "")

	rels, _ := data["relationships"].(map[string]interface{})
	rel, ok := rels["agent"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected an agent relationship, got %v", data)
	}
	if d, present := rel["data"]; !present || d != nil {
		t.Errorf("expected agent data null, got %v", rel)
	}
	if attrs, _ := data["attributes"].(map[string]interface{}); len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}
}

// AC-14: table output shows the workspace's VCS, agent and project IDs.
func TestCmdWorkspaceGetTableShowsRelationships(t *testing.T) {
	resetGlobalFlags()
	ws := testutil.FixtureWorkspace()
	ws.Vcs = &terrakube.VCS{ID: testVcsID}
	ws.Agent = &terrakube.Agent{ID: testAgentID}
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	}))
	defer ts.Close()

	out, err := executeCommand("workspace", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--output", "table")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"VCS", "AGENT", "PROJECT", testVcsID, testAgentID} {
		if !strings.Contains(strings.ToUpper(out), strings.ToUpper(want)) {
			t.Errorf("expected %q in table output:\n%s", want, out)
		}
	}
}
