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

func TestCmdTagListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/tag") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTagList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("tag", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production") {
		t.Errorf("expected output to contain tag name 'production', got: %s", out)
	}
	if !strings.Contains(out, "staging") {
		t.Errorf("expected output to contain tag name 'staging', got: %s", out)
	}
}

func TestCmdTagGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/tag/tag-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTag())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("tag", "get", "--organization-id", "org-123", "--id", "tag-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production") {
		t.Errorf("expected output to contain tag name 'production', got: %s", out)
	}
}

func TestCmdTagCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureTag())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"tag", "create",
		"--organization-id", "org-123",
		"--name", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "production") {
		t.Errorf("expected output to contain 'production', got: %s", out)
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
	if attrs["name"] != "production" {
		t.Errorf("expected name 'production', got %v", attrs["name"])
	}
}

func TestCmdTagDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "tag/tag-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("tag", "delete", "--organization-id", "org-123", "--id", "tag-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "tag deleted") {
		t.Errorf("expected 'tag deleted' in output, got: %s", out)
	}
}

func TestCmdTagListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("tag", "list")
	if err == nil {
		t.Fatal("expected error for tag list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization flags, got: %v", err)
	}
}
