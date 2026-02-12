package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.WorkspaceAccess]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "workspace-access",
		Aliases: []string{"workspace-accesses"},
		Parents: []resource.ParentScope{
			{
				Name:      "organization",
				Flag:      "organization",
				ShortFlag: "o",
				Aliases:   []string{"org"},
				IDFlag:    "organization-id",
				Resolver:  orgResolver,
			},
			{
				Name:      "workspace",
				Flag:      "workspace",
				ShortFlag: "w",
				Aliases:   []string{"ws"},
				IDFlag:    "workspace-id",
				Resolver:  workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "ManageState", Flag: "manage-state", Type: resource.Bool, Description: "Manage state permission"},
			{StructField: "ManageWorkspace", Flag: "manage-workspace", Type: resource.Bool, Description: "Manage workspace permission"},
			{StructField: "ManageJob", Flag: "manage-job", Type: resource.Bool, Description: "Manage job permission"},
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Team name for access"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.WorkspaceAccess, error) {
			return c.WorkspaceAccess.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.WorkspaceAccess, error) {
			return c.WorkspaceAccess.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.WorkspaceAccess) (*terrakube.WorkspaceAccess, error) {
			return c.WorkspaceAccess.Create(ctx, pIDs[0], pIDs[1], a)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.WorkspaceAccess) (*terrakube.WorkspaceAccess, error) {
			return c.WorkspaceAccess.Update(ctx, pIDs[0], pIDs[1], a)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.WorkspaceAccess.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
