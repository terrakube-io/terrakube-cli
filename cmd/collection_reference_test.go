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

func TestCmdCollectionReferenceListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/collection/d4e5f6a7-b8c9-0123-defa-234567890123/reference") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionReferenceList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-reference", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--collection-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "cr1a2b3c") {
		t.Errorf("expected output to contain reference ID 'cr1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "cr2b3c4d") {
		t.Errorf("expected output to contain reference ID 'cr2b3c4d', got: %s", out)
	}
}

func TestCmdCollectionReferenceGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "reference/ref-789") {
			t.Errorf("expected flat path with reference/ref-789, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionReference())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-reference", "get",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--collection-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
		"--id", "ref-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "cr1a2b3c") {
		t.Errorf("expected output to contain reference ID 'cr1a2b3c', got: %s", out)
	}
}

func TestCmdCollectionReferenceCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/collection/d4e5f6a7-b8c9-0123-defa-234567890123/reference") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionReference())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-reference", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--collection-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
		"--description", "Production workspace reference",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "cr1a2b3c") {
		t.Errorf("expected output to contain 'cr1a2b3c', got: %s", out)
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
	if attrs["description"] != "Production workspace reference" {
		t.Errorf("expected description 'Production workspace reference', got %v", attrs["description"])
	}
}

func TestCmdCollectionReferenceDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "reference/ref-del") {
			t.Errorf("expected flat path with reference/ref-del, got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-reference", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--collection-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
		"--id", "ref-del",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "collection-reference deleted") {
		t.Errorf("expected 'collection-reference deleted' in output, got: %s", out)
	}
}

func TestCmdCollectionReferenceListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("collection-reference", "list", "--collection-id", "d4e5f6a7-b8c9-0123-defa-234567890123")
	if err == nil {
		t.Fatal("expected error for collection-reference list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
