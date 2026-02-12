package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Action]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "action",
		Fields: []resource.FieldDef{
			{StructField: "Action", Flag: "action", Type: resource.String, Description: "Action identifier"},
			{StructField: "Active", Flag: "active", Type: resource.Bool, Description: "Whether the action is active"},
			{StructField: "Category", Flag: "category", Type: resource.String, Description: "Action category"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Action description"},
			{StructField: "DisplayCriteria", Flag: "display-criteria", Type: resource.String, Description: "Display criteria"},
			{StructField: "Label", Flag: "label", Type: resource.String, Description: "Action label"},
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Action name"},
			{StructField: "Type", Flag: "type", Type: resource.String, Description: "Action type"},
			{StructField: "Version", Flag: "version", Type: resource.String, Description: "Action version"},
		},
		List: func(ctx context.Context, c *terrakube.Client, _ []string, opts *terrakube.ListOptions) ([]*terrakube.Action, error) {
			return c.Actions.List(ctx, opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.Action, error) {
			return c.Actions.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, _ []string, a *terrakube.Action) (*terrakube.Action, error) {
			return c.Actions.Create(ctx, a)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, a *terrakube.Action) (*terrakube.Action, error) {
			return c.Actions.Update(ctx, a)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.Actions.Delete(ctx, id)
		},
	})
}
