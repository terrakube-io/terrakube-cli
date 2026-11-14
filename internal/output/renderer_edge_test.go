package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write fail") }

func TestRenderJSON_WriteError(t *testing.T) {
	err := Render(failWriter{}, singleResource(), "json")
	if err == nil || err.Error() != "write fail" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderJSON_NonJSONAPIData(t *testing.T) {
	for name, tc := range map[string]struct {
		data any
		want string
	}{
		"plain struct":   {struct{ A int }{1}, "{\n    \"A\": 1\n}\n"},
		"string":         {"hi", "\"hi\"\n"},
		"nil pointer":    {(*sampleResource)(nil), "null\n"},
		"nil":            {nil, "null\n"},
		"non-api slice":  {[]int{1, 2}, "[\n    1,\n    2\n]\n"},
		"no primary tag": {[]struct{ ID string }{{"x"}}, "[\n    {\n        \"ID\": \"x\"\n    }\n]\n"},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Render(&buf, tc.data, "json"); err != nil {
				t.Fatal(err)
			}
			if buf.String() != tc.want {
				t.Errorf("got %q want %q", buf.String(), tc.want)
			}
		})
	}
}

func TestRenderJSON_StructValueIsAddressed(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, *singleResource(), "json"); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["id"] != "abc-123" || m["type"] != "sample" {
		t.Errorf("expected JSON:API unwrapped object, got %v", m)
	}
}

func TestRenderJSON_HideNulls(t *testing.T) {
	r := &sampleResource{ID: "a", Name: "n"} // nil Description
	render := func() map[string]any {
		var buf bytes.Buffer
		if err := Render(&buf, r, "json"); err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		return m["attributes"].(map[string]any)
	}

	if _, ok := render()["description"]; ok {
		t.Error("HideNulls=true: expected null description stripped")
	}

	orig := HideNulls
	HideNulls = false
	t.Cleanup(func() { HideNulls = orig })
	attrs := render()
	v, ok := attrs["description"]
	if !ok || v != nil {
		t.Errorf("HideNulls=false: expected description null kept, got %v (present=%v)", v, ok)
	}
}

type badPrimary struct {
	ID bool `jsonapi:"primary,bad"`
}

func TestRenderJSON_MarshalPayloadFailureFallsBack(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, &badPrimary{ID: true}, "json"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "{\n    \"ID\": true\n}\n" {
		t.Errorf("unexpected fallback output %q", buf.String())
	}
}

func TestStripNulls(t *testing.T) {
	got, err := stripNulls(json.RawMessage(`{"a":null,"b":{"c":null,"d":1},"e":[{"f":null,"g":2},null]}`))
	if err != nil {
		t.Fatal(err)
	}
	// Nulls inside objects (nested and within arrays) go; array elements stay.
	if string(got) != `{"b":{"d":1},"e":[{"g":2},null]}` {
		t.Errorf("got %s", got)
	}

	if _, err := stripNulls(json.RawMessage(`{bad`)); err == nil {
		t.Error("expected error for invalid json")
	}
}

type yamlFail struct{}

func (yamlFail) MarshalYAML() (any, error) { return nil, errors.New("yaml fail") }

func TestRenderYAML_MarshalError(t *testing.T) {
	err := Render(&bytes.Buffer{}, yamlFail{}, "yaml")
	if err == nil || err.Error() != "failed to marshal yaml: yaml fail" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderYAML_WriteError(t *testing.T) {
	err := Render(failWriter{}, singleResource(), "yaml")
	if err == nil || err.Error() != "write fail" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderTSV_WriteError(t *testing.T) {
	err := Render(failWriter{}, singleResource(), "tsv")
	if err == nil || err.Error() != "write fail" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderTable_EmptySliceWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, []*sampleResource{}, "table"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

type formatted struct {
	ID    string
	At    *time.Time
	Num   *int
	Flag  *bool
	Label *string
	Big   int64
	Small int8
	Text  string
	Skip  *nested `jsonapi:"relation,nested"`
	priv  string
}

func TestRenderTSV_FieldFormatting(t *testing.T) {
	at := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	n, b, s := 7, true, "lbl"
	d := &formatted{ID: "f1", At: &at, Num: &n, Flag: &b, Label: &s, Big: -9, Small: 3, Text: "t", Skip: &nested{ID: "r"}, priv: "p"}

	var buf bytes.Buffer
	if err := Render(&buf, d, "tsv"); err != nil {
		t.Fatal(err)
	}
	want := "f1\t2026-01-02T15:04:05Z\t7\ttrue\tlbl\t-9\t3\tt\n"
	if buf.String() != want {
		t.Errorf("got %q want %q", buf.String(), want)
	}

	buf.Reset()
	if err := Render(&buf, &formatted{ID: "f2"}, "tsv"); err != nil {
		t.Fatal(err)
	}
	if want := "f2\t\t\t\t\t0\t0\t\n"; buf.String() != want {
		t.Errorf("nil pointers: got %q want %q", buf.String(), want)
	}
}

func TestRenderJSON_MarshalError(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, map[string]any{"c": make(chan int)}, "json")
	if err == nil || err.Error() != "failed to marshal json: json: unsupported type: chan int" {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}
