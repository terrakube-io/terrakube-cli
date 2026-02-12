package cmd

import (
	"net/http"
	"strings"
	"testing"

	"terrakube/testutil"

	"github.com/google/jsonapi"
)

func TestCmdStepListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/c3d4e5f6-a7b8-9012-cdef-123456789012/step") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		steps := testutil.FixtureStepList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, steps)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"step", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "st1a2b3c") {
		t.Errorf("expected output to contain 'st1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "Plan Step") {
		t.Errorf("expected output to contain 'Plan Step', got: %s", out)
	}
}

func TestCmdStepDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/c3d4e5f6-a7b8-9012-cdef-123456789012/step/st-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"step", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012",
		"--id", "st-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdStepListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("step", "list", "--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012")
	if err == nil {
		t.Fatal("expected error for step list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
