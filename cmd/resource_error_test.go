package cmd

import (
	"net/http"
	"strings"
	"testing"
)

// TestResourceFrameworkSingleErrorOutput verifies that when parent flags are missing,
// the error message is printed exactly once, not duplicated.
// Regression test for tkc-0ab.
func TestResourceFrameworkSingleErrorOutput(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errorSubstr string
	}{
		{
			name:        "3-parent resource (address) missing organization",
			args:        []string{"address", "list", "--job-id", "test123"},
			errorSubstr: "--organization is required",
		},
		{
			name:        "2-parent resource (workspace-access) missing organization",
			args:        []string{"workspace-access", "list", "--workspace-id", "test456"},
			errorSubstr: "--organization is required",
		},
		{
			name:        "2-parent resource (collection-item) missing organization",
			args:        []string{"collection-item", "list", "--collection-id", "test789"},
			errorSubstr: "--organization is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()

			// Set up test server (won't be called because of early error)
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			ts := setupTestServer(handler)
			defer ts.Close()

			out, err := executeCommand(tt.args...)
			if err == nil {
				t.Fatal("expected error for missing parent flags, got nil")
			}

			// Check that error is returned
			if !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Errorf("expected error to contain %q, got: %v", tt.errorSubstr, err)
			}

			// Check that error is NOT duplicated in captured output
			// (With SilenceErrors: true, Cobra doesn't print to SetErr,
			// errors are only returned, so output should be empty)
			if strings.Contains(out, "Error:") {
				t.Errorf("error should not be printed to output with SilenceErrors: true, got output: %s", out)
			}

			// Count occurrences of the error substring in output
			// Should be 0 because of SilenceErrors: true
			count := strings.Count(out, tt.errorSubstr)
			if count > 0 {
				t.Errorf("error substring appears %d time(s) in output, expected 0 with SilenceErrors: true", count)
			}
		})
	}
}
