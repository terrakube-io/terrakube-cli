package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

var moduleCmd = resource.Register(rootCmd, resource.Config[terrakube.Module]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "module",
		Aliases: []string{"modules", "mod"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Module name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Module description"},
			{StructField: "Provider", Flag: "provider", Short: "p", Type: resource.String, Required: true, Description: "Provider name (e.g. aws, azurerm)"},
			{StructField: "Source", Flag: "source", Short: "s", Type: resource.String, Required: true, Description: "Source Git URL"},
			{StructField: "Folder", Flag: "folder", Type: resource.String, Description: "Folder path inside repo"},
			{StructField: "TagPrefix", Flag: "tag-prefix", Type: resource.String, Description: "Tag prefix for versioning"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Module, error) {
			return c.Modules.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Module, error) {
			return c.Modules.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, m *terrakube.Module) (*terrakube.Module, error) {
			return c.Modules.Create(ctx, pIDs[0], m)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, m *terrakube.Module) (*terrakube.Module, error) {
			return c.Modules.Update(ctx, pIDs[0], m)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Modules.Delete(ctx, pIDs[0], id)
		},
	})
