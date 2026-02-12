package resource

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

// resolveParents resolves parent resource IDs from flags.
// Each parent's unified flag accepts either a UUID (used directly) or a name (resolved via Resolver).
func resolveParents(ctx context.Context, client *terrakube.Client, cmd *cobra.Command, parents []ParentScope) ([]string, error) {
	ids := make([]string, 0, len(parents))

	for _, p := range parents {
		id, err := resolveParent(ctx, client, cmd, p, ids)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func resolveParent(ctx context.Context, client *terrakube.Client, cmd *cobra.Command, p ParentScope, resolvedIDs []string) (string, error) {
	val, _ := cmd.Flags().GetString(p.Flag)
	if val == "" {
		return "", fmt.Errorf("--%s is required", p.Flag)
	}

	if IsUUID(val) {
		return val, nil
	}

	if p.Resolver == nil {
		return "", fmt.Errorf("--%s: %q is not a valid UUID and name resolution is not configured for %s", p.Flag, val, p.Name)
	}

	return p.Resolver(ctx, client, resolvedIDs, val)
}

// IsUUID returns true if s is a valid UUID.
func IsUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
