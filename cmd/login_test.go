package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/jsonapi"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"
	"gopkg.in/yaml.v3"
)

// TestMain isolates the package from the developer's HOME and TERRAKUBE_* env.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "terrakube-cli-test-home")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("HOME", home)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, envPrefix+"_") {
			_ = os.Unsetenv(kv[:strings.Index(kv, "=")])
		}
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

func orgListServer(t *testing.T) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{})
	}))
	t.Cleanup(ts.Close)
	return ts
}

func readConfig(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	settings := map[string]any{}
	if err := yaml.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parsing config: %v", err)
	}
	return settings
}

func assertMode0600(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config mode = %o, want 600", info.Mode().Perm())
	}
}

func TestCmdLoginWritesOnlyCredentials(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("custom: keep\ntoken: old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := executeCommand("login", "--config", cfg, "--api-url", ts.URL+"/api/v1/", "--pat", "my-pat"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readConfig(t, cfg)
	want := map[string]any{"api_url": ts.URL, "token": "my-pat", "custom": "keep"}
	if len(got) != len(want) {
		t.Errorf("config = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("config[%q] = %v, want %v", k, got[k], v)
		}
	}
	assertMode0600(t, cfg)
}

func TestCmdLoginUnreachableServer(t *testing.T) {
	resetGlobalFlags()
	ts := httptest.NewServer(http.NotFoundHandler())
	ts.Close()
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	_, err := executeCommand("login", "--config", cfg, "--api-url", ts.URL, "--pat", "my-pat")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if _, statErr := os.Stat(cfg); !os.IsNotExist(statErr) {
		t.Errorf("config must not be written on failed login, stat err: %v", statErr)
	}
}

func TestCmdLoginFromEnv(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	t.Setenv("TERRAKUBE_API_URL", ts.URL)
	t.Setenv("TERRAKUBE_PAT", "env-pat")

	if _, err := executeCommand("login", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readConfig(t, cfg)
	if got["api_url"] != ts.URL || got["token"] != "env-pat" {
		t.Errorf("config = %v", got)
	}
}

func TestCmdLoginHelpExample(t *testing.T) {
	if strings.Contains(loginCmd.Example, "%!") {
		t.Errorf("login example has formatting error:\n%s", loginCmd.Example)
	}
}

func TestCmdLogoutRemovesToken(t *testing.T) {
	resetGlobalFlags()
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("api_url: http://x\ntoken: secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := executeCommand("logout", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readConfig(t, cfg)
	if _, ok := got["token"]; ok || got["api_url"] != "http://x" {
		t.Errorf("config = %v, want only api_url", got)
	}
	assertMode0600(t, cfg)
}

func TestCmdNewClientNotLoggedIn(t *testing.T) {
	resetGlobalFlags()

	_, err := executeCommand("organization", "list")
	if err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("expected not-logged-in error, got: %v", err)
	}
}

func TestCmdEnvFillsParentFlagsOnly(t *testing.T) {
	resetGlobalFlags()
	const orgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	t.Setenv("TERRAKUBE_ORGANIZATION_ID", orgID)
	t.Setenv("TERRAKUBE_NAME", "oops")

	var patchPath, patchBody string
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patchPath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			patchBody = string(b)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, &terrakube.Workspace{ID: "ws-1", Name: "orig"})
	}))
	defer ts.Close()

	if _, err := executeCommand("workspace", "update", "--id", "ws-1", "--description", "d"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(patchPath, orgID) {
		t.Errorf("PATCH path %q does not use org from TERRAKUBE_ORGANIZATION_ID", patchPath)
	}
	if strings.Contains(patchBody, "oops") {
		t.Errorf("TERRAKUBE_NAME leaked into PATCH body: %s", patchBody)
	}
}

func TestCmdInvalidOutputFailsBeforeRequest(t *testing.T) {
	resetGlobalFlags()
	var requests atomic.Int32
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
	}))
	defer ts.Close()

	_, err := executeCommand("--output", "foo", "organization", "create", "--name", "x")
	if err == nil || !strings.Contains(err.Error(), "output") {
		t.Fatalf("expected invalid output error, got: %v", err)
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("expected no HTTP requests, got %d", n)
	}
}

func TestCmdNewClientFallsBackToPATEnv(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	viper.Set("api_url", ts.URL)
	t.Setenv("TERRAKUBE_PAT", "env-pat")

	if _, err := executeCommand("organization", "list"); err != nil {
		t.Fatalf("expected TERRAKUBE_PAT to authenticate, got: %v", err)
	}
}

func TestCmdLoginRemovesLegacyPatKey(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("pat: old\ntoken: old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := executeCommand("login", "--config", cfg, "--api-url", ts.URL, "--pat", "new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readConfig(t, cfg); got["pat"] != nil || got["token"] != "new" {
		t.Errorf("config = %v, want token=new and no pat", got)
	}
}

func TestCmdLogoutRemovesLegacyPatKey(t *testing.T) {
	resetGlobalFlags()
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("api-url: http://x\npat: secret\ntoken: secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := executeCommand("logout", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readConfig(t, cfg)
	if got["pat"] != nil || got["token"] != nil || got["api-url"] != "http://x" {
		t.Errorf("config = %v, want only api-url", got)
	}
}

func TestCmdLogoutWithoutConfigCreatesNothing(t *testing.T) {
	resetGlobalFlags()
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	if _, err := executeCommand("logout", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Errorf("logout must not create a config file, stat err: %v", err)
	}
}

func TestCmdMalformedConfigReported(t *testing.T) {
	resetGlobalFlags()
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(cfg, []byte("api_url: [unclosed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand("organization", "list", "--config", cfg)
	if err == nil || !strings.Contains(err.Error(), "reading config file") {
		t.Fatalf("expected config parse error, got: %v", err)
	}

	// A missing config file is not an error: it falls through to "not logged in".
	_, err = executeCommand("organization", "list", "--config", filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Fatalf("expected not-logged-in error for missing config, got: %v", err)
	}
}

func TestCmdLoginWritesThroughSymlink(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")
	link := filepath.Join(dir, "link.yaml")
	if err := os.WriteFile(target, []byte("token: old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if _, err := executeCommand("login", "--config", link, "--api-url", ts.URL, "--pat", "new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("config symlink was replaced (err=%v)", err)
	}
	if got := readConfig(t, target); got["token"] != "new" {
		t.Errorf("target config = %v, want token=new", got)
	}
}

// writeLoggedInConfig writes a config pointing at a server that records the bearer token.
func writeLoggedInConfig(t *testing.T, name, token string) (cfg string, gotToken *string) {
	t.Helper()
	gotToken = new(string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotToken = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{})
	}))
	t.Cleanup(ts.Close)
	cfg = filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(cfg, []byte("api_url: "+ts.URL+"\ntoken: "+token+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Drop viper.Set overrides from resetGlobalFlags so the config file is used.
	viper.Reset()
	return cfg, gotToken
}

func TestCmdUnsupportedConfigExtensionReadAsYAML(t *testing.T) {
	resetGlobalFlags()
	cfg, gotToken := writeLoggedInConfig(t, "terrakube.cfg", "cfg-token")

	if _, err := executeCommand("organization", "list", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *gotToken != "cfg-token" {
		t.Errorf("token = %q, want cfg-token", *gotToken)
	}
}

func TestCmdTokenEnvBeatsConfig(t *testing.T) {
	tests := []struct {
		name, token, pat, want string
	}{
		{"token env", "env-token", "", "env-token"},
		{"pat env", "", "env-pat", "env-pat"},
		{"token beats pat", "env-token", "env-pat", "env-token"},
		{"config only", "", "", "stale"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			cfg, gotToken := writeLoggedInConfig(t, "cfg.yaml", "stale")
			t.Setenv("TERRAKUBE_TOKEN", tt.token)
			t.Setenv("TERRAKUBE_PAT", tt.pat)

			if _, err := executeCommand("organization", "list", "--config", cfg); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *gotToken != tt.want {
				t.Errorf("token = %q, want %q", *gotToken, tt.want)
			}
		})
	}
}
