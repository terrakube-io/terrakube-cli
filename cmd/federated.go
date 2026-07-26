package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Federated]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "federated",
		Aliases: []string{"federation"},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Federated connection name"},
			{StructField: "IssuerURL", Flag: "issuer-url", Type: resource.String, Required: true, Description: "OIDC issuer URL"},
			{StructField: "Audience", Flag: "audience", Type: resource.String, Required: true, Description: "Audience value"},
		},
		List: func(ctx context.Context, c *terrakube.Client, _ []string, opts *terrakube.ListOptions) ([]*terrakube.Federated, error) {
			return c.Federated.List(ctx, opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.Federated, error) {
			return c.Federated.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, _ []string, f *terrakube.Federated) (*terrakube.Federated, error) {
			return c.Federated.Create(ctx, f)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, f *terrakube.Federated) (*terrakube.Federated, error) {
			return c.Federated.Update(ctx, f)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.Federated.Delete(ctx, id)
		},
	})
}
