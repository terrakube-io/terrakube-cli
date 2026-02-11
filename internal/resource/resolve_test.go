package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

func TestResolveParents_IDFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")
	_ = cmd.Flags().Set("org-id", "abc-123")

	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id", NameFlag: "org-name",
	}}

	ids, err := resolveParents(context.Background(), nil, cmd, parents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "abc-123" {
		t.Errorf("expected [abc-123], got %v", ids)
	}
}

func TestResolveParents_IDFlagTakesPrecedence(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")
	_ = cmd.Flags().Set("org-id", "id-value")
	_ = cmd.Flags().Set("org-name", "name-value")

	called := false
	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id", NameFlag: "org-name",
		Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, _ string) (string, error) {
			called = true
			return "resolved", nil
		},
	}}

	ids, err := resolveParents(context.Background(), nil, cmd, parents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids[0] != "id-value" {
		t.Errorf("expected id-value, got %s", ids[0])
	}
	if called {
		t.Error("resolver should not be called when ID flag is set")
	}
}

func TestResolveParents_NameFlagCallsResolver(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")
	_ = cmd.Flags().Set("org-name", "acme")

	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id", NameFlag: "org-name",
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

func TestResolveParents_NeitherIDNorName(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")

	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id", NameFlag: "org-name",
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error when neither ID nor name is set")
	}
	if !strings.Contains(err.Error(), "org-id") || !strings.Contains(err.Error(), "org-name") {
		t.Errorf("error should mention both flags, got: %v", err)
	}
}

func TestResolveParents_NoNameFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")

	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id",
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error when ID not set and no name flag configured")
	}
	if !strings.Contains(err.Error(), "org-id") {
		t.Errorf("error should mention id flag, got: %v", err)
	}
}

func TestResolveParents_MultiParent(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")
	cmd.Flags().String("ws-id", "", "")
	cmd.Flags().String("ws-name", "", "")
	_ = cmd.Flags().Set("org-name", "acme")
	_ = cmd.Flags().Set("ws-name", "prod")

	parents := []ParentScope{
		{
			Name: "org", IDFlag: "org-id", NameFlag: "org-name",
			Resolver: func(_ context.Context, _ *terrakube.Client, resolvedIDs []string, name string) (string, error) {
				if len(resolvedIDs) != 0 {
					return "", fmt.Errorf("org resolver should have no resolved IDs, got %v", resolvedIDs)
				}
				return "org-id-123", nil
			},
		},
		{
			Name: "workspace", IDFlag: "ws-id", NameFlag: "ws-name",
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
	if ids[1] != "ws-id-456" {
		t.Errorf("expected ws-id-456, got %s", ids[1])
	}
}

func TestResolveParents_ResolverError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("org-id", "", "")
	cmd.Flags().String("org-name", "", "")
	_ = cmd.Flags().Set("org-name", "nonexistent")

	parents := []ParentScope{{
		Name: "org", IDFlag: "org-id", NameFlag: "org-name",
		Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, _ string) (string, error) {
			return "", fmt.Errorf("no organization found with name %q", "nonexistent")
		},
	}}

	_, err := resolveParents(context.Background(), nil, cmd, parents)
	if err == nil {
		t.Fatal("expected error from resolver")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to contain name, got: %v", err)
	}
}
