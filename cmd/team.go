package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

var teamCmd = resource.Register(rootCmd, resource.Config[terrakube.Team]{
	Runtime: resource.Runtime{
		NewClient:  newClient,
		GetContext: getContext,
		GetOutput:  func() string { return output },
	},
	Name:    "team",
	Aliases: []string{"teams"},
	Parents: []resource.ParentScope{{
		Name:      "organization",
		Flag:      "organization",
		ShortFlag: "o",
		Aliases:   []string{"org"},
		IDFlag:    "organization-id",
		Resolver:  orgResolver,
	}},
	Fields: []resource.FieldDef{
		{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Team name"},
		{StructField: "ManageState", Flag: "manage-state", Type: resource.Bool, Description: "Allow state management"},
		{StructField: "ManageWorkspace", Flag: "manage-workspace", Type: resource.Bool, Description: "Allow workspace management"},
		{StructField: "ManageModule", Flag: "manage-module", Type: resource.Bool, Description: "Allow module management"},
		{StructField: "ManageProvider", Flag: "manage-provider", Type: resource.Bool, Description: "Allow provider management"},
		{StructField: "ManageVcs", Flag: "manage-vcs", Type: resource.Bool, Description: "Allow VCS management"},
		{StructField: "ManageTemplate", Flag: "manage-template", Type: resource.Bool, Description: "Allow template management"},
		{StructField: "ManageJob", Flag: "manage-job", Type: resource.Bool, Description: "Allow job management"},
		{StructField: "ManageCollection", Flag: "manage-collection", Type: resource.Bool, Description: "Allow collection management"},
		{StructField: "PlanJob", Flag: "plan-job", Type: resource.Bool, Description: "Allow plan job execution"},
		{StructField: "ApproveJob", Flag: "approve-job", Type: resource.Bool, Description: "Allow approving jobs"},
		{StructField: "Role", Flag: "role", Type: resource.String, Description: "Team role"},
	},
	List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Team, error) {
		return c.Teams.List(ctx, pIDs[0], opts)
	},
	Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Team, error) {
		return c.Teams.Get(ctx, pIDs[0], id)
	},
	Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Team) (*terrakube.Team, error) {
		return c.Teams.Create(ctx, pIDs[0], t)
	},
	Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Team) (*terrakube.Team, error) {
		return c.Teams.Update(ctx, pIDs[0], t)
	},
	Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
		return c.Teams.Delete(ctx, pIDs[0], id)
	},
})
