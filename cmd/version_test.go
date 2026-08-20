package cmd

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	resetGlobalFlags()

	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error running --version: %v", err)
	}

	expectedFields := []string{
		"Version:",
		"Git Commit:",
		"Built At:",
		"Go Version:",
		"OS/Arch:",
		runtime.Version(),
		runtime.GOOS + "/" + runtime.GOARCH,
	}

	for _, field := range expectedFields {
		if !strings.Contains(out, field) {
			t.Errorf("expected output to contain %q, got:\n%s", field, out)
		}
	}
}

func TestVersionShortFlag(t *testing.T) {
	resetGlobalFlags()

	out, err := executeCommand("-v")
	if err != nil {
		t.Fatalf("unexpected error running -v: %v", err)
	}

	if !strings.Contains(out, "Version:") {
		t.Errorf("expected output to contain 'Version:', got:\n%s", out)
	}
}

func TestGetBuildInfo(t *testing.T) {
	oldVersion := Version
	oldCommit := Commit
	oldDate := Date
	defer func() {
		Version = oldVersion
		Commit = oldCommit
		Date = oldDate
	}()

	Version = "v1.2.3"
	Commit = "abcdef1"
	Date = "2026-08-20T16:00:00Z"

	info := GetBuildInfo()
	if info.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", info.Version)
	}
	if info.Commit != "abcdef1" {
		t.Errorf("expected commit abcdef1, got %s", info.Commit)
	}
	if info.Date != "2026-08-20T16:00:00Z" {
		t.Errorf("expected date 2026-08-20T16:00:00Z, got %s", info.Date)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %s, got %s", runtime.Version(), info.GoVersion)
	}
	if info.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Errorf("expected Platform %s/%s, got %s", runtime.GOOS, runtime.GOARCH, info.Platform)
	}

	formatted := FormatVersion()
	if !strings.Contains(formatted, "Version:    v1.2.3") {
		t.Errorf("unexpected formatted output:\n%s", formatted)
	}
	if !strings.Contains(formatted, "Git Commit: abcdef1") {
		t.Errorf("unexpected formatted output:\n%s", formatted)
	}
}
