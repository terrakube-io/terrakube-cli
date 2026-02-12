package resource

import (
	"context"
	"fmt"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

func TestIsUUID(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"valid UUID", "e5ad0642-f9b3-48b3-9bf4-35997febe1fb", true},
		{"valid UUID uppercase", "E5AD0642-F9B3-48B3-9BF4-35997FEBE1FB", true},
		{"too short", "abc-123", false},
		{"missing dashes", "e5ad0642f9b348b39bf435997febe1fb", true},
		{"non-hex characters", "e5ad0642-f9b3-48b3-9bf4-35997febe1gz", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsUUID(tc.val)
			if got != tc.want {
				t.Errorf("IsUUID(%q) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
}

func TestResolveParents_UUIDValue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")
	_ = cmd.Flags().Set("org", "e5ad0642-f9b3-48b3-9bf4-35997febe1fb")

	called := false
	parents := []ParentScope{{
		Name: "org", Flag: "org", IDFlag: "org-id",
		Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, _ string) (string, error) {
			called = true
			return "resolved", nil
		},
	}}

	ids, err := resolveParents(context.Background(), nil, cmd, parents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "e5ad0642-f9b3-48b3-9bf4-35997febe1fb" {
		t.Errorf("expected UUID passthrough, got %v", ids)
	}
	if called {
		t.Error("resolver should not be called when value is a UUID")
	}
}

func TestResolveParents_NameValueCallsResolver(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")
	_ = cmd.Flags().Set("org", "acme")

	parents := []ParentScope{{
		Name: "org", Flag: "org", IDFlag: "org-id",
		Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, name string) (string, error) {
			if name != "acme" {
				return "", fmt.Errorf("unexpected name: %s", name)
			}
			return "resolved-id", nil
		},
	}}

	ids, err := resolveParents(context.Background(), nil, cmd, parents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids[0] != "resolved-id" {
		t.Errorf("expected resolved-id, got %s", ids[0])
	}
}

func TestResolveParents_EmptyValue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")

	parents := []ParentScope{{
		Name: "org", Flag: "org", IDFlag: "org-id",
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error when flag value is empty")
	}
	if got := err.Error(); got != "--org is required" {
		t.Errorf("expected error '--org is required', got: %v", got)
	}
}

func TestResolveParents_NonUUIDWithNoResolver(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")
	_ = cmd.Flags().Set("org", "acme")

	parents := []ParentScope{{
		Name: "org", Flag: "org", IDFlag: "org-id",
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error when value is not a UUID and no resolver configured")
	}
	errMsg := err.Error()
	if errMsg != `--org: "acme" is not a valid UUID and name resolution is not configured for org` {
		t.Errorf("unexpected error: %v", errMsg)
	}
}

func TestResolveParents_MultiParent(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")
	cmd.Flags().String("ws", "", "")
	_ = cmd.Flags().Set("org", "acme")
	_ = cmd.Flags().Set("ws", "e5ad0642-f9b3-48b3-9bf4-35997febe1fb")

	parents := []ParentScope{
		{
			Name: "org", Flag: "org", IDFlag: "org-id",
			Resolver: func(_ context.Context, _ *terrakube.Client, resolvedIDs []string, name string) (string, error) {
				if len(resolvedIDs) != 0 {
					return "", fmt.Errorf("org resolver should have no resolved IDs, got %v", resolvedIDs)
				}
				return "org-id-123", nil
			},
		},
		{
			Name: "workspace", Flag: "ws", IDFlag: "ws-id",
			Resolver: func(_ context.Context, _ *terrakube.Client, resolvedIDs []string, name string) (string, error) {
				if len(resolvedIDs) != 1 || resolvedIDs[0] != "org-id-123" {
					return "", fmt.Errorf("ws resolver expected resolved org ID, got %v", resolvedIDs)
				}
				return "ws-id-456", nil
			},
		},
	}

	ids, err := resolveParents(context.Background(), nil, cmd, parents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "org-id-123" {
		t.Errorf("expected org-id-123, got %s", ids[0])
	}
	// ws was a UUID, so should be used directly (not resolved)
	if ids[1] != "e5ad0642-f9b3-48b3-9bf4-35997febe1fb" {
		t.Errorf("expected UUID passthrough for ws, got %s", ids[1])
	}
}

func TestResolveParents_ResolverError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org", "", "")
	_ = cmd.Flags().Set("org", "nonexistent")

	parents := []ParentScope{{
		Name: "org", Flag: "org", IDFlag: "org-id",
		Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, _ string) (string, error) {
			return "", fmt.Errorf("no organization found with name %q", "nonexistent")
		},
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error from resolver")
	}
	if got := err.Error(); got != `no organization found with name "nonexistent"` {
		t.Errorf("unexpected error: %v", got)
	}
}
