package cmd

import (
	"context"
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func workspaceResolver(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error) {
	wss, err := c.Workspaces.List(ctx, resolvedParentIDs[0], &terrakube.ListOptions{Filter: "name==" + name})
	if err != nil {
		return "", err
	}
	if len(wss) == 0 {
		return "", fmt.Errorf("no workspace found with name %q", name)
	}
	if len(wss) > 1 {
		return "", fmt.Errorf("multiple workspaces match name %q, use --workspace-id", name)
	}
	return wss[0].ID, nil
}

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.WorkspaceTag]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "workspace-tag",
		Aliases: []string{"wstag"},
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:     "workspace",
				IDFlag:   "workspace-id",
				NameFlag: "workspace-name",
				Resolver: workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "TagID", Flag: "tag-id", Type: resource.String, Required: true, Description: "Tag ID to associate"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.WorkspaceTag, error) {
			return c.WorkspaceTags.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.WorkspaceTag, error) {
			return c.WorkspaceTags.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.WorkspaceTag) (*terrakube.WorkspaceTag, error) {
			return c.WorkspaceTags.Create(ctx, pIDs[0], pIDs[1], t)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.WorkspaceTag) (*terrakube.WorkspaceTag, error) {
			return c.WorkspaceTags.Update(ctx, pIDs[0], pIDs[1], t)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.WorkspaceTags.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
