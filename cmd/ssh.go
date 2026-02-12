package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.SSH]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "ssh",
		Parents: []resource.ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
			Resolver: orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "SSH key name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "SSH key description"},
			{StructField: "PrivateKey", Flag: "private-key", Type: resource.String, Required: true, Description: "SSH private key"},
			{StructField: "SSHType", Flag: "ssh-type", Type: resource.String, Required: true, Description: "SSH key type (rsa, ed25519)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.SSH, error) {
			return c.SSH.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.SSH, error) {
			return c.SSH.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.SSH) (*terrakube.SSH, error) {
			return c.SSH.Create(ctx, pIDs[0], s)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, s *terrakube.SSH) (*terrakube.SSH, error) {
			return c.SSH.Update(ctx, pIDs[0], s)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.SSH.Delete(ctx, pIDs[0], id)
		},
	})
}
