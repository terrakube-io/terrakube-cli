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

func TestCmdAgentListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/agent") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureAgentList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("agent", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "k8s-runner") {
		t.Errorf("expected output to contain Agent name 'k8s-runner', got: %s", out)
	}
	if !strings.Contains(out, "docker-runner") {
		t.Errorf("expected output to contain Agent name 'docker-runner', got: %s", out)
	}
}

func TestCmdAgentGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/agent/agent-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureAgent())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("agent", "get", "--organization-id", "org-123", "--id", "agent-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "k8s-runner") {
		t.Errorf("expected output to contain Agent name 'k8s-runner', got: %s", out)
	}
}

func TestCmdAgentCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureAgent())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"agent", "create",
		"--organization-id", "org-123",
		"--name", "k8s-runner",
		"--description", "Kubernetes-based runner agent",
		"--url", "https://agent.example.com",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "k8s-runner") {
		t.Errorf("expected output to contain 'k8s-runner', got: %s", out)
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
	if attrs["name"] != "k8s-runner" {
		t.Errorf("expected name 'k8s-runner', got %v", attrs["name"])
	}
	if attrs["description"] != "Kubernetes-based runner agent" {
		t.Errorf("expected description 'Kubernetes-based runner agent', got %v", attrs["description"])
	}
	if attrs["url"] != "https://agent.example.com" {
		t.Errorf("expected url 'https://agent.example.com', got %v", attrs["url"])
	}
}

func TestCmdAgentDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "agent/agent-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("agent", "delete", "--organization-id", "org-123", "--id", "agent-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "agent deleted") {
		t.Errorf("expected 'agent deleted' in output, got: %s", out)
	}
}

func TestCmdAgentListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("agent", "list")
	if err == nil {
		t.Fatal("expected error for agent list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization flags, got: %v", err)
	}
}
