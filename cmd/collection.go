package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Collection]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "collection",
		Aliases: []string{"collections"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Collection name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Collection description"},
			{StructField: "Priority", Flag: "priority", Type: resource.Int, Description: "Collection priority"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Collection, error) {
			return c.Collections.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Collection, error) {
			return c.Collections.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, col *terrakube.Collection) (*terrakube.Collection, error) {
			return c.Collections.Create(ctx, pIDs[0], col)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, col *terrakube.Collection) (*terrakube.Collection, error) {
			return c.Collections.Update(ctx, pIDs[0], col)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Collections.Delete(ctx, pIDs[0], id)
		},
	})
}
