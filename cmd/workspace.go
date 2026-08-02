package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Workspace]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "workspace",
		Aliases: []string{"ws", "workspaces"},
		Parents: []resource.ParentScope{
			{
				Name:      "organization",
				Flag:      "organization",
				ShortFlag: "o",
				Aliases:   []string{"org"},
				IDFlag:    "organization-id",
				Resolver:  orgResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Workspace name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Workspace description"},
			{StructField: "Source", Flag: "source", Short: "s", Type: resource.String, Description: "Repository source URL"},
			{StructField: "Branch", Flag: "branch", Short: "b", Type: resource.String, Description: "VCS branch"},
			{StructField: "Folder", Flag: "folder", Short: "f", Type: resource.String, Description: "Workspace working folder"},
			{StructField: "IaCType", Flag: "iac-type", Short: "t", Type: resource.String, Description: "IaC type (terraform, tofu)"},
			{StructField: "IaCVersion", Flag: "iac-version", Short: "v", Type: resource.String, Description: "Terraform/Tofu version"},
			{StructField: "ExecutionMode", Flag: "execution-mode", Short: "e", Type: resource.String, Description: "Execution mode (remote, local)"},
			{StructField: "Deleted", Flag: "deleted", Type: resource.Bool, Description: "Mark workspace as deleted"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Workspace, error) {
			return c.Workspaces.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Workspace, error) {
			return c.Workspaces.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, ws *terrakube.Workspace) (*terrakube.Workspace, error) {
			return c.Workspaces.Create(ctx, pIDs[0], ws)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, ws *terrakube.Workspace) (*terrakube.Workspace, error) {
			return c.Workspaces.Update(ctx, pIDs[0], ws)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Workspaces.Delete(ctx, pIDs[0], id)
		},
	})
}
