package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Address]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "address",
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
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Address name"},
			{StructField: "Type", Flag: "type", Type: resource.String, Required: true, Description: "Address type"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Address, error) {
			return c.Addresses.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Address, error) {
			return c.Addresses.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.Address) (*terrakube.Address, error) {
			return c.Addresses.Create(ctx, pIDs[0], pIDs[1], a)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.Address) (*terrakube.Address, error) {
			return c.Addresses.Update(ctx, pIDs[0], pIDs[1], a)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Addresses.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
