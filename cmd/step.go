package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Step]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "step",
		Aliases: []string{"steps"},
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
				Name:      "job",
				Flag:      "job",
				ShortFlag: "j",
				IDFlag:    "job-id",
				RawID:     true,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Step name"},
			{StructField: "Output", Flag: "output", Type: resource.String, Hidden: true, Description: "Step output (executor only)"},
			{StructField: "Status", Flag: "status", Type: resource.String, Hidden: true, Description: "Step status (executor only)"},
			{StructField: "StepNumber", Flag: "step-number", Type: resource.Int, Description: "Step number"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Step, error) {
			return c.Steps.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Step, error) {
			return c.Steps.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.Step, error) {
			return c.Steps.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.Step, error) {
			return c.Steps.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Steps.Delete(ctx, pIDs[0], pIDs[1], id)
		},
		// Only executor or super-service tokens may call these; the server answers 403 otherwise.
		HiddenVerbs: []string{"create", "delete"},
	})
}
