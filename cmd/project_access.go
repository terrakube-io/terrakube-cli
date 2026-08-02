package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.ProjectAccess]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "project-access",
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
				Name:   "project",
				Flag:   "project",
				IDFlag: "project-id",
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Team name"},
			{StructField: "ManageState", Flag: "manage-state", Type: resource.Bool, Description: "Allow state management"},
			{StructField: "ManageJob", Flag: "manage-job", Type: resource.Bool, Description: "Allow job management"},
			{StructField: "ManageWorkspace", Flag: "manage-workspace", Type: resource.Bool, Description: "Allow workspace management"},
			{StructField: "PlanJob", Flag: "plan-job", Type: resource.Bool, Description: "Allow plan job execution"},
			{StructField: "ApproveJob", Flag: "approve-job", Type: resource.Bool, Description: "Allow approving jobs"},
			{StructField: "Role", Flag: "role", Type: resource.String, Description: "Project access role"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.ProjectAccess, error) {
			return c.ProjectAccess.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.ProjectAccess, error) {
			return c.ProjectAccess.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.ProjectAccess) (*terrakube.ProjectAccess, error) {
			return c.ProjectAccess.Create(ctx, pIDs[0], pIDs[1], a)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.ProjectAccess) (*terrakube.ProjectAccess, error) {
			return c.ProjectAccess.Update(ctx, pIDs[0], pIDs[1], a)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.ProjectAccess.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
