package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type sampleResource struct {
	ID          string  `jsonapi:"primary,sample"`
	Name        string  `jsonapi:"attr,name"`
	Description *string `jsonapi:"attr,description"`
	Active      bool    `jsonapi:"attr,active"`
	Count       int     `jsonapi:"attr,count"`
	Related     *nested `jsonapi:"relation,nested"`
}

type nested struct {
	ID string `jsonapi:"primary,nested"`
}

func strPtr(s string) *string { return &s }

func singleResource() *sampleResource {
	return &sampleResource{
		ID:          "abc-123",
		Name:        "test-item",
		Description: strPtr("a description"),
		Active:      true,
		Count:       42,
		Related:     &nested{ID: "rel-1"},
	}
}

func resourceSlice() []*sampleResource {
	return []*sampleResource{
		singleResource(),
		{
			ID:     "def-456",
			Name:   "other-item",
			Active: false,
			Count:  0,
		},
	}
}

func TestRenderJSON_Single(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	if m["ID"] != "abc-123" {
		t.Errorf("expected ID abc-123, got %v", m["ID"])
	}
	if m["Name"] != "test-item" {
		t.Errorf("expected Name test-item, got %v", m["Name"])
	}
	if m["Active"] != true {
		t.Errorf("expected Active true, got %v", m["Active"])
	}
}

func TestRenderJSON_Slice(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, resourceSlice(), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		t.Fatalf("output is not valid json array: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRenderYAML_Single(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id: abc-123") {
		t.Errorf("expected yaml to contain 'id: abc-123', got:\n%s", out)
	}
	if !strings.Contains(out, "name: test-item") {
		t.Errorf("expected yaml to contain 'name: test-item', got:\n%s", out)
	}
	if !strings.Contains(out, "active: true") {
		t.Errorf("expected yaml to contain 'active: true', got:\n%s", out)
	}
}

func TestRenderYAML_Slice(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, resourceSlice(), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "abc-123") || !strings.Contains(out, "def-456") {
		t.Errorf("expected yaml to contain both IDs, got:\n%s", out)
	}
}

func TestRenderTable_Single(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "abc-123") {
		t.Errorf("expected table to contain ID, got:\n%s", out)
	}
	if !strings.Contains(out, "test-item") {
		t.Errorf("expected table to contain Name, got:\n%s", out)
	}
	// Relation field should be skipped
	if strings.Contains(out, "rel-1") {
		t.Errorf("expected table to skip relation field, got:\n%s", out)
	}
}

func TestRenderTable_Slice(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, resourceSlice(), "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "abc-123") || !strings.Contains(out, "def-456") {
		t.Errorf("expected table to contain both IDs, got:\n%s", out)
	}
}

func TestRenderTSV_Single(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "tsv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	cols := strings.Split(lines[0], "\t")
	if cols[0] != "abc-123" {
		t.Errorf("expected first column abc-123, got %s", cols[0])
	}
	if cols[1] != "test-item" {
		t.Errorf("expected second column test-item, got %s", cols[1])
	}
}

func TestRenderTSV_Slice(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, resourceSlice(), "tsv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestRenderNone(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for 'none' format, got %q", buf.String())
	}
}

func TestRenderUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected unsupported error, got: %v", err)
	}
}

func TestRenderTable_NilPointerField(t *testing.T) {
	r := &sampleResource{
		ID:     "nil-ptr",
		Name:   "no-desc",
		Active: false,
	}
	var buf bytes.Buffer
	err := Render(&buf, r, "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "nil-ptr") {
		t.Errorf("expected table to contain ID, got:\n%s", out)
	}
}

func TestRenderTSV_BoolFormatting(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, singleResource(), "tsv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "true") {
		t.Errorf("expected true in tsv output, got:\n%s", out)
	}
}
