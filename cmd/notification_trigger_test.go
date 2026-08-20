package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"terrakube/testutil"

	"github.com/google/jsonapi"
)

func TestCmdNotificationTriggerListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/notificationConfiguration/c1a2b3c4-d4e5-6789-abcd-ef1234567890/triggers") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		triggers := testutil.FixtureNotificationTriggerList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, triggers)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-trigger", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--notification-configuration-id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "d1a2b3c4") {
		t.Errorf("expected output to contain 'd1a2b3c4', got: %s", out)
	}
	if !strings.Contains(out, "completed") {
		t.Errorf("expected output to contain 'completed', got: %s", out)
	}
}

func TestCmdNotificationTriggerGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/notificationConfiguration/c1a2b3c4-d4e5-6789-abcd-ef1234567890/triggers/d1a2b3c4-d4e5-6789-abcd-ef1234567890") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		trigger := testutil.FixtureNotificationTrigger()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, trigger)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-trigger", "get",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--notification-configuration-id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
		"--id", "d1a2b3c4-d4e5-6789-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "d1a2b3c4") {
		t.Errorf("expected output to contain 'd1a2b3c4', got: %s", out)
	}
	if !strings.Contains(out, "completed") {
		t.Errorf("expected output to contain 'completed', got: %s", out)
	}
}

func TestCmdNotificationTriggerCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var reqBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		reqBody, _ = io.ReadAll(r.Body)

		trigger := testutil.FixtureNotificationTrigger()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_ = jsonapi.MarshalPayload(w, trigger)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-trigger", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--notification-configuration-id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
		"--job-status", "completed",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "completed") {
		t.Errorf("expected output to contain 'completed', got: %s", out)
	}
	if !bytes.Contains(reqBody, []byte("completed")) {
		t.Errorf("expected request body to contain 'completed', got: %s", string(reqBody))
	}
}

func TestCmdNotificationTriggerDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-trigger", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--notification-configuration-id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
		"--id", "d1a2b3c4-d4e5-6789-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "notification-trigger deleted") {
		t.Errorf("expected 'notification-trigger deleted' in output, got: %s", out)
	}
}
