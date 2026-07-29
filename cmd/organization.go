package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Organization]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "organization",
		Aliases: []string{"org", "orgs", "organizations"},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Organization name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Organization description"},
			{StructField: "ExecutionMode", Flag: "execution-mode", Short: "e", Type: resource.String, Description: "Execution mode (remote or local)"},
			{StructField: "Icon", Flag: "icon", Short: "i", Type: resource.String, Description: "Organization icon"},
			{StructField: "Disabled", Flag: "disabled", Type: resource.Bool, Description: "Whether organization is disabled"},
		},
		List: func(ctx context.Context, c *terrakube.Client, _ []string, opts *terrakube.ListOptions) ([]*terrakube.Organization, error) {
			return c.Organizations.List(ctx, opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.Organization, error) {
			return c.Organizations.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, _ []string, org *terrakube.Organization) (*terrakube.Organization, error) {
			return c.Organizations.Create(ctx, org)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, org *terrakube.Organization) (*terrakube.Organization, error) {
			return c.Organizations.Update(ctx, org)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.Organizations.Delete(ctx, id)
		},
	})
}
