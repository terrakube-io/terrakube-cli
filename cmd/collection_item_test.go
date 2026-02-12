package cmd

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"

	"terrakube/testutil"
)

func TestCmdCollectionItemListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/collection/col-123/item") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionItemList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-item", "list",
		"--organization-id", "org-abc",
		"--collection-id", "col-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ci1a2b3c") {
		t.Errorf("expected output to contain 'ci1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "DB_HOST") {
		t.Errorf("expected output to contain 'DB_HOST', got: %s", out)
	}
}

func TestCmdCollectionItemDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-abc/collection/col-123/item/item-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"collection-item", "delete",
		"--organization-id", "org-abc",
		"--collection-id", "col-123",
		"--id", "item-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdCollectionItemListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("collection-item", "list", "--collection-id", "col-123")
	if err == nil {
		t.Fatal("expected error for collection-item list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
