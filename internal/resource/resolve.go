package resource

import (
	"context"
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

// resolveParents resolves parent resource IDs from flags.
// For each parent scope, checks --<parent>-id first, then --<parent>-name.
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
	idVal, _ := cmd.Flags().GetString(p.IDFlag)
	if idVal != "" {
		return idVal, nil
	}

	if p.NameFlag == "" {
		return "", fmt.Errorf("--%s is required", p.IDFlag)
	}

	nameVal, _ := cmd.Flags().GetString(p.NameFlag)
	if nameVal == "" {
		return "", fmt.Errorf("either --%s or --%s is required", p.IDFlag, p.NameFlag)
	}

	if p.Resolver == nil {
		return "", fmt.Errorf("name resolution not configured for %s", p.Name)
	}

	return p.Resolver(ctx, client, resolvedIDs, nameVal)
}
