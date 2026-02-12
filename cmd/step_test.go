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
		if !strings.Contains(r.URL.Path, "organization/org-abc/job/job-123/step") {
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
		"--organization-id", "org-abc",
		"--job-id", "job-123",
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
		if !strings.Contains(r.URL.Path, "organization/org-abc/job/job-123/step/st-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"step", "delete",
		"--organization-id", "org-abc",
		"--job-id", "job-123",
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

	_, err := executeCommand("step", "list", "--job-id", "job-123")
	if err == nil {
		t.Fatal("expected error for step list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") && !strings.Contains(err.Error(), "organization-name") {
		t.Errorf("expected error to mention organization-id or organization-name, got: %v", err)
	}
}
