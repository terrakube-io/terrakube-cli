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

func TestCmdNotificationConfigurationListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/notificationConfiguration") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		configs := testutil.FixtureNotificationConfigurationList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, configs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-configuration", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "c1a2b3c4") {
		t.Errorf("expected output to contain 'c1a2b3c4', got: %s", out)
	}
	if !strings.Contains(out, "slack-alerts") {
		t.Errorf("expected output to contain 'slack-alerts', got: %s", out)
	}
}

func TestCmdNotificationConfigurationListWorkspaceE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/workspace/d4e5f6a7-b8c9-0123-defa-234567890123/notificationConfiguration") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		configs := testutil.FixtureNotificationConfigurationList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, configs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-configuration", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "c1a2b3c4") {
		t.Errorf("expected output to contain 'c1a2b3c4', got: %s", out)
	}
}

func TestCmdNotificationConfigurationGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/notificationConfiguration/c1a2b3c4-d4e5-6789-abcd-ef1234567890") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		config := testutil.FixtureNotificationConfiguration()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, config)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-configuration", "get",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "c1a2b3c4") {
		t.Errorf("expected output to contain 'c1a2b3c4', got: %s", out)
	}
	if !strings.Contains(out, "slack-alerts") {
		t.Errorf("expected output to contain 'slack-alerts', got: %s", out)
	}
}

func TestCmdNotificationConfigurationCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var reqBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		reqBody, _ = io.ReadAll(r.Body)

		config := testutil.FixtureNotificationConfiguration()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_ = jsonapi.MarshalPayload(w, config)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"notification-configuration", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--name", "slack-alerts",
		"--channel-type", "SLACK",
		"--destination-url", "https://hooks.slack.com/services/T00/B00/X00",
		"--message-style", "DETAILED",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "slack-alerts") {
		t.Errorf("expected output to contain 'slack-alerts', got: %s", out)
	}
	if !bytes.Contains(reqBody, []byte("slack-alerts")) {
		t.Errorf("expected request body to contain 'slack-alerts', got: %s", string(reqBody))
	}
}

func TestCmdNotificationConfigurationDeleteE2E(t *testing.T) {
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
		"notification-configuration", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--id", "c1a2b3c4-d4e5-6789-abcd-ef1234567890",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "notification-configuration deleted") {
		t.Errorf("expected 'notification-configuration deleted' in output, got: %s", out)
	}
}
