package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

// --- renderOutput tests ---

func TestRenderOutput_None(t *testing.T) {
	ws := terrakube.Workspace{
		ID:   "ws-ghost",
		Name: "invisible",
	}

	got := captureStdout(t, func() {
		renderOutput(ws, "none")
	})

	if got != "" {
		t.Errorf("expected no output for 'none' format, got %q", got)
	}
}

// Unknown output format now triggers a fatal error via outputpkg.Render.
// We verify this by checking that Render returns an error directly.
func TestRenderOutput_UnknownFormat(t *testing.T) {
	ws := &terrakube.Workspace{
		ID:   "ws-unknown-fmt",
		Name: "invisible",
	}

	var buf bytes.Buffer
	err := outputpkg.Render(&buf, ws, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format 'xml'")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected unsupported error, got: %v", err)
	}
}

func TestRenderOutput_FlatStruct_JSON(t *testing.T) {
	org := terrakube.Organization{
		ID:            "org-json",
		Name:          "json-org",
		ExecutionMode: "remote",
	}

	got := captureStdout(t, func() {
		renderOutput(org, "json")
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, got)
	}

	// JSON:API format: {type, id, attributes: {...}}
	if parsed["id"] != "org-json" {
		t.Errorf("expected id org-json, got %v", parsed["id"])
	}
	if parsed["type"] != "organization" {
		t.Errorf("expected type organization, got %v", parsed["type"])
	}
	attrs, ok := parsed["attributes"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected attributes object, got %T", parsed["attributes"])
	}
	if attrs["name"] != "json-org" {
		t.Errorf("expected name json-org, got %v", attrs["name"])
	}
}

func TestRenderOutput_FlatStruct_Table(t *testing.T) {
	teams := []*terrakube.Team{
		{ID: "team-1", Name: "devs", ManageWorkspace: true, ManageModule: false},
	}

	got := captureStdout(t, func() {
		renderOutput(teams, "table")
	})

	if !strings.Contains(got, "team-1") {
		t.Error("table output should contain team ID")
	}
	if !strings.Contains(got, "devs") {
		t.Error("table output should contain team name")
	}
	upper := strings.ToUpper(got)
	if !strings.Contains(upper, "MANAGEWORKSPACE") {
		t.Error("table header should contain MANAGEWORKSPACE")
	}
}

func TestRenderOutput_FlatStruct_TSV(t *testing.T) {
	orgs := []*terrakube.Organization{
		{ID: "org-tsv", Name: "tsv-org", ExecutionMode: "local"},
	}

	got := captureStdout(t, func() {
		renderOutput(orgs, "tsv")
	})

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 TSV line, got %d", len(lines))
	}
	fields := strings.Split(lines[0], "\t")
	if fields[0] != "org-tsv" {
		t.Errorf("first field should be ID, got %q", fields[0])
	}
}
