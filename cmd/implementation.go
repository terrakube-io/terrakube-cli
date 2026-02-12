package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Implementation]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "implementation",
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
			{
				Name:   "provider-version",
				IDFlag: "provider-version-id",
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Os", Flag: "os", Type: resource.String, Required: true, Description: "Operating system"},
			{StructField: "Arch", Flag: "arch", Type: resource.String, Required: true, Description: "Architecture"},
			{StructField: "Filename", Flag: "filename", Type: resource.String, Required: true, Description: "Filename"},
			{StructField: "DownloadURL", Flag: "download-url", Type: resource.String, Description: "Download URL"},
			{StructField: "ShasumsURL", Flag: "shasums-url", Type: resource.String, Description: "SHA sums URL"},
			{StructField: "ShasumsSignatureURL", Flag: "shasums-signature-url", Type: resource.String, Description: "SHA sums signature URL"},
			{StructField: "Shasum", Flag: "shasum", Type: resource.String, Description: "SHA sum"},
			{StructField: "KeyID", Flag: "key-id", Type: resource.String, Description: "GPG key ID"},
			{StructField: "ASCIIArmor", Flag: "ascii-armor", Type: resource.String, Description: "ASCII armor GPG key"},
			{StructField: "TrustSignature", Flag: "trust-signature", Type: resource.String, Description: "Trust signature"},
			{StructField: "Source", Flag: "source", Type: resource.String, Description: "Source"},
			{StructField: "SourceURL", Flag: "source-url", Type: resource.String, Description: "Source URL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Implementation, error) {
			return c.Implementations.List(ctx, pIDs[0], pIDs[1], pIDs[2], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Implementation, error) {
			return c.Implementations.Get(ctx, pIDs[0], pIDs[1], pIDs[2], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, impl *terrakube.Implementation) (*terrakube.Implementation, error) {
			return c.Implementations.Create(ctx, pIDs[0], pIDs[1], pIDs[2], impl)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, impl *terrakube.Implementation) (*terrakube.Implementation, error) {
			return c.Implementations.Update(ctx, pIDs[0], pIDs[1], pIDs[2], impl)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Implementations.Delete(ctx, pIDs[0], pIDs[1], pIDs[2], id)
		},
	})
}
