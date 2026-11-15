package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdLogoutOutputMessages(t *testing.T) {
	t.Run("stored token", func(t *testing.T) {
		resetGlobalFlags()
		cfg := writeTempFile(t, "cfg.yaml", "token: secret\n")

		out, err := executeCommand("logout", "--config", cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Successfully logged out") {
			t.Errorf("expected success message, got %q", out)
		}
	})

	t.Run("no config", func(t *testing.T) {
		resetGlobalFlags()
		out, err := executeCommand("logout", "--config", filepath.Join(t.TempDir(), "missing.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "No stored credentials") {
			t.Errorf("expected no-credentials message, got %q", out)
		}
	})
}

func TestCmdLogoutNoHome(t *testing.T) {
	resetGlobalFlags()
	t.Cleanup(resetGlobalFlags)
	cfgFile = ""
	t.Setenv("HOME", "")

	// Run RunE directly: through Execute, initConfig would exit first.
	if err := logoutCmd.RunE(logoutCmd, nil); err == nil {
		t.Fatal("expected error when home directory is unknown")
	}
}

func TestCmdLogoutSaveFailures(t *testing.T) {
	tests := []struct {
		name    string
		cfg     func(t *testing.T) string
		wantErr string
	}{
		{
			name:    "malformed yaml",
			cfg:     func(t *testing.T) string { return writeTempFile(t, "cfg.yaml", "a: [unclosed\n") },
			wantErr: "parsing",
		},
		{
			name:    "non-yaml config",
			cfg:     func(t *testing.T) string { return writeTempFile(t, "cfg.json", `{"token":"x"}`) },
			wantErr: "must be YAML, not json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			cfg := tt.cfg(t)
			before, _ := os.ReadFile(cfg)

			out, err := executeCommand("logout", "--config", cfg)
			if err == nil || !strings.Contains(err.Error(), "saving configuration") || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected saving configuration error containing %q, got %v", tt.wantErr, err)
			}
			if strings.Contains(out, "Successfully logged out") {
				t.Errorf("must not report success, got %q", out)
			}
			after, _ := os.ReadFile(cfg)
			if string(before) != string(after) {
				t.Errorf("config was modified: %q -> %q", before, after)
			}
		})
	}
}
