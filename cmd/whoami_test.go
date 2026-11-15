package cmd

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

func whoamiServer(t *testing.T) *[]string {
	t.Helper()
	var paths []string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/current-teams"):
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"groups":["TERRAKUBE_ADMIN","DEVS"]}`))
		case strings.HasSuffix(r.URL.Path, "/api/v1/organization"):
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{{ID: runOrg, Name: "my-org"}})
		case strings.HasSuffix(r.URL.Path, "/organization/"+runOrg+"/workspace"):
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, []*terrakube.Workspace{{ID: runWs, Name: "prod"}})
		case strings.HasSuffix(r.URL.Path, "/permissions/organization/"+runOrg):
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"manageWorkspace":true,"planJob":true}`))
		case strings.HasSuffix(r.URL.Path, "/permissions/organization/"+runOrg+"/workspace/"+runWs):
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"planJob":true,"approveJob":true}`))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	t.Cleanup(ts.Close)
	return &paths
}

// AC-44: whoami prints the teams, and the permissions with --org / --workspace.
func TestCmdWhoami(t *testing.T) {
	resetGlobalFlags()
	whoamiServer(t)
	out, err := executeCommand("whoami")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"TERRAKUBE_ADMIN"`) || strings.Contains(out, "permissions") {
		t.Errorf("teams only expected, got %q", out)
	}

	resetGlobalFlags()
	whoamiServer(t)
	out, err = executeCommand("whoami", "--org", "my-org")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"manageWorkspace": true`) || !strings.Contains(out, `"organization": "`+runOrg+`"`) {
		t.Errorf("org permissions expected, got %q", out)
	}

	resetGlobalFlags()
	whoamiServer(t)
	out, err = executeCommand("whoami", "--org", "my-org", "--workspace", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"approveJob": true`) || !strings.Contains(out, `"workspace": "`+runWs+`"`) {
		t.Errorf("workspace permissions expected, got %q", out)
	}
}

func TestCmdWhoamiWorkspaceNeedsOrg(t *testing.T) {
	resetGlobalFlags()
	paths := whoamiServer(t)
	_, err := executeCommand("whoami", "--workspace", "prod")
	if err == nil || err.Error() != "--workspace needs --org" {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*paths) != 0 {
		t.Errorf("expected no request, got %v", *paths)
	}
}

// Errors after the teams were read: an unknown org name, a failing
// permissions call, an unknown workspace name.
func TestCmdWhoamiLaterErrors(t *testing.T) {
	for _, args := range [][]string{
		{"whoami", "--org", "ghost"},
		{"whoami", "--org", runOrg},
		{"whoami", "--org", runOrg, "--workspace", "ghost"},
		{"whoami", "--org", runOrg, "--workspace", runWs},
	} {
		resetGlobalFlags()
		ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/current-teams"):
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"groups":["DEVS"]}`))
			case strings.HasSuffix(r.URL.Path, "/api/v1/organization"), strings.HasSuffix(r.URL.Path, "/workspace"):
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_, _ = w.Write([]byte(`{"data":[]}`))
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		if _, err := executeCommand(args...); err == nil {
			t.Errorf("%v: expected an error", args)
		}
		ts.Close()
	}
}
