package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// helperEnv selects a helper mode: the test binary re-executes itself with
// this variable set to observe code paths that end in os.Exit.
const helperEnv = "TKCLI_TEST_HELPER"

const noHomeHelperMode = "execute-no-home"

// TestMain isolates the package from the developer's HOME and TERRAKUBE_* env
// and dispatches helper-process modes before any per-run state is created.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, envPrefix+"_") {
			_ = os.Unsetenv(kv[:strings.Index(kv, "=")])
		}
	}
	if mode := os.Getenv(helperEnv); mode != "" {
		runHelperMode(mode) // never returns
	}

	home, err := os.MkdirTemp("", "terrakube-cli-test-home")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("HOME", home)
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

// runHelperMode runs in a re-executed test binary and always exits.
func runHelperMode(mode string) {
	switch mode {
	case "execute-not-logged-in":
		cfgFile = os.DevNull
		rootCmd.SetArgs([]string{"organization", "list"})
		Execute()
	case noHomeHelperMode:
		_ = os.Unsetenv("HOME")
		cfgFile = ""
		rootCmd.SetArgs([]string{"organization", "list"})
		Execute()
	case "execute-version":
		cfgFile = os.DevNull
		rootCmd.SetArgs([]string{"--version"})
		Execute()
	case "render-error":
		renderOutput(map[string]string{"a": "b"}, "bogus")
	default:
		fmt.Fprintf(os.Stderr, "unknown %s mode %q\n", helperEnv, mode)
		os.Exit(2)
	}
	os.Exit(0)
}

// runHelperProcess re-executes the test binary in the given helper mode and
// returns its exit code and combined output.
func runHelperProcess(t *testing.T, mode string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), helperEnv+"="+mode)
	if mode != noHomeHelperMode {
		cmd.Env = append(cmd.Env, "HOME="+t.TempDir())
	}
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, string(out)
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), string(out)
	default:
		t.Fatalf("running helper: %v", err)
		return -1, ""
	}
}
