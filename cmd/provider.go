package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Provider]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "provider",
		Aliases: []string{"providers"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Provider name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Provider description"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Provider, error) {
			return c.Providers.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Provider, error) {
			return c.Providers.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p *terrakube.Provider) (*terrakube.Provider, error) {
			return c.Providers.Create(ctx, pIDs[0], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, p *terrakube.Provider) (*terrakube.Provider, error) {
			return c.Providers.Update(ctx, pIDs[0], p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Providers.Delete(ctx, pIDs[0], id)
		},
	})
}
