package cmd

import (
	"net/http"
	"strings"
	"testing"

	"terrakube/testutil"

	"github.com/google/jsonapi"
)

func TestCmdAddressListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/c3d4e5f6-a7b8-9012-cdef-123456789012/address") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		addrs := testutil.FixtureAddressList()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, addrs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"address", "list",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ad1a2b3c") {
		t.Errorf("expected output to contain 'ad1a2b3c', got: %s", out)
	}
	if !strings.Contains(out, "aws_vpc.main") {
		t.Errorf("expected output to contain 'aws_vpc.main', got: %s", out)
	}
}

func TestCmdAddressDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/job/c3d4e5f6-a7b8-9012-cdef-123456789012/address/ad-789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"address", "delete",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012",
		"--id", "ad-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCmdAddressListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("address", "list", "--job-id", "c3d4e5f6-a7b8-9012-cdef-123456789012")
	if err == nil {
		t.Fatal("expected error for address list without org flag, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
