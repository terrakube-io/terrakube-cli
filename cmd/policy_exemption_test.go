package cmd

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"
)

const polEval = "c3d4e5f6-a7b8-4901-9cde-f12345678901"

// AC-33: policy-exemption create sends the rule, justification, expiry,
// ticket, policy set and target, with names resolved.
func TestCmdPolicyExemptionCreate(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"policySet": []*terrakube.PolicySet{{ID: polSet, Name: polName("baseline")}},
		"workspace": []*terrakube.Workspace{{ID: polWs, Name: "prod"}},
	}, &terrakube.PolicyExemption{ID: "e1"})

	if _, err := executeCommand("policy-exemption", "create", "-o", polOrg, "--policy-set", "baseline",
		"--rule", "deny_public_s3", "--justification", "legacy bucket", "--expires-at", "2027-01-31T00:00:00Z",
		"--ticket", "SEC-1", "--workspace", "prod"); err != nil {
		t.Fatal(err)
	}
	path, attrs, rels := onlyWrite(t, calls())
	if !strings.HasSuffix(path, "/organization/"+polOrg+"/policyExemption") {
		t.Errorf("path %s", path)
	}
	want := map[string]any{"ruleId": "deny_public_s3", "justification": "legacy bucket", "expiresAt": "2027-01-31T00:00:00Z", "ticketReference": "SEC-1"}
	if len(attrs) != len(want) {
		t.Errorf("attributes %v, want %v", attrs, want)
	}
	for k, v := range want {
		if attrs[k] != v {
			t.Errorf("%s = %v, want %v", k, attrs[k], v)
		}
	}
	if relID(rels, "policySet") != polSet || relID(rels, "workspace") != polWs || len(rels) != 2 {
		t.Errorf("relationships %v", rels)
	}
}

// AC-33: --rule, --justification and --expires-at are required, and the
// target is a workspace or a project, not both. Each fails before any request.
func TestCmdPolicyExemptionValidation(t *testing.T) {
	full := []string{"--policy-set", polSet, "--rule", "r", "--justification", "j", "--expires-at", "2027-01-31T00:00:00Z", "--ticket", "OPS-1"}
	without := func(flag string) []string {
		var out []string
		for i := 0; i < len(full); i += 2 {
			if full[i] != flag {
				out = append(out, full[i], full[i+1])
			}
		}
		return out
	}
	cases := map[string]struct {
		args []string
		want string
	}{
		"no rule":          {without("--rule"), `"rule"`},
		"no justification": {without("--justification"), `"justification"`},
		"no expiry":        {without("--expires-at"), `"expires-at"`},
		"no ticket":        {without("--ticket"), `"ticket"`},
		"two targets":      {append(append([]string{}, full...), "--workspace", polWs, "--project", polWs), "workspace project"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, nil, nil)
			_, err := executeCommand(append([]string{"policy-exemption", "create", "-o", "my-org"}, c.args...)...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error naming %s, got %v", c.want, err)
			}
			if got := calls(); len(got) != 0 {
				t.Errorf("expected no request, got %v", got)
			}
		})
	}
}

// AC-34: policy-evaluation has list and get only.
func TestCmdPolicyEvaluationReadOnly(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, map[string]any{
		"policyEvaluation": []*terrakube.PolicyEvaluation{{ID: polEval, Status: polName("FAILED")}},
	}, nil)
	out, err := executeCommand("policy-evaluation", "list", "-o", polOrg, "--job", "42")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, polEval) || !strings.HasSuffix(calls()[0].path, "/organization/"+polOrg+"/job/42/policyEvaluation") {
		t.Errorf("out %s, calls %v", out, calls())
	}
	var verbs []string
	for _, sub := range findSubCmd("policy-evaluation").Commands() {
		verbs = append(verbs, sub.Name())
	}
	if strings.Join(verbs, ",") != "get,list" {
		t.Errorf("policy-evaluation verbs %v, want get and list only", verbs)
	}
}

// fakeJWT returns an unsigned token whose payload is claims.
func fakeJWT(claims string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc([]byte(claims)) + ".sig"
}

// AC-34: policy-override create sends the justification, plus who overrode
// and when, which the server stores as required columns.
func TestCmdPolicyOverrideCreate(t *testing.T) {
	cases := []struct {
		token, by string
	}{
		{fakeJWT(`{"email":"ann@example.com","name":"Ann (Token)"}`), "ann@example.com"},
		{fakeJWT(`{"preferred_username":"ann"}`), "ann"},
		{fakeJWT(`{"name":"Ann (Token)"}`), "Ann (Token)"},
		{fakeJWT(`not json`), "cli-user"},
		{"opaque-token", "cli-user"},
	}
	for _, c := range cases {
		t.Run(c.by, func(t *testing.T) {
			resetGlobalFlags()
			calls := fakeAPI(t, nil, &terrakube.PolicyOverride{ID: "o1"})
			viper.Set("token", c.token)
			before := time.Now().UTC().Truncate(time.Second)
			if _, err := executeCommand("policy-override", "create", "-o", polOrg, "--job", "42",
				"--policy-evaluation", polEval, "--justification", "accepted risk"); err != nil {
				t.Fatal(err)
			}
			path, attrs, _ := onlyWrite(t, calls())
			if !strings.HasSuffix(path, "/job/42/policyEvaluation/"+polEval+"/override") {
				t.Errorf("path %s", path)
			}
			if attrs["justification"] != "accepted risk" || attrs["overriddenBy"] != c.by {
				t.Errorf("attributes %v", attrs)
			}
			at, err := time.Parse(time.RFC3339, attrs["overriddenAt"].(string))
			if err != nil || at.Before(before) || at.After(time.Now().Add(time.Second)) {
				t.Errorf("overriddenAt %v (%v)", attrs["overriddenAt"], err)
			}
		})
	}
}

// AC-34: --justification is required.
func TestCmdPolicyOverrideNeedsJustification(t *testing.T) {
	resetGlobalFlags()
	calls := fakeAPI(t, nil, nil)
	_, err := executeCommand("policy-override", "create", "-o", polOrg, "--job", "42", "--policy-evaluation", polEval)
	if err == nil || !strings.Contains(err.Error(), `"justification"`) {
		t.Fatalf("expected a required-flag error, got %v", err)
	}
	if got := calls(); len(got) != 0 {
		t.Errorf("expected no request, got %v", got)
	}
}
