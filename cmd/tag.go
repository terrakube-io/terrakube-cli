package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Tag]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "tag",
		Parents: []resource.ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
			Resolver: orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Tag name"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Tag, error) {
			return c.Tags.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Tag, error) {
			return c.Tags.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Tag) (*terrakube.Tag, error) {
			return c.Tags.Create(ctx, pIDs[0], t)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Tag) (*terrakube.Tag, error) {
			return c.Tags.Update(ctx, pIDs[0], t)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Tags.Delete(ctx, pIDs[0], id)
		},
	})
}
