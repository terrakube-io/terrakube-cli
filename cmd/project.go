package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Project]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "project",
		Aliases: []string{"projects"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Project name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Project description"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Project, error) {
			return c.Projects.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Project, error) {
			return c.Projects.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p *terrakube.Project) (*terrakube.Project, error) {
			return c.Projects.Create(ctx, pIDs[0], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, p *terrakube.Project) (*terrakube.Project, error) {
			return c.Projects.Update(ctx, pIDs[0], p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Projects.Delete(ctx, pIDs[0], id)
		},
	})
}
