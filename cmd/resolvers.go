package cmd

import (
	"context"
	"fmt"
	"strings"

	terrakube "github.com/terrakube-io/terrakube-go"
)

// nameFilter builds an RSQL equality filter with the name as a quoted string,
// so spaces, ';', ',', quotes and parentheses cannot alter the filter.
func nameFilter(name string) *terrakube.ListOptions {
	quoted := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(name)
	return &terrakube.ListOptions{Filter: `name=="` + quoted + `"`}
}

// resolveByName returns the ID of the single item whose name equals name.
// Exact matching is re-checked here because Elide treats '*' as a wildcard
// even in quoted RSQL values. kind is the singular resource name, also used
// as the ID flag name. fields returns an item's ID and name.
func resolveByName[T any](items []*T, err error, kind, name string, fields func(*T) (string, string)) (string, error) {
	if err != nil {
		return "", err
	}
	var ids []string
	for _, it := range items {
		if id, n := fields(it); n == name {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("no %s found with name %q", kind, name)
	}
	if len(ids) > 1 {
		return "", fmt.Errorf("multiple %ss match name %q, use --%s with the ID", kind, name, kind)
	}
	return ids[0], nil
}

func orgResolver(ctx context.Context, c *terrakube.Client, _ []string, name string) (string, error) {
	items, err := c.Organizations.List(ctx, nameFilter(name))
	return resolveByName(items, err, "organization", name, func(o *terrakube.Organization) (string, string) { return o.ID, o.Name })
}

func workspaceResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Workspaces.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "workspace", name, func(w *terrakube.Workspace) (string, string) { return w.ID, w.Name })
}

func collectionResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Collections.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "collection", name, func(x *terrakube.Collection) (string, string) { return x.ID, x.Name })
}

func providerResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Providers.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "provider", name, func(x *terrakube.Provider) (string, string) { return x.ID, x.Name })
}

func moduleResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Modules.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "module", name, func(x *terrakube.Module) (string, string) { return x.ID, x.Name })
}

func notificationConfigurationResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.NotificationConfigurations.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "notification-configuration", name, func(x *terrakube.NotificationConfiguration) (string, string) { return x.ID, x.Name })
}
