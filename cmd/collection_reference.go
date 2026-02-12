package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.CollectionReference]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "collection-reference",
		Aliases: []string{"collection-references", "collection-refs"},
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
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Reference description"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.CollectionReference, error) {
			return c.CollectionReferences.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.CollectionReference, error) {
			return c.CollectionReferences.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, ref *terrakube.CollectionReference) (*terrakube.CollectionReference, error) {
			return c.CollectionReferences.Create(ctx, pIDs[0], pIDs[1], ref)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, ref *terrakube.CollectionReference) (*terrakube.CollectionReference, error) {
			return c.CollectionReferences.Update(ctx, ref)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.CollectionReferences.Delete(ctx, id)
		},
	})
}
