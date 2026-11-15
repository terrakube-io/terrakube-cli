package cmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

func TestExecuteExitsOnError(t *testing.T) {
	code, out := runHelperProcess(t, "execute-not-logged-in")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (out %q)", code, out)
	}
	if !strings.Contains(out, "Error: not logged in") {
		t.Errorf("expected error message on output, got %q", out)
	}
}

func TestInitConfigExitsWithoutHome(t *testing.T) {
	code, out := runHelperProcess(t, "execute-no-home")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (out %q)", code, out)
	}
	if !strings.Contains(out, "HOME") {
		t.Errorf("expected HOME error, got %q", out)
	}
}

func TestRenderOutputExitsOnUnsupportedFormat(t *testing.T) {
	code, out := runHelperProcess(t, "render-error")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (out %q)", code, out)
	}
	if !strings.Contains(out, "Failed to render output") || !strings.Contains(out, "unsupported output format") {
		t.Errorf("unexpected output %q", out)
	}
}

func TestExecuteSuccess(t *testing.T) {
	code, out := runHelperProcess(t, "execute-version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (out %q)", code, out)
	}
	if !strings.Contains(out, "Version:") {
		t.Errorf("expected version output, got %q", out)
	}
}

func TestHelperProcessRejectsUnknownMode(t *testing.T) {
	code, out := runHelperProcess(t, "bogus")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 (out %q)", code, out)
	}
	if !strings.Contains(out, `unknown `+helperEnv+` mode "bogus"`) {
		t.Errorf("unexpected output %q", out)
	}
}

func TestConfigPath(t *testing.T) {
	resetGlobalFlags()
	t.Cleanup(resetGlobalFlags)

	cfgFile = "/explicit/path.yaml"
	if got, err := configPath(); err != nil || got != "/explicit/path.yaml" {
		t.Errorf("configPath() = %q, %v; want explicit path", got, err)
	}

	cfgFile = ""
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got, err := configPath(); err != nil || got != filepath.Join(home, ".terrakube-cli.yaml") {
		t.Errorf("configPath() = %q, %v; want default under HOME", got, err)
	}

	t.Setenv("HOME", "")
	if got, err := configPath(); err == nil {
		t.Errorf("expected error without HOME, got path %q", got)
	}
	resetGlobalFlags()
}

func TestConfigType(t *testing.T) {
	tests := map[string]string{
		"c.yaml":  "yaml",
		"c.yml":   "yml",
		"c.json":  "json",
		"c.toml":  "toml",
		"c":       "yaml",
		"c.cfg":   "yaml",
		"dir/c.x": "yaml",
	}
	for path, want := range tests {
		if got := configType(path); got != want {
			t.Errorf("configType(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestApplyEnvInvalidValueFailsBeforeRequest(t *testing.T) {
	resetGlobalFlags()
	t.Setenv("TERRAKUBE_HIDE_NULLS", "maybe")
	var requests atomic.Int32
	ts := setupTestServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer ts.Close()

	_, err := executeCommand("organization", "list")
	if err == nil || !strings.Contains(err.Error(), "invalid TERRAKUBE_HIDE_NULLS") {
		t.Fatalf("expected invalid env error, got %v", err)
	}
	if requests.Load() != 0 {
		t.Errorf("expected no requests, got %d", requests.Load())
	}
}

func TestApplyEnvOutputAndFlagPrecedence(t *testing.T) {
	orgServer := func() func() {
		ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{{ID: "o1", Name: "acme"}})
		}))
		return ts.Close
	}

	t.Run("env selects format", func(t *testing.T) {
		resetGlobalFlags()
		t.Setenv("TERRAKUBE_OUTPUT", "yaml")
		defer orgServer()()

		out, err := executeCommand("organization", "list")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "name: acme") {
			t.Errorf("expected yaml output, got %q", out)
		}
	})

	t.Run("explicit flag beats env", func(t *testing.T) {
		resetGlobalFlags()
		t.Setenv("TERRAKUBE_OUTPUT", "yaml")
		defer orgServer()()

		out, err := executeCommand("organization", "list", "--output", "json")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out, "name: acme") || !strings.Contains(out, `"acme"`) {
			t.Errorf("expected json output, got %q", out)
		}
	})
}

func TestInitConfigVerbosePrintsConfigFile(t *testing.T) {
	resetGlobalFlags()
	cfg, _ := writeLoggedInConfig(t, "cfg.yaml", "tok")

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()
	_, err := executeCommand("organization", "list", "--config", cfg, "--verbose")
	_ = w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Using config file: "+cfg) {
		t.Errorf("expected verbose config message, got %q", buf.String())
	}
}

func TestInitConfigQuietByDefault(t *testing.T) {
	resetGlobalFlags()
	cfg, _ := writeLoggedInConfig(t, "cfg.yaml", "tok")

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()
	_, err := executeCommand("organization", "list", "--config", cfg)
	_ = w.Close()
	os.Stderr = old
	got, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Contains(got, []byte("Using config file")) {
		t.Errorf("unexpected verbose message without --verbose: %q", got)
	}
}
