package cmd

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
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

// --- Flat struct (terrakube-go) tests ---

func TestSplitInterface_FlatStruct_Single(t *testing.T) {
	desc := "Test org"
	org := terrakube.Organization{
		ID:            "org-flat-1",
		Name:          "flat-org",
		Description:   &desc,
		ExecutionMode: "remote",
	}

	rows, headers := splitInterface(org)

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if headers[0] != "ID" {
		t.Errorf("first header should be ID, got %q", headers[0])
	}
	if rows[0][0] != "org-flat-1" {
		t.Errorf("expected ID org-flat-1, got %q", rows[0][0])
	}

	// Find Name column
	nameIdx := -1
	for i, h := range headers {
		if h == "Name" {
			nameIdx = i
			break
		}
	}
	if nameIdx == -1 {
		t.Fatalf("missing Name header, got %v", headers)
	}
	if rows[0][nameIdx] != "flat-org" {
		t.Errorf("expected Name flat-org, got %q", rows[0][nameIdx])
	}
}

func TestSplitInterface_FlatStruct_Slice(t *testing.T) {
	orgs := []*terrakube.Organization{
		{ID: "org-a", Name: "alpha"},
		{ID: "org-b", Name: "beta"},
		{ID: "org-c", Name: "gamma"},
	}

	rows, headers := splitInterface(orgs)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	expectedIDs := []string{"org-a", "org-b", "org-c"}
	for i, id := range expectedIDs {
		if rows[i][0] != id {
			t.Errorf("row %d: expected ID %q, got %q", i, id, rows[i][0])
		}
	}

	// Verify no duplicate headers
	seen := make(map[string]int)
	for _, h := range headers {
		seen[h]++
		if seen[h] > 1 {
			t.Errorf("duplicate header %q", h)
		}
	}
}

func TestSplitInterface_FlatStruct_BoolFields(t *testing.T) {
	team := terrakube.Team{
		ID:              "team-bools",
		Name:            "admins",
		ManageWorkspace: true,
		ManageModule:    false,
		ManageState:     false,
	}

	rows, headers := splitInterface(team)

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// Find bool columns
	for i, h := range headers {
		switch h {
		case "ManageWorkspace":
			if rows[0][i] != "true" {
				t.Errorf("ManageWorkspace: expected true, got %q", rows[0][i])
			}
		case "ManageModule":
			if rows[0][i] != "false" {
				t.Errorf("ManageModule: expected false, got %q", rows[0][i])
			}
		case "ManageState":
			if rows[0][i] != "false" {
				t.Errorf("ManageState: expected false, got %q", rows[0][i])
			}
		}
	}
}

func TestSplitInterface_FlatStruct_PointerFields(t *testing.T) {
	desc := "A workspace"
	org := terrakube.Organization{
		ID:          "org-ptrs",
		Name:        "ptr-org",
		Description: &desc,
		Icon:        nil,
	}

	rows, headers := splitInterface(org)

	descIdx := -1
	iconIdx := -1
	for i, h := range headers {
		switch h {
		case "Description":
			descIdx = i
		case "Icon":
			iconIdx = i
		}
	}

	if descIdx == -1 {
		t.Fatal("missing Description header")
	}
	if rows[0][descIdx] != "A workspace" {
		t.Errorf("Description: expected 'A workspace', got %q", rows[0][descIdx])
	}

	if iconIdx == -1 {
		t.Fatal("missing Icon header")
	}
	if rows[0][iconIdx] != "" {
		t.Errorf("nil Icon: expected empty, got %q", rows[0][iconIdx])
	}
}

func TestSplitInterface_FlatStruct_SkipsRelations(t *testing.T) {
	// Workspace has a Vcs relation field -- it should be excluded from table output
	ws := terrakube.Workspace{
		ID:   "ws-rel",
		Name: "rel-test",
	}

	_, headers := splitInterface(ws)

	for _, h := range headers {
		if h == "Vcs" {
			t.Error("Vcs relation field should be excluded from headers")
		}
	}
}

func TestSplitInterface_EmptySlice(t *testing.T) {
	empty := []*terrakube.Workspace{}

	rows, headers := splitInterface(empty)

	if len(rows) != 0 {
		t.Errorf("expected 0 rows for empty slice, got %d", len(rows))
	}

	// Only the "ID" header should exist (no element to inspect)
	if len(headers) != 1 {
		t.Errorf("expected 1 header (ID) for empty slice, got %d: %v", len(headers), headers)
	}
	if headers[0] != "ID" {
		t.Errorf("expected header 'ID', got %q", headers[0])
	}
}

// splitInterface panics on non-struct input (e.g. a plain string) because
// reflect operations on a non-struct value cannot find fields.
func TestSplitInterface_NonStructInput(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when passing a plain string to splitInterface, but did not panic")
		}
		t.Logf("confirmed panic on non-struct input: %v", r)
	}()

	splitInterface("not a struct")
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

// Unknown output format silently produces nothing -- no error or warning.
func TestRenderOutput_UnknownFormat(t *testing.T) {
	ws := terrakube.Workspace{
		ID:   "ws-unknown-fmt",
		Name: "invisible",
	}

	got := captureStdout(t, func() {
		renderOutput(ws, "xml")
	})

	// The switch in renderOutput has no default case, so unknown formats
	// silently fall through and produce zero output.
	if got != "" {
		t.Errorf("expected no output for unknown format 'xml', got %q", got)
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

	// Flat struct: fields at top level, no nested "attributes"
	if parsed["Name"] != "json-org" {
		t.Errorf("expected Name json-org, got %v", parsed["Name"])
	}
	if _, has := parsed["attributes"]; has {
		t.Error("flat struct JSON should not have nested 'attributes' key")
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
