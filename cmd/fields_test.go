package cmd

import (
	"reflect"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

// fieldCase runs one create or update and expects exactly attrs in the body.
type fieldCase struct {
	name  string
	args  []string
	lists map[string]any
	resp  any
	attrs map[string]any
}

func runFieldCases(t *testing.T, cases []fieldCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, c.lists, c.resp)
			if _, err := executeCommand(c.args...); err != nil {
				t.Fatal(err)
			}
			_, attrs, _ := onlyWrite(t, calls())
			if !reflect.DeepEqual(attrs, c.attrs) {
				t.Errorf("attributes %v, want %v", attrs, c.attrs)
			}
		})
	}
}

const fieldID = "99999999-9999-4999-8999-999999999999"

// AC-36 (plan 2.5): workspace settings are sent as attributes.
func TestCmdWorkspaceSettingsFields(t *testing.T) {
	ws := &terrakube.Workspace{ID: polWs, Name: "prod"}
	runFieldCases(t, []fieldCase{
		{
			name: "update booleans",
			args: []string{
				"workspace", "update", "-o", polOrg, "--id", polWs, "--require-approve-to-apply",
				"--restrict-apply-to-configured-branch", "--allow-remote-apply", "--global-remote-state=false",
			},
			resp: ws,
			attrs: map[string]any{
				"requireApproveToApply": true, "restrictApplyToConfiguredBranch": true,
				"allowRemoteApply": true, "globalRemoteState": false,
			},
		},
		{
			name:  "shared ids and module ssh key by name",
			args:  []string{"workspace", "update", "-o", polOrg, "--id", polWs, "--shared-ids", "a,b", "--module-ssh-key", "deploy"},
			lists: map[string]any{"ssh": []*terrakube.SSH{{ID: fieldID, Name: "deploy"}}},
			resp:  ws,
			attrs: map[string]any{"sharedIds": "a,b", "moduleSshKey": fieldID},
		},
		{
			name:  "clear module ssh key (null)",
			args:  []string{"workspace", "update", "-o", polOrg, "--id", polWs, "--module-ssh-key", ""},
			resp:  ws,
			attrs: map[string]any{"moduleSshKey": nil},
		},
	})
}

// AC-36 (plan 2.6): plan, approve and destroy permissions and roles.
func TestCmdWorkspaceAccessFields(t *testing.T) {
	runFieldCases(t, []fieldCase{{
		name: "create",
		args: []string{
			"workspace-access", "create", "-o", polOrg, "-w", polWs, "--name", "ops",
			"--plan-job", "--approve-job", "--destroy-job=false", "--role", "writer",
		},
		resp:  &terrakube.WorkspaceAccess{ID: "a1"},
		attrs: map[string]any{"name": "ops", "planJob": true, "approveJob": true, "destroyJob": false, "role": "writer"},
	}})
}

func TestCmdTeamFields(t *testing.T) {
	runFieldCases(t, []fieldCase{{
		name:  "update",
		args:  []string{"team", "update", "-o", polOrg, "--id", fieldID, "--manage-policies", "--destroy-job"},
		resp:  &terrakube.Team{ID: fieldID},
		attrs: map[string]any{"managePolicies": true, "destroyJob": true},
	}})
}

func TestCmdProjectAccessFields(t *testing.T) {
	runFieldCases(t, []fieldCase{{
		name:  "update",
		args:  []string{"project-access", "update", "-o", polOrg, "--project", fieldID, "--id", fieldID, "--destroy-job"},
		resp:  &terrakube.ProjectAccess{ID: fieldID},
		attrs: map[string]any{"destroyJob": true},
	}})
}

// AC-36 (plan 2.7): registry status, deprecation and namespace fields.
func TestCmdProviderAndModuleVersionFields(t *testing.T) {
	runFieldCases(t, []fieldCase{
		{
			name:  "module-version",
			args:  []string{"module-version", "update", "-o", polOrg, "--module", fieldID, "--id", fieldID, "--status", "DEPRECATED", "--deprecation-message", "use 2.x"},
			resp:  &terrakube.ModuleVersion{ID: fieldID},
			attrs: map[string]any{"status": "DEPRECATED", "deprecationMessage": "use 2.x"},
		},
		{
			name:  "provider-version",
			args:  []string{"provider-version", "update", "-o", polOrg, "--provider", fieldID, "--id", fieldID, "--status", "DEPRECATED", "--deprecation-message", "use 2.x"},
			resp:  &terrakube.ProviderVersion{ID: fieldID},
			attrs: map[string]any{"status": "DEPRECATED", "deprecationMessage": "use 2.x"},
		},
		{
			name:  "provider",
			args:  []string{"provider", "create", "-o", polOrg, "--name", "aws", "--imported", "--registry-namespace", "hashicorp"},
			resp:  &terrakube.Provider{ID: fieldID},
			attrs: map[string]any{"name": "aws", "imported": true, "registryNamespace": "hashicorp"},
		},
	})
}

// AC-36 (plan 2.7): module --vcs and --ssh are relationships, by name.
func TestCmdModuleFields(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"vcs": []*terrakube.VCS{{ID: testVcsID, Name: "github"}},
		"ssh": []*terrakube.SSH{{ID: fieldID, Name: "deploy"}},
	}, &terrakube.Module{ID: fieldID})
	if _, err := executeCommand("module", "create", "-o", polOrg, "--name", "vpc", "--vcs", "github", "--ssh", "deploy"); err != nil {
		t.Fatal(err)
	}
	_, attrs, rels := onlyWrite(t, calls())
	if len(attrs) != 1 || relID(rels, "vcs") != testVcsID || relID(rels, "ssh") != fieldID {
		t.Errorf("attributes %v, relationships %v", attrs, rels)
	}
}

// AC-36 (plan 2.8): leftover fields of job, schedule, workspace tag, variable and template.
func TestCmdJobScheduleTagVariableTemplateFields(t *testing.T) {
	tpl := writeTempFile(t, "flow.yaml", "flow:\n  - type: terraformPlan\n")
	runFieldCases(t, []fieldCase{
		{
			name: "job variables",
			args: []string{
				"job", "create", "-o", polOrg, "-w", polWs, "--template-reference", runTpl,
				"--variables", "region=eu-west-1", "--variables", "size=3",
			},
			resp:  &terrakube.Job{ID: "42"},
			attrs: map[string]any{"templateReference": runTpl, "variables": map[string]any{"region": "eu-west-1", "size": "3"}},
		},
		{
			name:  "workspace-schedule",
			args:  []string{"workspace-schedule", "update", "-o", polOrg, "-w", polWs, "--id", fieldID, "--enabled=false", "--description", "nightly"},
			resp:  &terrakube.WorkspaceSchedule{ID: fieldID},
			attrs: map[string]any{"enabled": false, "description": "nightly"},
		},
		{
			name:  "workspace-tag",
			args:  []string{"workspace-tag", "update", "-o", polOrg, "-w", polWs, "--id", fieldID, "--value", "prod"},
			resp:  &terrakube.WorkspaceTag{ID: fieldID},
			attrs: map[string]any{"value": "prod"},
		},
		{
			name:  "variable",
			args:  []string{"variable", "update", "-o", polOrg, "-w", polWs, "--id", fieldID, "--incomplete"},
			resp:  &terrakube.Variable{ID: fieldID},
			attrs: map[string]any{"incomplete": true},
		},
		{
			name:  "template content file",
			args:  []string{"template", "create", "-o", polOrg, "--name", "Plan", "--content-file", tpl},
			resp:  &terrakube.Template{ID: fieldID},
			attrs: map[string]any{"name": "Plan", "tcl": "flow:\n  - type: terraformPlan\n"},
		},
	})
}

// AC-36 (plan 2.8): notification-configuration --templates replaces the
// template set, by name; "" empties it.
func TestCmdNotificationConfigurationFields(t *testing.T) {
	for _, c := range []struct {
		arg  string
		want []any
	}{
		{"Plan,Apply", []any{runTpl, runTpl2}},
		{"", []any{}},
	} {
		t.Run(c.arg, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, map[string]any{
				"template": []*terrakube.Template{{ID: runTpl, Name: "Plan"}, {ID: runTpl2, Name: "Apply"}},
			}, &terrakube.NotificationConfiguration{ID: fieldID})
			if _, err := executeCommand("notification-configuration", "update", "-o", polOrg, "--id", fieldID, "--templates", c.arg); err != nil {
				t.Fatal(err)
			}
			_, _, rels := onlyWrite(t, calls())
			tpls, _ := rels["templates"].(map[string]any)
			data, ok := tpls["data"].([]any)
			if !ok || len(data) != len(c.want) {
				t.Fatalf("templates relationship %v", rels)
			}
			for i, d := range data {
				if m := d.(map[string]any); m["id"] != c.want[i] || m["type"] != "template" {
					t.Errorf("templates[%d] = %v", i, m)
				}
			}
		})
	}
}
