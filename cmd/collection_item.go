package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.CollectionItem]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "collection-item",
		Aliases: []string{"collection-items"},
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
				Name:     "collection",
				Flag:     "collection",
				IDFlag:   "collection-id",
				Resolver: collectionResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Key", Flag: "key", Short: "k", Type: resource.String, Required: true, Description: "Item key"},
			{StructField: "Value", Flag: "value", Short: "v", Type: resource.String, Required: true, Description: "Item value"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Item description"},
			{StructField: "Category", Flag: "category", Type: resource.String, Description: "Item category"},
			{StructField: "Sensitive", Flag: "sensitive", Type: resource.Bool, Description: "Whether the item is sensitive"},
			{StructField: "Hcl", Flag: "hcl", Type: resource.Bool, Description: "Whether the value is HCL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.CollectionItem, error) {
			return c.CollectionItems.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.CollectionItem, error) {
			return c.CollectionItems.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.CollectionItem, error) {
			return c.CollectionItems.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.CollectionItem, error) {
			return c.CollectionItems.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.CollectionItems.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
