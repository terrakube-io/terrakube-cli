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

func TestCmdSSHListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/ssh") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureSSHList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("ssh", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deploy-key") {
		t.Errorf("expected output to contain SSH name 'deploy-key', got: %s", out)
	}
	if !strings.Contains(out, "backup-key") {
		t.Errorf("expected output to contain SSH name 'backup-key', got: %s", out)
	}
}

func TestCmdSSHGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/ssh/ssh-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureSSH())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("ssh", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "ssh-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deploy-key") {
		t.Errorf("expected output to contain SSH name 'deploy-key', got: %s", out)
	}
}

func TestCmdSSHCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureSSH())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"ssh", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--name", "deploy-key",
		"--private-key", "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
		"--ssh-type", "rsa",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deploy-key") {
		t.Errorf("expected output to contain 'deploy-key', got: %s", out)
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
	if attrs["name"] != "deploy-key" {
		t.Errorf("expected name 'deploy-key', got %v", attrs["name"])
	}
	if attrs["privateKey"] != "-----BEGIN RSA PRIVATE KEY-----\nMIIE..." {
		t.Errorf("expected privateKey, got %v", attrs["privateKey"])
	}
	if attrs["sshType"] != "rsa" {
		t.Errorf("expected sshType 'rsa', got %v", attrs["sshType"])
	}
}

func TestCmdSSHDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "ssh/ssh-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("ssh", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "ssh-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ssh deleted") {
		t.Errorf("expected 'ssh deleted' in output, got: %s", out)
	}
}

func TestCmdSSHListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("ssh", "list")
	if err == nil {
		t.Fatal("expected error for ssh list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
