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

func TestCmdCollectionListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/collection") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("collection", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "shared-vars") {
		t.Errorf("expected output to contain collection name 'shared-vars', got: %s", out)
	}
	if !strings.Contains(out, "env-config") {
		t.Errorf("expected output to contain collection name 'env-config', got: %s", out)
	}
}

func TestCmdCollectionGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/collection/col-456") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollection())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("collection", "get", "--organization-id", "org-123", "--id", "col-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "shared-vars") {
		t.Errorf("expected output to contain collection name 'shared-vars', got: %s", out)
	}
}

func TestCmdCollectionCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollection())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection", "create",
		"--organization-id", "org-123",
		"--name", "shared-vars",
		"--description", "Shared variable collection",
		"--priority", "10",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "shared-vars") {
		t.Errorf("expected output to contain 'shared-vars', got: %s", out)
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
	if attrs["name"] != "shared-vars" {
		t.Errorf("expected name 'shared-vars', got %v", attrs["name"])
	}
	if attrs["description"] != "Shared variable collection" {
		t.Errorf("expected description 'Shared variable collection', got %v", attrs["description"])
	}
	if attrs["priority"] != float64(10) {
		t.Errorf("expected priority 10, got %v", attrs["priority"])
	}
}

func TestCmdCollectionDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "collection/col-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("collection", "delete", "--organization-id", "org-123", "--id", "col-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "collection deleted") {
		t.Errorf("expected 'collection deleted' in output, got: %s", out)
	}
}

func TestCmdCollectionListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("collection", "list")
	if err == nil {
		t.Fatal("expected error for collection list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization flags, got: %v", err)
	}
}
