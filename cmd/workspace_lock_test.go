package cmd

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// lockServer resolves workspace "prod" by name and records PATCH bodies.
func lockServer(t *testing.T) func() []string {
	t.Helper()
	var mu sync.Mutex
	var patches []string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/organization/"+runOrg+"/workspace"):
			_ = jsonapi.MarshalPayload(w, []*terrakube.Workspace{{ID: runWs, Name: "prod"}})
		case r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/workspace/"+runWs):
			b, _ := io.ReadAll(r.Body)
			mu.Lock()
			patches = append(patches, string(b))
			mu.Unlock()
			_ = jsonapi.MarshalPayload(w, &terrakube.Workspace{ID: runWs, Name: "prod"})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(ts.Close)
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), patches...)
	}
}

// AC-30: lock sets locked and the reason; unlock clears locked.
func TestCmdWorkspaceLockUnlock(t *testing.T) {
	resetGlobalFlags()
	patches := lockServer(t)
	if _, err := executeCommand("workspace", "lock", "prod", "-o", runOrg, "--reason", "release freeze"); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCommand("workspace", "unlock", runWs, "-o", runOrg); err != nil {
		t.Fatal(err)
	}
	got := patches()
	if len(got) != 2 {
		t.Fatalf("expected two PATCHes, got %v", got)
	}
	if !strings.Contains(got[0], `"locked":true`) || !strings.Contains(got[0], `"lockDescription":"release freeze"`) {
		t.Errorf("lock body: %s", got[0])
	}
	if !strings.Contains(got[1], `"attributes":{"locked":false}`) {
		t.Errorf("unlock body: %s", got[1])
	}
}
