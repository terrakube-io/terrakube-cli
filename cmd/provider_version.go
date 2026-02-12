package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.ProviderVersion]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "provider-version",
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:     "provider",
				IDFlag:   "provider-id",
				NameFlag: "provider-name",
				Resolver: providerResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "VersionNumber", Flag: "version-number", Type: resource.String, Required: true, Description: "Provider version number"},
			{StructField: "Protocols", Flag: "protocols", Type: resource.String, Description: "Supported protocols"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.ProviderVersion, error) {
			return c.ProviderVersions.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.ProviderVersion, error) {
			return c.ProviderVersions.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.ProviderVersion) (*terrakube.ProviderVersion, error) {
			return c.ProviderVersions.Create(ctx, pIDs[0], pIDs[1], v)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.ProviderVersion) (*terrakube.ProviderVersion, error) {
			return c.ProviderVersions.Update(ctx, pIDs[0], pIDs[1], v)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.ProviderVersions.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
