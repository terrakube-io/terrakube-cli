package cmd

import (
	"context"
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func orgResolver(ctx context.Context, c *terrakube.Client, _ []string, name string) (string, error) {
	orgs, err := c.Organizations.List(ctx, &terrakube.ListOptions{Filter: "name==" + name})
	if err != nil {
		return "", err
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organization found with name %q", name)
	}
	if len(orgs) > 1 {
		return "", fmt.Errorf("multiple organizations match name %q, use --organization-id", name)
	}
	return orgs[0].ID, nil
}

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Template]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "template",
		Aliases: []string{"tpl"},
		Parents: []resource.ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
			Resolver: orgResolver,
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
