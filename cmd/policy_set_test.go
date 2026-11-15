package cmd

import (
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

const (
	polOrg = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	polSet = "b2c3d4e5-f6a7-4890-abcd-ef1234567890"
	polWs  = "d4e5f6a7-b8c9-0123-defa-234567890123"
)

func polName(s string) *string { return &s }

// AC-31: policy-set create sends only the set flags, and --vcs takes a name.
func TestCmdPolicySetCreate(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"vcs": []*terrakube.VCS{{ID: testVcsID, Name: "github"}},
	}, &terrakube.PolicySet{ID: polSet, Name: polName("baseline")})

	if _, err := executeCommand("policy-set", "create", "-o", polOrg, "--name", "baseline",
		"--enforcement-level", "SOFT_MANDATORY", "--global", "--repository", "https://git/x", "--vcs", "github"); err != nil {
		t.Fatal(err)
	}
	path, attrs, rels := onlyWrite(t, calls())
	if !strings.HasSuffix(path, "/organization/"+polOrg+"/policySet") {
		t.Errorf("path %s", path)
	}
	want := map[string]any{"name": "baseline", "enforcementLevel": "SOFT_MANDATORY", "global": true, "repository": "https://git/x"}
	if len(attrs) != len(want) {
		t.Errorf("attributes %v, want %v", attrs, want)
	}
	for k, v := range want {
		if attrs[k] != v {
			t.Errorf("%s = %v, want %v", k, attrs[k], v)
		}
	}
	if relID(rels, "vcs") != testVcsID {
		t.Errorf("vcs relationship %v", rels)
	}
}

// AC-31: policy-set-parameter lives under a policy set given by name.
func TestCmdPolicySetParameterCreate(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"policySet": []*terrakube.PolicySet{{ID: polSet, Name: polName("baseline")}, {ID: "other", Name: polName("baseline-2")}, {ID: "unnamed"}},
	}, &terrakube.PolicySetParameter{ID: "p1"})

	if _, err := executeCommand("policy-set-parameter", "create", "-o", polOrg, "--policy-set", "baseline",
		"--key", "max_size", "--value", "3"); err != nil {
		t.Fatal(err)
	}
	path, attrs, _ := onlyWrite(t, calls())
	if !strings.HasSuffix(path, "/policySet/"+polSet+"/parameters") {
		t.Errorf("path %s", path)
	}
	if attrs["key"] != "max_size" || attrs["value"] != "3" || len(attrs) != 2 {
		t.Errorf("attributes %v", attrs)
	}
}

// AC-31: an attachment links the policy set to one workspace, project or tag, by name.
func TestCmdPolicyAttachmentCreate(t *testing.T) {
	cases := []struct {
		flag, rel, list string
		items           any
	}{
		{"--workspace", "workspace", "workspace", []*terrakube.Workspace{{ID: polWs, Name: "prod"}}},
		{"--project", "project", "project", []*terrakube.Project{{ID: polWs, Name: "prod"}}},
		{"--tag", "tag", "tag", []*terrakube.Tag{{ID: polWs, Name: "prod"}}},
	}
	for _, c := range cases {
		t.Run(c.rel, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, map[string]any{c.list: c.items}, &terrakube.PolicyAttachment{ID: "a1"})
			if _, err := executeCommand("policy-attachment", "create", "-o", polOrg, "--policy-set", polSet, c.flag, "prod"); err != nil {
				t.Fatal(err)
			}
			path, attrs, rels := onlyWrite(t, calls())
			if !strings.HasSuffix(path, "/policySet/"+polSet+"/attachments") {
				t.Errorf("path %s", path)
			}
			if len(attrs) != 0 || len(rels) != 1 || relID(rels, c.rel) != polWs {
				t.Errorf("attributes %v, relationships %v", attrs, rels)
			}
		})
	}
}

// AC-32: zero targets or more than one fail before any request.
func TestCmdPolicyAttachmentNeedsExactlyOneTarget(t *testing.T) {
	cases := map[string][]string{
		"create without target": {"create"},
		"create with two":       {"create", "--workspace", polWs, "--tag", polWs},
		"create with three":     {"create", "--workspace", polWs, "--project", polWs, "--tag", polWs},
		"update with two":       {"update", "--id", polWs, "--project", polWs, "--tag", polWs},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, nil, nil)
			full := append([]string{"policy-attachment", args[0], "-o", "my-org", "--policy-set", "baseline"}, args[1:]...)
			_, err := executeCommand(full...)
			if err == nil || !strings.Contains(err.Error(), "workspace project tag") {
				t.Fatalf("expected a target error, got %v", err)
			}
			if got := calls(); len(got) != 0 {
				t.Errorf("expected no request, got %v", got)
			}
		})
	}
}
