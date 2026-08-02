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

func TestCmdModuleListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/module") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureModuleList())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("module", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "vpc") {
		t.Errorf("expected output to contain 'vpc', got: %s", out)
	}
	if !strings.Contains(out, "s3-bucket") {
		t.Errorf("expected output to contain 's3-bucket', got: %s", out)
	}
}

func TestCmdModuleGetE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/module/f6a7b8c9-d0e1-2345-fabc-456789012345") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureModule())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("module", "get", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "f6a7b8c9-d0e1-2345-fabc-456789012345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "vpc") {
		t.Errorf("expected output to contain 'vpc', got: %s", out)
	}
}

func TestCmdModuleCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, testutil.FixtureModule())
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module", "create",
		"--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"--name", "vpc",
		"--description", "AWS VPC module",
		"--provider", "aws",
		"--source", "https://github.com/acme-corp/terraform-aws-vpc.git",
		"--tag-prefix", "v",
		"--folder", "/modules/vpc",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "vpc") {
		t.Errorf("expected output to contain 'vpc', got: %s", out)
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
	if attrs["name"] != "vpc" {
		t.Errorf("expected name 'vpc', got %v", attrs["name"])
	}
	if attrs["provider"] != "aws" {
		t.Errorf("expected provider 'aws', got %v", attrs["provider"])
	}
}

func TestCmdModuleDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "module/mod-del") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("module", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--id", "mod-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "module deleted") {
		t.Errorf("expected 'module deleted' in output, got: %s", out)
	}
}

func TestCmdModuleListMissingOrg(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("module", "list")
	if err == nil {
		t.Fatal("expected error for module list without org flags, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}
