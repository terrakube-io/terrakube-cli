package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.ModuleVersion]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "module-version",
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:     "module",
				IDFlag:   "module-id",
				NameFlag: "module-name",
				Resolver: moduleResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Version", Flag: "version", Type: resource.String, Required: true, Description: "Module version"},
			{StructField: "Commit", Flag: "commit", Type: resource.String, Description: "Commit reference"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.ModuleVersion, error) {
			return c.ModuleVersions.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.ModuleVersion, error) {
			return c.ModuleVersions.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.ModuleVersion) (*terrakube.ModuleVersion, error) {
			return c.ModuleVersions.Create(ctx, pIDs[0], pIDs[1], v)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.ModuleVersion) (*terrakube.ModuleVersion, error) {
			return c.ModuleVersions.Update(ctx, pIDs[0], pIDs[1], v)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.ModuleVersions.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
