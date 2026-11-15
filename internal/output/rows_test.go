package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderRows(t *testing.T) {
	rows := [][]string{{"a", "1"}, {"b", "2"}}
	var buf bytes.Buffer
	if err := RenderRows(&buf, []string{"name", "value"}, rows, "tsv"); err != nil || buf.String() != "a\t1\nb\t2\n" {
		t.Errorf("tsv %q, err %v", buf.String(), err)
	}
	buf.Reset()
	if err := RenderRows(&buf, []string{"name", "value"}, rows, "table"); err != nil || !strings.Contains(buf.String(), "NAME") || !strings.Contains(buf.String(), "b") {
		t.Errorf("table %q, err %v", buf.String(), err)
	}
	if err := RenderRows(&buf, nil, rows, "json"); err == nil {
		t.Error("expected an error for json")
	}
}
