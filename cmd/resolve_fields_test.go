package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/testutil"
)

const resolveOrgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// nameLookupServer answers list requests for kind with items and records every
// write request body, so tests can check what was (or was not) written.
func nameLookupServer(t *testing.T, kind string, items interface{}, writeResp interface{}) *[][]byte {
	t.Helper()
	var writes [][]byte
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Method == http.MethodGet {
			if !strings.HasSuffix(r.URL.Path, "/"+kind) {
				t.Errorf("unexpected lookup %s", r.URL.Path)
			}
			_ = jsonapi.MarshalPayload(w, items)
			return
		}
		b, _ := io.ReadAll(r.Body)
		writes = append(writes, b)
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_ = jsonapi.MarshalPayload(w, writeResp)
	}))
	t.Cleanup(ts.Close)
	return &writes
}

func writeData(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var payload struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("parse %q: %v", body, err)
	}
	return payload.Data
}

// AC-6: a template name on `job create -t` is resolved to its ID.
func TestResolveJobTemplateByName(t *testing.T) {
	resetGlobalFlags()
	writes := nameLookupServer(t, "template", []*terrakube.Template{{ID: "tpl-plan", Name: "Plan"}, {ID: "tpl-x", Name: "Plan and apply"}}, testutil.FixtureJob())

	if _, err := executeCommand("job", "create", "--organization-id", resolveOrgID,
		"--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123", "-t", "Plan"); err != nil {
		t.Fatal(err)
	}
	if len(*writes) != 1 {
		t.Fatalf("expected one write, got %d", len(*writes))
	}
	attrs, _ := writeData(t, (*writes)[0])["attributes"].(map[string]interface{})
	if attrs["templateReference"] != "tpl-plan" {
		t.Errorf("expected templateReference tpl-plan, got %v", attrs)
	}
}

// AC-6, AC-4a: workspace --vcs and --template take names.
func TestResolveWorkspaceLinksByName(t *testing.T) {
	resetGlobalFlags()
	var writes [][]byte
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/vcs"):
			_ = jsonapi.MarshalPayload(w, []*terrakube.VCS{{ID: "vcs-gh", Name: "github"}})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/template"):
			_ = jsonapi.MarshalPayload(w, []*terrakube.Template{{ID: "tpl-apply", Name: "Plan and apply"}})
		default:
			b, _ := io.ReadAll(r.Body)
			writes = append(writes, b)
			w.WriteHeader(http.StatusCreated)
			_ = jsonapi.MarshalPayload(w, testutil.FixtureWorkspace())
		}
	}))
	defer ts.Close()

	if _, err := executeCommand("workspace", "create", "--organization-id", resolveOrgID,
		"--name", "n", "--vcs", "github", "--template", "Plan and apply"); err != nil {
		t.Fatal(err)
	}
	if len(writes) != 1 {
		t.Fatalf("expected one write, got %d", len(writes))
	}
	data := writeData(t, writes[0])
	rels, _ := data["relationships"].(map[string]interface{})
	vcs, _ := rels["vcs"].(map[string]interface{})
	if d, _ := vcs["data"].(map[string]interface{}); d["id"] != "vcs-gh" {
		t.Errorf("expected vcs relationship vcs-gh, got %v", rels)
	}
	attrs, _ := data["attributes"].(map[string]interface{})
	if attrs["defaultTemplate"] != "tpl-apply" {
		t.Errorf("expected defaultTemplate tpl-apply, got %v", attrs)
	}
}

// AC-7: an unknown name fails with the organization in the message and writes nothing.
func TestResolveNotFoundWritesNothing(t *testing.T) {
	resetGlobalFlags()
	writes := nameLookupServer(t, "template", []*terrakube.Template{}, testutil.FixtureJob())

	_, err := executeCommand("job", "create", "--organization-id", resolveOrgID,
		"--workspace-id", "d4e5f6a7-b8c9-0123-defa-234567890123", "-t", "ghost")
	want := `template "ghost" not found in organization ` + resolveOrgID
	if err == nil || err.Error() != want {
		t.Fatalf("expected %q, got %v", want, err)
	}
	if len(*writes) != 0 {
		t.Errorf("expected no write, got %d", len(*writes))
	}
}

// AC-8: an ambiguous name lists the matching IDs and writes nothing.
func TestResolveAmbiguousWritesNothing(t *testing.T) {
	resetGlobalFlags()
	writes := nameLookupServer(t, "agent", []*terrakube.Agent{{ID: "ag-1", Name: "pool"}, {ID: "ag-2", Name: "pool"}}, testutil.FixtureWorkspace())

	_, err := executeCommand("workspace", "update", "--organization-id", resolveOrgID,
		"--id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--agent", "pool")
	want := `multiple agents match "pool": ag-1, ag-2; pass the ID instead`
	if err == nil || err.Error() != want {
		t.Fatalf("expected %q, got %v", want, err)
	}
	if len(*writes) != 0 {
		t.Errorf("expected no write, got %d", len(*writes))
	}
}
