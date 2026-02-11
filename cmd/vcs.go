package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.VCS]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "vcs",
		Parents: []resource.ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
			Resolver: orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "VCS connection name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Required: true, Description: "VCS description"},
			{StructField: "VcsType", Flag: "vcs-type", Type: resource.String, Required: true, Description: "VCS type (GITHUB, GITLAB, BITBUCKET, AZURE_DEVOPS)"},
			{StructField: "ConnectionType", Flag: "connection-type", Type: resource.String, Required: true, Description: "Connection type (OAUTH, SSH)"},
			{StructField: "ClientID", Flag: "client-id", Type: resource.String, Required: true, Description: "OAuth client ID"},
			{StructField: "ClientSecret", Flag: "client-secret", Type: resource.String, Required: true, Description: "OAuth client secret"},
			{StructField: "PrivateKey", Flag: "private-key", Type: resource.String, Description: "SSH private key"},
			{StructField: "Endpoint", Flag: "endpoint", Type: resource.String, Required: true, Description: "VCS endpoint URL"},
			{StructField: "APIURL", Flag: "vcs-api-url", Type: resource.String, Required: true, Description: "VCS API URL"},
			{StructField: "Status", Flag: "status", Type: resource.String, Description: "VCS connection status"},
			{StructField: "Callback", Flag: "callback", Type: resource.String, Description: "OAuth callback URL"},
			{StructField: "AccessToken", Flag: "access-token", Type: resource.String, Description: "Access token"},
			{StructField: "RedirectURL", Flag: "redirect-url", Type: resource.String, Description: "OAuth redirect URL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.VCS, error) {
			return c.VCS.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.VCS, error) {
			return c.VCS.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.VCS) (*terrakube.VCS, error) {
			return c.VCS.Create(ctx, pIDs[0], v)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.VCS) (*terrakube.VCS, error) {
			return c.VCS.Update(ctx, pIDs[0], v)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.VCS.Delete(ctx, pIDs[0], id)
		},
	})
}
