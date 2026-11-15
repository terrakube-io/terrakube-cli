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
		Name:    "address",
		Aliases: []string{"addresses"},
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
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Address name"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Address, error) {
			return c.Addresses.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Address, error) {
			return c.Addresses.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.Address, error) {
			return c.Addresses.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.Address, error) {
			return c.Addresses.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Addresses.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
