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

func TestCmdJobListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureJobList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("job", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "plan") {
		t.Errorf("expected output to contain 'plan', got: %s", out)
	}
	if !strings.Contains(out, "apply") {
		t.Errorf("expected output to contain 'apply', got: %s", out)
	}
}

func TestCmdJobGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/d0e1f2a3-b4c5-6789-defa-890123456789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureJob())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("job", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--id", "d0e1f2a3-b4c5-6789-defa-890123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "plan") {
		t.Errorf("expected output to contain 'plan', got: %s", out)
	}
}

func TestCmdJobCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureJob())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"job", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
		"--command", "plan",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "plan") {
		t.Errorf("expected output to contain 'plan', got: %s", out)
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
	if attrs["command"] != "plan" {
		t.Errorf("expected command 'plan', got %v", attrs["command"])
	}
	relationships, ok := data["relationships"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain relationships")
	}
	wsRel, ok := relationships["workspace"].(map[string]interface{})
	if !ok {
		t.Fatal("expected workspace relationship")
	}
	wsData, ok := wsRel["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected workspace relationship data")
	}
	if wsData["id"] != "d4e5f6a7-b8c9-0123-defa-234567890123" {
		t.Errorf("expected workspace ID 'd4e5f6a7-b8c9-0123-defa-234567890123', got %v", wsData["id"])
	}
}

func TestCmdJobDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "job/job-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("job", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--id", "job-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "job deleted") {
		t.Errorf("expected 'job deleted' in output, got: %s", out)
	}
}

func TestCmdJobListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("job", "list")
	if err == nil {
		t.Fatal("expected error for job list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
