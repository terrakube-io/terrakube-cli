package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.WorkspaceSchedule]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "workspace-schedule",
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
			{StructField: "Schedule", Flag: "schedule", Type: resource.String, Required: true, Description: "Cron expression"},
			{StructField: "TemplateID", Flag: "template-id", Type: resource.String, Required: true, Description: "Template reference ID"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.WorkspaceSchedule, error) {
			return c.WorkspaceSchedules.List(ctx, pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.WorkspaceSchedule, error) {
			return c.WorkspaceSchedules.Get(ctx, pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.WorkspaceSchedule) (*terrakube.WorkspaceSchedule, error) {
			return c.WorkspaceSchedules.Create(ctx, pIDs[1], s)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.WorkspaceSchedule) (*terrakube.WorkspaceSchedule, error) {
			return c.WorkspaceSchedules.Update(ctx, pIDs[1], s)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.WorkspaceSchedules.Delete(ctx, pIDs[1], id)
		},
	})
}
