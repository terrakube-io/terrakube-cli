package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.GithubAppToken]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "github-app-token",
		Aliases: []string{"github-app-tokens"},
		Fields: []resource.FieldDef{
			{StructField: "AppID", Flag: "app-id", Type: resource.String, Required: true, Description: "GitHub App ID"},
			{StructField: "InstallationID", Flag: "installation-id", Type: resource.String, Required: true, Description: "GitHub App installation ID"},
			{StructField: "Owner", Flag: "owner", Type: resource.String, Required: true, Description: "GitHub App owner"},
			{StructField: "ExpiresAt", Flag: "expires-at", Type: resource.String, Description: "Token expiration (RFC 3339, e.g. 2026-01-02T15:04:05Z)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, _ []string, opts *terrakube.ListOptions) ([]*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.List(ctx, opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, _ []string, p terrakube.Patch) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.CreateFields(ctx, p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, id string, p terrakube.Patch) (*terrakube.GithubAppToken, error) {
			return c.GithubAppTokens.UpdateFields(ctx, id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.GithubAppTokens.Delete(ctx, id)
		},
	})
}
