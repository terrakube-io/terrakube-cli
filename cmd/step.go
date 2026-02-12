package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Step]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "step",
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:   "job",
				IDFlag: "job-id",
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Step name"},
			{StructField: "Output", Flag: "output", Type: resource.String, Description: "Step output"},
			{StructField: "Status", Flag: "status", Type: resource.String, Description: "Step status"},
			{StructField: "StepNumber", Flag: "step-number", Type: resource.Int, Description: "Step number"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Step, error) {
			return c.Steps.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Step, error) {
			return c.Steps.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.Step) (*terrakube.Step, error) {
			return c.Steps.Create(ctx, pIDs[0], pIDs[1], s)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.Step) (*terrakube.Step, error) {
			return c.Steps.Update(ctx, pIDs[0], pIDs[1], s)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Steps.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
