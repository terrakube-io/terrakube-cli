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

func TestCmdGithubAppTokenListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/github_app_token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureGithubAppTokenList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("github-app-token", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acme-corp") {
		t.Errorf("expected output to contain owner 'acme-corp', got: %s", out)
	}
	if !strings.Contains(out, "globex-corp") {
		t.Errorf("expected output to contain owner 'globex-corp', got: %s", out)
	}
}

func TestCmdGithubAppTokenGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "github_app_token/ga-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureGithubAppToken())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("github-app-token", "get", "--id", "ga-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "acme-corp") {
		t.Errorf("expected output to contain owner 'acme-corp', got: %s", out)
	}
}

func TestCmdGithubAppTokenCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureGithubAppToken())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"github-app-token", "create",
		"--app-id", "12345",
		"--installation-id", "67890",
		"--owner", "acme-corp",
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
	if attrs["appId"] != "12345" {
		t.Errorf("expected appId '12345', got %v", attrs["appId"])
	}
	if attrs["installationId"] != "67890" {
		t.Errorf("expected installationId '67890', got %v", attrs["installationId"])
	}
	if attrs["owner"] != "acme-corp" {
		t.Errorf("expected owner 'acme-corp', got %v", attrs["owner"])
	}
}

func TestCmdGithubAppTokenDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "github_app_token/ga-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("github-app-token", "delete", "--id", "ga-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "github-app-token deleted") {
		t.Errorf("expected 'github-app-token deleted' in output, got: %s", out)
	}
}
