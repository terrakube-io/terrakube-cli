package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.RunTrigger]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "run-trigger",
		Aliases: []string{"run-triggers"},
		// --workspace is the destination: the workspace that gets the runs.
		Parents: runScope,
		Fields: []resource.FieldDef{
			{StructField: "SourceWorkspace", Flag: "source", Type: resource.Relationship, RelType: "workspace", Resolver: workspaceResolver, Required: true, Description: "Source workspace whose applies start a run in --workspace, by name or ID (fixed once created)"},
			{StructField: "Enabled", Flag: "enabled", Type: resource.Bool, Description: "Fire the trigger (server default: true)"},
			{StructField: "Template", Flag: "template", Short: "t", Type: resource.Relationship, RelType: "template", Resolver: templateResolver, Description: "Template for the triggered run, by name or ID (\"\" uses the destination's default template)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.RunTrigger, error) {
			return c.RunTriggers.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.RunTrigger, error) {
			return c.RunTriggers.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.RunTrigger, error) {
			// Sent explicitly, as the UI does, rather than left to the parent path.
			p.Relationships["destinationWorkspace"] = terrakube.Relationship{Type: "workspace", ID: &pIDs[1]}
			return c.RunTriggers.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.RunTrigger, error) {
			return c.RunTriggers.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.RunTriggers.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
