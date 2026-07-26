package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

var workspaceCmd = resource.Register(rootCmd, resource.Config[terrakube.Workspace]{
	Runtime: resource.Runtime{
		NewClient:  newClient,
		GetContext: getContext,
		GetOutput:  func() string { return output },
	},
	Name:    "workspace",
	Aliases: []string{"workspaces", "ws", "wrk"},
	Parents: []resource.ParentScope{{
		Name:      "organization",
		Flag:      "organization",
		ShortFlag: "o",
		Aliases:   []string{"org"},
		IDFlag:    "organization-id",
		Resolver:  orgResolver,
	}},
	Fields: []resource.FieldDef{
		{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Workspace name"},
		{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Workspace description"},
		{StructField: "Source", Flag: "source", Type: resource.String, Description: "Git repository source URL"},
		{StructField: "Branch", Flag: "branch", Type: resource.String, Description: "Git repository branch"},
		{StructField: "Folder", Flag: "folder", Type: resource.String, Description: "Working directory folder inside repo"},
		{StructField: "TemplateID", Flag: "default-template", Type: resource.String, Description: "Default workflow template ID"},
		{StructField: "IaCType", Flag: "iac-type", Type: resource.String, Description: "IaC tool type (terraform, opentofu)"},
		{StructField: "IaCVersion", Flag: "terraform-version", Type: resource.String, Description: "IaC tool version"},
		{StructField: "ExecutionMode", Flag: "execution-mode", Type: resource.String, Description: "Execution mode (remote, local)"},
		{StructField: "Locked", Flag: "locked", Type: resource.Bool, Description: "Lock workspace"},
		{StructField: "AllowRemoteApply", Flag: "allow-remote-apply", Type: resource.Bool, Description: "Allow remote apply"},
		{StructField: "GlobalRemoteState", Flag: "global-remote-state", Type: resource.Bool, Description: "Enable global remote state access"},
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
