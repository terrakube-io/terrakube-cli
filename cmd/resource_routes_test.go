package cmd

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
)

const routeAPI = "/api/v1/"

// Distinct UUIDs per parent position, so a closure that passes parent IDs in
// the wrong order (or the wrong one) produces a visibly different path.
var routeParents = []string{
	"11111111-1111-4111-8111-111111111111",
	"22222222-2222-4222-8222-222222222222",
	"33333333-3333-4333-8333-333333333333",
}

const routeID = "99999999-9999-4999-8999-999999999999"

// routeCase describes one registered resource. base and itemBase are the
// expected API paths taken from the terrakube-go client source ({n} is the
// n-th parent flag value); itemBase defaults to base when empty. verbs
// defaults to all five.
type routeCase struct {
	name     string // subtest name, defaults to cmd
	cmd      string
	typ      string // JSON:API type of the response body
	parents  []string
	base     string
	itemBase string
	create   []string // required create flags
	verbs    []string
	extra    []string // extra args for every verb (e.g. an optional parent)
}

var routeAllVerbs = []string{"list", "get", "create", "update", "delete"}

const (
	routeOrg   = "organization/{1}/"
	routeOrgWs = routeOrg + "workspace/{2}/"
)

var routeCases = []routeCase{
	{cmd: "action", typ: "action", base: "action", create: []string{"--name", "n"}},
	{cmd: "address", typ: "address", parents: []string{"organization", "job"}, base: routeOrg + "job/{2}/address", create: []string{"--name", "n"}},
	{cmd: "agent", typ: "agent", parents: []string{"organization"}, base: routeOrg + "agent", create: []string{"--name", "n", "--url", "http://a"}},
	{cmd: "collection", typ: "collection", parents: []string{"organization"}, base: routeOrg + "collection", create: []string{"--name", "n"}},
	{cmd: "collection-item", typ: "item", parents: []string{"organization", "collection"}, base: routeOrg + "collection/{2}/item", create: []string{"--key", "k", "--value", "v"}},
	{
		cmd: "collection-reference", typ: "reference", parents: []string{"organization", "collection", "workspace"},
		base: routeOrg + "collection/{2}/reference", itemBase: "reference",
	},
	{cmd: "federated", typ: "federated", base: "federated", create: []string{"--audience", "a", "--issuer-url", "http://i", "--name", "n"}},
	{cmd: "federated-claim", typ: "federated_claim", parents: []string{"federated"}, base: "federated/{1}/claims", create: []string{"--claim-key", "k", "--claim-value", "v"}},
	{cmd: "github-app-token", typ: "github_app_token", base: "github_app_token", create: []string{"--app-id", "1", "--installation-id", "2", "--owner", "o"}},
	{cmd: "history", typ: "history", parents: []string{"organization", "workspace"}, base: routeOrgWs + "history"},
	{
		cmd: "implementation", typ: "implementation", parents: []string{"organization", "provider", "provider-version"},
		base: routeOrg + "provider/{2}/version/{3}/implementation", create: []string{"--arch", "a", "--filename", "f", "--os", "o"},
	},
	{cmd: "job", typ: "job", parents: []string{"organization", "workspace"}, base: routeOrg + "job", create: []string{"--template-reference", "t"}},
	{cmd: "module", typ: "module", parents: []string{"organization"}, base: routeOrg + "module", create: []string{"--name", "n"}},
	{cmd: "module-version", typ: "module_version", parents: []string{"organization", "module"}, base: routeOrg + "module/{2}/version", create: []string{"--version", "1.0.0"}},
	{
		cmd: "notification-configuration", typ: "notification_configuration", parents: []string{"organization"},
		base: routeOrg + "notificationConfiguration", create: []string{"--name", "n", "--channel-type", "SLACK", "--destination-url", "http://d"},
	},
	{
		name: "notification-configuration-with-workspace", cmd: "notification-configuration", typ: "notification_configuration",
		parents: []string{"organization"}, base: routeOrgWs + "notificationConfiguration",
		create: []string{"--name", "n", "--channel-type", "SLACK", "--destination-url", "http://d"},
		extra:  []string{"--workspace", routeParents[1]}, verbs: []string{"list", "create"},
	},
	{
		cmd: "notification-trigger", typ: "notification_trigger", parents: []string{"organization", "notification-configuration"},
		base: routeOrg + "notificationConfiguration/{2}/triggers", create: []string{"--job-status", "COMPLETED"},
	},
	{cmd: "organization", typ: "organization", base: "organization", create: []string{"--name", "n"}},
	{
		cmd: "organization-variable", typ: "globalvar", parents: []string{"organization"}, base: routeOrg + "globalvar",
		create: []string{"--category", "TERRAFORM", "--key", "k", "--value", "v"},
	},
	{cmd: "project", typ: "project", parents: []string{"organization"}, base: routeOrg + "project", create: []string{"--name", "n"}},
	{cmd: "project-access", typ: "project_access", parents: []string{"organization", "project"}, base: routeOrg + "project/{2}/projectAccess", create: []string{"--name", "n"}},
	{cmd: "provider", typ: "provider", parents: []string{"organization"}, base: routeOrg + "provider", create: []string{"--name", "n"}},
	{cmd: "provider-version", typ: "version", parents: []string{"organization", "provider"}, base: routeOrg + "provider/{2}/version", create: []string{"--version-number", "1.0.0"}},
	{cmd: "ssh", typ: "ssh", parents: []string{"organization"}, base: routeOrg + "ssh", create: []string{"--name", "n", "--private-key", "k", "--ssh-type", "rsa"}},
	{cmd: "step", typ: "step", parents: []string{"organization", "job"}, base: routeOrg + "job/{2}/step", create: []string{"--name", "n"}},
	{cmd: "tag", typ: "tag", parents: []string{"organization"}, base: routeOrg + "tag", create: []string{"--name", "n"}},
	{cmd: "team", typ: "team", parents: []string{"organization"}, base: routeOrg + "team", create: []string{"--name", "n"}},
	{cmd: "template", typ: "template", parents: []string{"organization"}, base: routeOrg + "template", create: []string{"--name", "n", "--content", "c"}},
	{cmd: "variable", typ: "variable", parents: []string{"organization", "workspace"}, base: routeOrgWs + "variable", create: []string{"--category", "TERRAFORM", "--key", "k", "--value", "v"}},
	{
		cmd: "vcs", typ: "vcs", parents: []string{"organization"}, base: routeOrg + "vcs",
		create: []string{
			"--name", "n", "--description", "d", "--vcs-type", "GITHUB", "--connection-type", "OAUTH",
			"--client-id", "i", "--client-secret", "s", "--endpoint", "http://e", "--vcs-api-url", "http://a",
		},
	},
	{cmd: "webhook", typ: "webhook", parents: []string{"organization", "workspace"}, base: routeOrgWs + "webhook"},
	{cmd: "webhook-event", typ: "webhook_event", parents: []string{"organization", "workspace", "webhook"}, base: routeOrgWs + "webhook/{3}/events"},
	{cmd: "workspace", typ: "workspace", parents: []string{"organization"}, base: routeOrg + "workspace", create: []string{"--name", "n"}},
	{cmd: "workspace-access", typ: "access", parents: []string{"organization", "workspace"}, base: routeOrgWs + "access", create: []string{"--name", "n"}},
	// The schedule API is keyed by workspace alone; the org flag is not part of the path.
	{
		cmd: "workspace-schedule", typ: "schedule", parents: []string{"organization", "workspace"}, base: "workspace/{2}/schedule",
		create: []string{"--schedule", "0 0 * * *", "--template-id", "t"},
	},
	{cmd: "workspace-tag", typ: "workspacetag", parents: []string{"organization", "workspace"}, base: routeOrgWs + "workspaceTag", create: []string{"--tag-id", "t"}},
}

// routeNotChecked lists resource-shaped commands this table deliberately skips.
var routeNotChecked = map[string]string{
	"team-token": "hand-written command on /access-token/v1/teams, covered by team_token_test.go",
}

func (c routeCase) verbList() []string {
	if c.verbs != nil {
		return c.verbs
	}
	return routeAllVerbs
}

func routeServer(t *testing.T, typ string) (reqs func() []string) {
	t.Helper()
	var mu sync.Mutex
	var got []string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = append(got, r.Method+" "+r.URL.Path)
		mu.Unlock()
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		case r.Method == http.MethodGet && !strings.HasSuffix(r.URL.Path, "/"+routeID):
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_, _ = fmt.Fprintf(w, `{"data":{"type":%q,"id":%q,"attributes":{}}}`, typ, routeID)
	}))
	t.Cleanup(ts.Close)
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(got)
	}
}

func TestResourceRoutes(t *testing.T) {
	for _, c := range routeCases {
		args := []string{}
		fill := strings.NewReplacer("{1}", routeParents[0], "{2}", routeParents[1], "{3}", routeParents[2]).Replace
		for i, name := range c.parents {
			args = append(args, "--"+name, routeParents[i])
		}
		args = append(args, c.extra...)
		base := routeAPI + fill(c.base)
		item := base
		if c.itemBase != "" {
			item = routeAPI + fill(c.itemBase)
		}
		item += "/" + routeID

		for _, verb := range c.verbList() {
			var verbArgs []string
			var want string
			switch verb {
			case "list":
				want = "GET " + base
			case "get":
				verbArgs, want = []string{"--id", routeID}, "GET "+item
			case "create":
				verbArgs, want = c.create, "POST "+base
			case "update":
				verbArgs, want = []string{"--id", routeID}, "PATCH "+item
			case "delete":
				verbArgs, want = []string{"--id", routeID}, "DELETE "+item
			}
			name := c.name
			if name == "" {
				name = c.cmd
			}
			t.Run(name+"/"+verb, func(t *testing.T) {
				resetGlobalFlags()
				reqs := routeServer(t, c.typ)
				full := slices.Concat([]string{c.cmd, verb}, args, verbArgs)
				if out, err := executeCommand(full...); err != nil {
					t.Fatalf("%v: %v (out=%q)", full, err, out)
				}
				if got := reqs(); !slices.Equal(got, []string{want}) {
					t.Errorf("%v\n got requests %q\nwant requests %q", full, got, []string{want})
				}
			})
		}
	}
}

// TestResourceRoutesCoverAllResources fails when a registered resource command
// is missing from routeCases (or lists different verbs), so new resources
// cannot skip route coverage.
func TestResourceRoutesCoverAllResources(t *testing.T) {
	covered := map[string][]string{}
	for _, c := range routeCases {
		covered[c.cmd] = append(covered[c.cmd], c.verbList()...)
	}
	for _, cmd := range rootCmd.Commands() {
		var verbs []string
		for _, sub := range cmd.Commands() {
			if slices.Contains(routeAllVerbs, sub.Name()) {
				verbs = append(verbs, sub.Name())
			}
		}
		if len(verbs) == 0 {
			continue
		}
		if _, skip := routeNotChecked[cmd.Name()]; skip {
			continue
		}
		have, ok := covered[cmd.Name()]
		if !ok {
			t.Errorf("resource command %q (verbs %v) has no entry in routeCases", cmd.Name(), verbs)
			continue
		}
		for _, v := range verbs {
			if !slices.Contains(have, v) {
				t.Errorf("resource command %q verb %q has no route case", cmd.Name(), v)
			}
		}
	}
	for cmd := range covered {
		if findSubCmd(cmd) == nil {
			t.Errorf("routeCases names unknown command %q", cmd)
		}
	}
}
