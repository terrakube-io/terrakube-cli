package cmd

import (
	"context"
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
)

func collectionResolver(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error) {
	cols, err := c.Collections.List(ctx, resolvedParentIDs[0], &terrakube.ListOptions{Filter: "name==" + name})
	if err != nil {
		return "", err
	}
	if len(cols) == 0 {
		return "", fmt.Errorf("no collection found with name %q", name)
	}
	if len(cols) > 1 {
		return "", fmt.Errorf("multiple collections match name %q, use --collection-id", name)
	}
	return cols[0].ID, nil
}

func providerResolver(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error) {
	provs, err := c.Providers.List(ctx, resolvedParentIDs[0], &terrakube.ListOptions{Filter: "name==" + name})
	if err != nil {
		return "", err
	}
	if len(provs) == 0 {
		return "", fmt.Errorf("no provider found with name %q", name)
	}
	if len(provs) > 1 {
		return "", fmt.Errorf("multiple providers match name %q, use --provider-id", name)
	}
	return provs[0].ID, nil
}

func moduleResolver(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error) {
	mods, err := c.Modules.List(ctx, resolvedParentIDs[0], &terrakube.ListOptions{Filter: "name==" + name})
	if err != nil {
		return "", err
	}
	if len(mods) == 0 {
		return "", fmt.Errorf("no module found with name %q", name)
	}
	if len(mods) > 1 {
		return "", fmt.Errorf("multiple modules match name %q, use --module-id", name)
	}
	return mods[0].ID, nil
}
