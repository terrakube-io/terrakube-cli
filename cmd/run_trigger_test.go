package cmd

import (
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

const trgSrc = "e5f6a7b8-c9d0-4234-8fab-345678901234"

// AC-35: a run trigger links a source workspace to the destination (-w), both
// by name, with --enabled and --template.
func TestCmdRunTriggerCreate(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"workspace": []*terrakube.Workspace{{ID: polWs, Name: "app"}, {ID: trgSrc, Name: "network"}},
		"template":  []*terrakube.Template{{ID: runTpl, Name: "Apply"}},
	}, &terrakube.RunTrigger{ID: "t1"})

	if _, err := executeCommand("run-trigger", "create", "-o", polOrg, "-w", "app", "--source", "network",
		"--template", "Apply", "--enabled=false"); err != nil {
		t.Fatal(err)
	}
	path, attrs, rels := onlyWrite(t, calls())
	if !strings.HasSuffix(path, "/organization/"+polOrg+"/workspace/"+polWs+"/runTriggers") {
		t.Errorf("path %s", path)
	}
	if v, ok := attrs["enabled"]; !ok || v != false || len(attrs) != 1 {
		t.Errorf("attributes %v", attrs)
	}
	if relID(rels, "sourceWorkspace") != trgSrc || relID(rels, "destinationWorkspace") != polWs || relID(rels, "template") != runTpl {
		t.Errorf("relationships %v", rels)
	}
}

// AC-35: --source is required on create; update changes enabled or template.
func TestCmdRunTriggerUpdate(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, nil, &terrakube.RunTrigger{ID: "t1"})
	if _, err := executeCommand("run-trigger", "create", "-o", polOrg, "-w", polWs); err == nil || !strings.Contains(err.Error(), `"source"`) {
		t.Fatalf("expected --source to be required, got %v", err)
	}
	if got := calls(); len(got) != 0 {
		t.Fatalf("expected no request, got %v", got)
	}
	if _, err := executeCommand("run-trigger", "update", "-o", polOrg, "-w", polWs, "--id", "t1", "--template", ""); err != nil {
		t.Fatal(err)
	}
	path, attrs, rels := onlyWrite(t, calls())
	if !strings.HasSuffix(path, "/runTriggers/t1") || len(attrs) != 0 || len(rels) != 1 {
		t.Errorf("path %s, attributes %v, relationships %v", path, attrs, rels)
	}
	if tpl, ok := rels["template"].(map[string]any); !ok || tpl["data"] != nil {
		t.Errorf("--template \"\" should clear the template, got %v", rels)
	}
}
