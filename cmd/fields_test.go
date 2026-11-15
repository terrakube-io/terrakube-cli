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
