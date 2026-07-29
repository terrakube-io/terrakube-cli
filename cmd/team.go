package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Team]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "team",
		Aliases: []string{"teams"},
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
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Team name"},
			{StructField: "ManageWorkspace", Flag: "manage-workspace", Type: resource.Bool, Description: "Manage Workspace permissions"},
			{StructField: "ManageModule", Flag: "manage-module", Type: resource.Bool, Description: "Manage Module permissions"},
			{StructField: "ManageProvider", Flag: "manage-provider", Type: resource.Bool, Description: "Manage Provider permissions"},
			{StructField: "ManageState", Flag: "manage-state", Type: resource.Bool, Description: "Manage State permissions"},
			{StructField: "ManageCollection", Flag: "manage-collection", Type: resource.Bool, Description: "Manage Collection permissions"},
			{StructField: "ManageVcs", Flag: "manage-vcs", Type: resource.Bool, Description: "Manage VCS permissions"},
			{StructField: "ManageTemplate", Flag: "manage-template", Type: resource.Bool, Description: "Manage Template permissions"},
			{StructField: "ManageJob", Flag: "manage-job", Type: resource.Bool, Description: "Manage Job permissions"},
			{StructField: "PlanJob", Flag: "plan-job", Type: resource.Bool, Description: "Plan Job permissions"},
			{StructField: "ApproveJob", Flag: "approve-job", Type: resource.Bool, Description: "Approve Job permissions"},
			{StructField: "Role", Flag: "role", Type: resource.String, Description: "Team role"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Team, error) {
			return c.Teams.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Team, error) {
			return c.Teams.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, team *terrakube.Team) (*terrakube.Team, error) {
			return c.Teams.Create(ctx, pIDs[0], team)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, team *terrakube.Team) (*terrakube.Team, error) {
			return c.Teams.Update(ctx, pIDs[0], team)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Teams.Delete(ctx, pIDs[0], id)
		},
	})
}
