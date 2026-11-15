package cmd

import (
	"context"
	"fmt"
	"strings"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

// nameFilter builds an RSQL equality filter with the name as a quoted string,
// so spaces, ';', ',', quotes and parentheses cannot alter the filter.
func nameFilter(name string) *terrakube.ListOptions {
	return &terrakube.ListOptions{Filter: "name==" + resource.QuoteRSQL(name)}
}

// resolveByName returns the ID of the single item whose name equals name.
// Exact matching is re-checked here because Elide treats '*' as a wildcard
// even in quoted RSQL values. kind is the singular resource name; orgID, when
// set, names the organization searched in the not-found message. fields
// returns an item's ID and name.
func resolveByName[T any](items []*T, err error, kind, name, orgID string, fields func(*T) (string, string)) (string, error) {
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
		if orgID == "" {
			return "", fmt.Errorf("%s %q not found", kind, name)
		}
		return "", fmt.Errorf("%s %q not found in organization %s", kind, name, orgID)
	}
	if len(ids) > 1 {
		return "", fmt.Errorf("multiple %ss match %q: %s; pass the ID instead", kind, name, strings.Join(ids, ", "))
	}
	return ids[0], nil
}

func orgResolver(ctx context.Context, c *terrakube.Client, _ []string, name string) (string, error) {
	items, err := c.Organizations.List(ctx, nameFilter(name))
	return resolveByName(items, err, "organization", name, "", func(o *terrakube.Organization) (string, string) { return o.ID, o.Name })
}

func workspaceResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Workspaces.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "workspace", name, pIDs[0], func(w *terrakube.Workspace) (string, string) { return w.ID, w.Name })
}

func collectionResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Collections.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "collection", name, pIDs[0], func(x *terrakube.Collection) (string, string) { return x.ID, x.Name })
}

func providerResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Providers.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "provider", name, pIDs[0], func(x *terrakube.Provider) (string, string) { return x.ID, x.Name })
}

func moduleResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Modules.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "module", name, pIDs[0], func(x *terrakube.Module) (string, string) { return x.ID, x.Name })
}

func notificationConfigurationResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.NotificationConfigurations.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "notification-configuration", name, pIDs[0], func(x *terrakube.NotificationConfiguration) (string, string) { return x.ID, x.Name })
}

// Resolvers for flags that link to an organization's template, VCS, SSH key,
// agent, project or tag. pIDs[0] is always the organization.

func templateResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Templates.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "template", name, pIDs[0], func(x *terrakube.Template) (string, string) { return x.ID, x.Name })
}

func vcsResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.VCS.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "vcs", name, pIDs[0], func(x *terrakube.VCS) (string, string) { return x.ID, x.Name })
}

func sshResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.SSH.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "ssh", name, pIDs[0], func(x *terrakube.SSH) (string, string) { return x.ID, x.Name })
}

func agentResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Agents.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "agent", name, pIDs[0], func(x *terrakube.Agent) (string, string) { return x.ID, x.Name })
}

func projectResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Projects.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "project", name, pIDs[0], func(x *terrakube.Project) (string, string) { return x.ID, x.Name })
}

func tagResolver(ctx context.Context, c *terrakube.Client, pIDs []string, name string) (string, error) {
	items, err := c.Tags.List(ctx, pIDs[0], nameFilter(name))
	return resolveByName(items, err, "tag", name, pIDs[0], func(x *terrakube.Tag) (string, string) { return x.ID, x.Name })
}
