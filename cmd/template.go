package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Template]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "template",
		Aliases: []string{"tpl", "templates"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Template name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Template description"},
			{StructField: "Version", Flag: "version", Type: resource.String, Description: "Template version"},
			{StructField: "Content", Flag: "content", Short: "c", Type: resource.String, Required: true, Description: "Template content (HCL)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Template, error) {
			return c.Templates.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Template, error) {
			return c.Templates.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Template) (*terrakube.Template, error) {
			return c.Templates.Create(ctx, pIDs[0], t)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, t *terrakube.Template) (*terrakube.Template, error) {
			return c.Templates.Update(ctx, pIDs[0], t)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Templates.Delete(ctx, pIDs[0], id)
		},
	})
}
