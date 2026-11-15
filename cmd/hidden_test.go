package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func subCmd(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	c := findSubCmd(path[0])
	for _, name := range path[1:] {
		if c == nil {
			break
		}
		var next *cobra.Command
		for _, sub := range c.Commands() {
			if sub.Name() == name {
				next = sub
			}
		}
		c = next
	}
	if c == nil {
		t.Fatalf("command %v not found", path)
	}
	return c
}

// AC-13: commands and flags only executors may use stay out of --help, but still run.
func TestHelpHidesExecutorOnly(t *testing.T) {
	for _, path := range [][]string{
		{"github-app-token", "create"},
		{"github-app-token", "update"},
		{"github-app-token", "delete"},
		{"step", "create"},
		{"step", "delete"},
		{"history", "create"},
		{"history", "update"},
		{"history", "delete"},
		{"job", "delete"},
	} {
		if !subCmd(t, path...).Hidden {
			t.Errorf("%s should be hidden", strings.Join(path, " "))
		}
	}
	for _, f := range []struct {
		path []string
		flag string
	}{
		{[]string{"job", "create"}, "job-output"},
		{[]string{"job", "update"}, "job-output"},
		{[]string{"job", "create"}, "override-source"},
		{[]string{"job", "update"}, "override-source"},
		{[]string{"step", "update"}, "output"},
		{[]string{"step", "update"}, "status"},
	} {
		flag := subCmd(t, f.path...).Flags().Lookup(f.flag)
		if flag == nil || !flag.Hidden {
			t.Errorf("%s --%s should be a hidden flag (got %v)", strings.Join(f.path, " "), f.flag, flag)
		}
	}
	// Still visible: the user-facing parts of the same resources.
	if subCmd(t, "job", "create").Hidden || subCmd(t, "history", "list").Hidden {
		t.Error("job create and history list must stay visible")
	}
	if flag := subCmd(t, "job", "update").Flags().Lookup("status"); flag == nil || flag.Hidden {
		t.Error("job update --status must stay visible")
	}
}
