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
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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
		"--workspace-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901",
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

func TestCmdCollectionReferenceCreateScopes(t *testing.T) {
	const (
		org        = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		collection = "d4e5f6a7-b8c9-0123-defa-234567890123"
		workspace  = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
		project    = "c3d4e5f6-a7b8-9012-cdef-123456789012"
	)
	base := []string{"collection-reference", "create", "--organization-id", org, "--collection-id", collection}

	cases := map[string]struct {
		args        []string
		wantBody    []string
		wantNoBody  []string
		wantFailure string
	}{
		"project": {
			args:       []string{"--project-id", project},
			wantBody:   []string{`"scope":"PROJECT"`, project},
			wantNoBody: []string{`"workspace"`},
		},
		"organization": {
			args:       []string{"--scope", "ORGANIZATION"},
			wantBody:   []string{`"scope":"ORGANIZATION"`},
			wantNoBody: []string{`"workspace"`, `"project"`},
		},
		"workspace sends no scope": {
			args:       []string{"--workspace-id", workspace},
			wantBody:   []string{workspace},
			wantNoBody: []string{"scope", `"project"`},
		},
		"no target": {args: nil, wantFailure: "--workspace, --project, or --scope ORGANIZATION"},
		"both targets": {
			args:        []string{"--workspace-id", workspace, "--project-id", project},
			wantFailure: "only one of --workspace and --project",
		},
		"organization with a project": {
			args:        []string{"--scope", "ORGANIZATION", "--project-id", project},
			wantFailure: "takes neither",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resetGlobalFlags()
			var captured []byte
			ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_ = jsonapi.MarshalPayload(w, testutil.FixtureCollectionReference())
			}))
			defer ts.Close()

			_, err := executeCommand(append(append([]string{}, base...), tc.args...)...)

			if tc.wantFailure != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantFailure) {
					t.Fatalf("expected an error containing %q, got %v", tc.wantFailure, err)
				}
				if captured != nil {
					t.Errorf("the API must not be called for an invalid combination, got %s", captured)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range tc.wantBody {
				if !strings.Contains(string(captured), want) {
					t.Errorf("body %s is missing %s", captured, want)
				}
			}
			for _, unwanted := range tc.wantNoBody {
				if strings.Contains(string(captured), unwanted) {
					t.Errorf("body %s must not contain %s", captured, unwanted)
				}
			}
		})
	}
}
