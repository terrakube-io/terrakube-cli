package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.GithubAppToken]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "github-app-token",
		Fields: []resource.FieldDef{
			{StructField: "AppID", Flag: "app-id", Type: resource.String, Required: true, Description: "GitHub App ID"},
			{StructField: "InstallationID", Flag: "installation-id", Type: resource.String, Required: true, Description: "GitHub App installation ID"},
			{StructField: "Owner", Flag: "owner", Type: resource.String, Required: true, Description: "GitHub App owner"},
		},
		List: func(ctx context.Context, c *terrakube.Client, _ []string, opts *terrakube.ListOptions) ([]*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.List(ctx, opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, _ []string, t *terrakube.GithubAppToken) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.Create(ctx, t)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, t *terrakube.GithubAppToken) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.Update(ctx, t)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.GithubAppTokens.Delete(ctx, id)
		},
	})
}
