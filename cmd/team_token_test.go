package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCmdTeamTokenList(t *testing.T) {
	resetGlobalFlags()

	var method, path string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"tt-1","description":"first","group":"dev"},{"id":"tt-2","description":"second","group":"ops"}]`))
	}))
	defer ts.Close()

	out, err := executeCommand("team-token", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != http.MethodGet {
		t.Errorf("expected GET, got %s", method)
	}
	if !strings.HasSuffix(path, "/access-token/v1/teams") {
		t.Errorf("unexpected path: %s", path)
	}
	for _, want := range []string{"tt-1", "tt-2", "first", "second"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got: %s", want, out)
		}
	}
}

func TestCmdTeamTokenDelete(t *testing.T) {
	resetGlobalFlags()

	var method, path string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	out, err := executeCommand("team-token", "delete", "--id", "tt-9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", method)
	}
	if !strings.HasSuffix(path, "/access-token/v1/teams/tt-9") {
		t.Errorf("unexpected path: %s", path)
	}
	if !strings.Contains(out, "Team token deleted") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestCmdTeamTokenCreateBody(t *testing.T) {
	resetGlobalFlags()

	var body []byte
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"tt-1","token":"secret-value"}`))
	}))
	defer ts.Close()

	out, err := executeCommand("team-token", "create", "-g", "dev", "-d", "ci", "--days", "1", "--hours", "2", "--minutes", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	want := map[string]any{"group": "dev", "description": "ci", "days": 1.0, "hours": 2.0, "minutes": 3.0}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("body[%q] = %v, want %v", k, got[k], v)
		}
	}
	if !strings.Contains(out, "secret-value") {
		t.Errorf("expected output to contain token value, got: %s", out)
	}
}

func TestCmdTeamTokenAPIErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"create", []string{"team-token", "create", "--group", "dev"}},
		{"list", []string{"team-token", "list"}},
		{"delete", []string{"team-token", "delete", "--id", "tt-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			var requests atomic.Int32
			ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				w.WriteHeader(http.StatusForbidden)
			}))
			defer ts.Close()

			out, err := executeCommand(tt.args...)
			if err == nil || !strings.Contains(err.Error(), "403") {
				t.Fatalf("expected 403 error, got %v", err)
			}
			if requests.Load() != 1 {
				t.Errorf("expected 1 request, got %d", requests.Load())
			}
			if strings.Contains(out, "Team token deleted") {
				t.Errorf("must not report success on error, got: %q", out)
			}
		})
	}
}

func TestCmdTeamTokenNotLoggedIn(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"create", []string{"team-token", "create", "--group", "dev"}},
		{"list", []string{"team-token", "list"}},
		{"delete", []string{"team-token", "delete", "--id", "tt-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			out, err := executeCommand(tt.args...)
			if err == nil || !strings.Contains(err.Error(), "not logged in") {
				t.Fatalf("expected not-logged-in error, got %v", err)
			}
			if strings.Contains(out, "Team token deleted") {
				t.Errorf("must not report success, got: %q", out)
			}
		})
	}
}

func TestCmdTeamTokenDeleteRequiresID(t *testing.T) {
	resetGlobalFlags()

	_, err := executeCommand("team-token", "delete")
	if err == nil || !strings.Contains(err.Error(), `"id"`) {
		t.Fatalf("expected required flag error for id, got: %v", err)
	}
}
