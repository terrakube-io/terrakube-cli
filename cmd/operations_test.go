package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

const opsJSON = `{"atomic:operations":[{"op":"add","data":{"type":"organization","attributes":{"name":"acme"}}}]}`

func TestCmdOperationsSubmit(t *testing.T) {
	resetGlobalFlags()
	file := writeTempFile(t, "ops.json", opsJSON)

	var method, path, contentType string
	var body []byte
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, contentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"atomic:results":[{"data":{"id":"new-org-id","type":"organization"}}]}`))
	}))
	defer ts.Close()

	out, err := executeCommand("operations", "--file", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != http.MethodPost {
		t.Errorf("expected POST, got %s", method)
	}
	if !strings.HasSuffix(path, "/operations") {
		t.Errorf("expected path ending in /operations, got %s", path)
	}
	if !strings.Contains(contentType, "https://jsonapi.org/ext/atomic") {
		t.Errorf("expected atomic ext content type, got %q", contentType)
	}
	var got struct {
		Ops []struct {
			Op   string                     `json:"op"`
			Data map[string]json.RawMessage `json:"data"`
		} `json:"atomic:operations"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, body)
	}
	if len(got.Ops) != 1 || got.Ops[0].Op != "add" || !strings.Contains(string(got.Ops[0].Data["attributes"]), `"acme"`) {
		t.Errorf("unexpected request body: %s", body)
	}
	if !strings.Contains(out, "new-org-id") {
		t.Errorf("expected output to contain result id, got: %s", out)
	}
}

func TestCmdOperationsAlias(t *testing.T) {
	resetGlobalFlags()
	file := writeTempFile(t, "ops.json", opsJSON)
	var hit atomic.Bool
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit.Store(true)
		_, _ = w.Write([]byte(`{"atomic:results":[]}`))
	}))
	defer ts.Close()

	if _, err := executeCommand("ops", "-f", file); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hit.Load() {
		t.Error("expected request to be sent via ops alias")
	}
}

func TestCmdOperationsErrors(t *testing.T) {
	tests := []struct {
		name    string
		file    func(t *testing.T) string
		status  int
		wantErr string
		wantReq int32
	}{
		{
			name:    "empty file flag",
			file:    func(*testing.T) string { return "" },
			wantErr: "flag --file is required",
		},
		{
			name:    "missing file",
			file:    func(t *testing.T) string { return filepath.Join(t.TempDir(), "nope.json") },
			wantErr: "reading operations file",
		},
		{
			name:    "invalid json",
			file:    func(t *testing.T) string { return writeTempFile(t, "bad.json", "{not json") },
			wantErr: "parsing operations JSON",
		},
		{
			name:    "api error",
			file:    func(t *testing.T) string { return writeTempFile(t, "ops.json", opsJSON) },
			status:  http.StatusInternalServerError,
			wantErr: "500",
			wantReq: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			var requests atomic.Int32
			ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				if tt.status != 0 {
					w.WriteHeader(tt.status)
				}
			}))
			defer ts.Close()

			_, err := executeCommand("operations", "--file", tt.file(t))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if n := requests.Load(); n != tt.wantReq {
				t.Errorf("expected %d requests, got %d", tt.wantReq, n)
			}
		})
	}
}

func TestCmdOperationsNotLoggedIn(t *testing.T) {
	resetGlobalFlags()
	file := writeTempFile(t, "ops.json", opsJSON)

	_, err := executeCommand("operations", "--file", file)
	if err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Fatalf("expected not-logged-in error, got: %v", err)
	}
}

func TestCmdOperationsRequiresFileFlag(t *testing.T) {
	resetGlobalFlags()

	_, err := executeCommand("operations")
	if err == nil || !strings.Contains(err.Error(), `"file"`) {
		t.Fatalf("expected required flag error for file, got: %v", err)
	}
}
