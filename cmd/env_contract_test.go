package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"terrakube/internal/resource"
)

const (
	envOrgA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	envOrgB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	envWsC  = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
)

// envVars are every TERRAKUBE_* variable the CLI reads; tests blank them so a
// developer's shell cannot leak in.
var envVars = []string{
	"TERRAKUBE_API_URL", "TERRAKUBE_TOKEN", "TERRAKUBE_PAT",
	"TERRAKUBE_ORGANIZATION", "TERRAKUBE_ORGANIZATION_ID",
	"TERRAKUBE_WORKSPACE", "TERRAKUBE_WORKSPACE_ID",
	"TERRAKUBE_OUTPUT", "TERRAKUBE_HIDE_NULLS",
}

// envSetup resets flags, blanks TERRAKUBE_* variables and drops viper's
// explicit overrides (viper.Set shadows environment variables, and
// resetGlobalFlags sets api_url/token to ""). initConfig re-initialises viper
// on every Execute.
func envSetup(t *testing.T) {
	t.Helper()
	resetGlobalFlags()
	viper.Reset()
	t.Cleanup(viper.Reset)
	for _, v := range envVars {
		t.Setenv(v, "")
	}
}

// envServer records "METHOD path" of every request. respond may write a body;
// when it returns false (or is nil) an empty JSON:API list is served.
func envServer(t *testing.T, respond func(w http.ResponseWriter, r *http.Request) bool) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var got []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = append(got, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if respond != nil && respond(w, r) {
			return
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"data":{"type":"notification_configuration","id":%q,"attributes":{}}}`, routeID)
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(ts.Close)
	return ts, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(got)
	}
}

// envAPI points the CLI at ts through the environment only.
func envAPI(t *testing.T, ts *httptest.Server) {
	t.Helper()
	t.Setenv("TERRAKUBE_API_URL", ts.URL)
	t.Setenv("TERRAKUBE_TOKEN", "env-token")
}

func envWriteConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func envExpect(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("requests = %q, want %q", got, want)
	}
}

// ----- 1. parent scope from the environment -----

func TestEnvParentScopeOptionalWorkspace(t *testing.T) {
	const orgOnly = "GET /api/v1/organization/" + envOrgA + "/notificationConfiguration"
	const withWs = "GET /api/v1/organization/" + envOrgA + "/workspace/" + envWsC + "/notificationConfiguration"
	create := []string{"notification-configuration", "create", "--name", "n", "--channel-type", "SLACK", "--destination-url", "http://d"}

	cases := []struct {
		name string
		env  map[string]string
		verb []string
		want string
	}{
		{"list/WORKSPACE", map[string]string{"TERRAKUBE_WORKSPACE": envWsC}, []string{"notification-configuration", "list"}, withWs},
		{"list/WORKSPACE_ID", map[string]string{"TERRAKUBE_WORKSPACE_ID": envWsC}, []string{"notification-configuration", "list"}, withWs},
		{"list/unset", nil, []string{"notification-configuration", "list"}, orgOnly},
		{"create/WORKSPACE", map[string]string{"TERRAKUBE_WORKSPACE": envWsC}, create, "POST /api/v1/organization/" + envOrgA + "/workspace/" + envWsC + "/notificationConfiguration"},
		{"create/WORKSPACE_ID", map[string]string{"TERRAKUBE_WORKSPACE_ID": envWsC}, create, "POST /api/v1/organization/" + envOrgA + "/workspace/" + envWsC + "/notificationConfiguration"},
		{"create/unset", nil, create, "POST /api/v1/organization/" + envOrgA + "/notificationConfiguration"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			envSetup(t)
			ts, reqs := envServer(t, nil)
			envAPI(t, ts)
			t.Setenv("TERRAKUBE_ORGANIZATION_ID", envOrgA)
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			if out, err := executeCommand(c.verb...); err != nil {
				t.Fatalf("%v: %v (out=%q)", c.verb, err, out)
			}
			envExpect(t, reqs(), c.want)
		})
	}
}

func TestEnvParentScopeOrganizationName(t *testing.T) {
	envSetup(t)
	ts, reqs := envServer(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/organization" {
			_, _ = fmt.Fprintf(w, `{"data":[{"type":"organization","id":%q,"attributes":{"name":"acme"}}]}`, envOrgA)
			return true
		}
		return false
	})
	envAPI(t, ts)
	t.Setenv("TERRAKUBE_ORGANIZATION", "acme")

	if out, err := executeCommand("workspace", "list"); err != nil {
		t.Fatalf("workspace list: %v (out=%q)", err, out)
	}
	envExpect(t, reqs(),
		"GET /api/v1/organization",
		"GET /api/v1/organization/"+envOrgA+"/workspace",
	)
}

func TestEnvParentScopeExplicitFlagBeatsEnv(t *testing.T) {
	for _, flag := range [][]string{{"--organization-id", envOrgB}, {"-o", envOrgB}, {"--organization", envOrgB}, {"--org", envOrgB}} {
		t.Run(flag[0], func(t *testing.T) {
			envSetup(t)
			ts, reqs := envServer(t, nil)
			envAPI(t, ts)
			t.Setenv("TERRAKUBE_ORGANIZATION_ID", envOrgA)

			if out, err := executeCommand(append([]string{"workspace", "list"}, flag...)...); err != nil {
				t.Fatalf("%v (out=%q)", err, out)
			}
			envExpect(t, reqs(), "GET /api/v1/organization/"+envOrgB+"/workspace")
		})
	}
}

// ----- 2. API URL and token from the environment -----

func TestEnvAPIURLAndTokenWithoutConfig(t *testing.T) {
	envSetup(t)
	cfgFile = envWriteConfig(t, "{}\n")
	var auth string
	ts, reqs := envServer(t, func(_ http.ResponseWriter, r *http.Request) bool {
		auth = r.Header.Get("Authorization")
		return false
	})
	t.Setenv("TERRAKUBE_API_URL", ts.URL)
	t.Setenv("TERRAKUBE_TOKEN", "env-token")

	if out, err := executeCommand("organization", "list"); err != nil {
		t.Fatalf("organization list: %v (out=%q)", err, out)
	}
	envExpect(t, reqs(), "GET /api/v1/organization")
	if auth != "Bearer env-token" {
		t.Fatalf("Authorization = %q, want Bearer env-token", auth)
	}
}

func TestEnvPATFallbackToken(t *testing.T) {
	envSetup(t)
	cfgFile = envWriteConfig(t, "{}\n")
	var auth string
	ts, reqs := envServer(t, func(_ http.ResponseWriter, r *http.Request) bool {
		auth = r.Header.Get("Authorization")
		return false
	})
	t.Setenv("TERRAKUBE_API_URL", ts.URL)
	t.Setenv("TERRAKUBE_PAT", "env-pat")

	if out, err := executeCommand("organization", "list"); err != nil {
		t.Fatalf("organization list: %v (out=%q)", err, out)
	}
	envExpect(t, reqs(), "GET /api/v1/organization")
	if auth != "Bearer env-pat" {
		t.Fatalf("Authorization = %q, want Bearer env-pat", auth)
	}
}

func TestEnvBeatsConfigFile(t *testing.T) {
	envSetup(t)
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("request reached the config file's server; env must win")
	}))
	t.Cleanup(dead.Close)
	cfgFile = envWriteConfig(t, "api_url: "+dead.URL+"\ntoken: config-token\n")

	var auth string
	ts, reqs := envServer(t, func(_ http.ResponseWriter, r *http.Request) bool {
		auth = r.Header.Get("Authorization")
		return false
	})
	t.Setenv("TERRAKUBE_API_URL", ts.URL)
	t.Setenv("TERRAKUBE_TOKEN", "env-token")

	if out, err := executeCommand("organization", "list"); err != nil {
		t.Fatalf("organization list: %v (out=%q)", err, out)
	}
	envExpect(t, reqs(), "GET /api/v1/organization")
	if auth != "Bearer env-token" {
		t.Fatalf("Authorization = %q, want Bearer env-token (env beats config)", auth)
	}
}

// ----- 3. parent flag aliases and shorthands -----

// envKnownAliases are the unified-flag aliases the README documents; the ID
// aliases are discovered from the flags' env annotation.
var envKnownAliases = map[string]string{"organization": "org", "workspace": "ws"}

func TestEnvParentFlagForms(t *testing.T) {
	tested := 0
	for _, c := range rootCmd.Commands() {
		if !slices.ContainsFunc(c.Commands(), func(s *cobra.Command) bool { return s.Name() == "list" }) {
			continue
		}
		listCmd, _, err := rootCmd.Find([]string{c.Name(), "list"})
		if err != nil {
			t.Fatal(err)
		}
		type parent struct {
			flag, short string
			forms       []string // flag forms (without value) that must select this parent
			canonical   string
		}
		var parents []parent
		for _, name := range envFlagNames(listCmd) {
			f := listCmd.Flags().Lookup(name)
			envNames := f.Annotations[resource.EnvAnnotation]
			if len(envNames) == 0 {
				continue
			}
			p := parent{flag: f.Name, short: f.Shorthand, canonical: "--" + f.Name}
			p.forms = append(p.forms, "--"+f.Name)
			for _, id := range envNames[1:] {
				p.forms = append(p.forms, "--"+id)
				p.canonical = "--" + id
			}
			if a, ok := envKnownAliases[f.Name]; ok {
				p.forms = append(p.forms, "--"+a)
			}
			if p.short != "" {
				p.forms = append(p.forms, "-"+p.short)
			}
			parents = append(parents, p)
		}
		if len(parents) == 0 {
			continue
		}
		tested++

		build := func(form string, at int) []string {
			args := []string{c.Name(), "list"}
			for i, p := range parents {
				flag := p.canonical
				if i == at {
					flag = form
				}
				args = append(args, flag, routeParents[i])
			}
			return args
		}
		run := func(t *testing.T, args []string) []string {
			resetGlobalFlags()
			reqs := routeServer(t, "x")
			if out, err := executeCommand(args...); err != nil {
				t.Fatalf("%v: %v (out=%q)", args, err, out)
			}
			got := reqs()
			if len(got) != 1 {
				t.Fatalf("%v: requests = %q, want exactly one", args, got)
			}
			return got
		}

		t.Run(c.Name(), func(t *testing.T) {
			want := run(t, build("", -1))
			for i, p := range parents {
				for _, form := range p.forms {
					got := run(t, build(form, i))
					if !slices.Equal(got, want) {
						t.Errorf("%s as %s: requests = %q, want %q", p.flag, form, got, want)
					}
				}
			}
		})
	}
	if tested < 30 {
		t.Fatalf("only %d resource commands with parent flags found, command tree walk is broken", tested)
	}
}

func envFlagNames(c *cobra.Command) []string {
	var names []string
	c.LocalFlags().VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// ----- 4. api_url path prefix -----

func TestEnvLoginKeepsPathPrefix(t *testing.T) {
	for _, suffix := range []string{"/terrakube/api/v1", "/terrakube/api/v1/", "/terrakube", "/terrakube/"} {
		t.Run(suffix, func(t *testing.T) {
			envSetup(t)
			ts, reqs := envServer(t, nil)
			cfgFile = envWriteConfig(t, "{}\n")

			if out, err := executeCommand("login", "--api-url", ts.URL+suffix, "--pat", "x"); err != nil {
				t.Fatalf("login: %v (out=%q)", err, out)
			}
			envExpect(t, reqs(), "GET /terrakube/api/v1/organization")

			data, err := os.ReadFile(cfgFile)
			if err != nil {
				t.Fatal(err)
			}
			var saved map[string]any
			if err := yaml.Unmarshal(data, &saved); err != nil {
				t.Fatal(err)
			}
			if saved["api_url"] != ts.URL+"/terrakube" {
				t.Fatalf("saved api_url = %v, want %s/terrakube", saved["api_url"], ts.URL)
			}
		})
	}
}

func TestEnvAPIURLPathPrefixIsKept(t *testing.T) {
	envSetup(t)
	ts, reqs := envServer(t, nil)
	cfgFile = envWriteConfig(t, "api_url: "+ts.URL+"/terrakube\ntoken: t\n")

	if out, err := executeCommand("organization", "list"); err != nil {
		t.Fatalf("organization list: %v (out=%q)", err, out)
	}
	envExpect(t, reqs(), "GET /terrakube/api/v1/organization")
}

// ----- 5. config file only provides api_url and token -----

func TestEnvConfigFileDoesNotFillFlags(t *testing.T) {
	envSetup(t)
	ts, reqs := envServer(t, nil)
	cfgFile = envWriteConfig(t, "api_url: "+ts.URL+"\ntoken: t\norganization-id: "+envOrgA+
		"\norganization: "+envOrgA+"\noutput: yaml\nname: x\nhide-nulls: false\n")

	out, err := executeCommand("workspace", "list")
	if err == nil {
		t.Fatalf("workspace list succeeded, want --organization is required (out=%q)", out)
	}
	if !strings.Contains(err.Error(), "--organization is required") {
		t.Fatalf("error = %v, want --organization is required", err)
	}
	envExpect(t, reqs())
}

func TestEnvConfigFileOutputIgnored(t *testing.T) {
	envSetup(t)
	ts, _ := envServer(t, func(w http.ResponseWriter, _ *http.Request) bool {
		_, _ = fmt.Fprintf(w, `{"data":[{"type":"organization","id":%q,"attributes":{"name":"acme"}}]}`, envOrgA)
		return true
	})
	cfgFile = envWriteConfig(t, "api_url: "+ts.URL+"\ntoken: t\noutput: yaml\n")
	out, err := executeCommand("organization", "list")
	if err != nil {
		t.Fatalf("organization list: %v (out=%q)", err, out)
	}
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil || len(parsed) != 1 {
		t.Fatalf("output is not a JSON array of one org (config output: yaml must be ignored): %q (%v)", out, err)
	}
}

// ----- 6. hide-nulls -----

func TestEnvHideNulls(t *testing.T) {
	cases := []struct {
		name string
		env  string
		args []string
		keep bool
	}{
		{"default", "", nil, false},
		{"flag-false", "", []string{"--hide-nulls=false"}, true},
		{"env-false", "false", nil, true},
		{"env-true", "true", nil, false},
		{"flag-true-beats-env-false", "false", []string{"--hide-nulls=true"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			envSetup(t)
			ts, _ := envServer(t, func(w http.ResponseWriter, _ *http.Request) bool {
				_, _ = fmt.Fprintf(w, `{"data":{"type":"organization","id":%q,"attributes":{"name":"acme","description":null,"icon":null}}}`, envOrgA)
				return true
			})
			envAPI(t, ts)
			if c.env != "" {
				t.Setenv("TERRAKUBE_HIDE_NULLS", c.env)
			}
			out, err := executeCommand(append([]string{"organization", "get", "--id", envOrgA}, c.args...)...)
			if err != nil {
				t.Fatalf("organization get: %v (out=%q)", err, out)
			}
			var got map[string]any
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("output not JSON: %q", out)
			}
			attrs, _ := got["attributes"].(map[string]any)
			if attrs["name"] != "acme" {
				t.Fatalf("name missing from %q", out)
			}
			v, present := attrs["description"]
			if c.keep && (!present || v != nil) {
				t.Errorf("description should be kept as null, got %v (present=%v) in %q", v, present, out)
			}
			if !c.keep && present {
				t.Errorf("null description should be stripped, got %v in %q", v, out)
			}
		})
	}
}
