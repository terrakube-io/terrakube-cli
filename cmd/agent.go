package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Agent]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "agent",
		Parents: []resource.ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
			Resolver: orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Agent name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Agent description"},
			{StructField: "URL", Flag: "url", Short: "u", Type: resource.String, Required: true, Description: "Agent URL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Agent, error) {
			return c.Agents.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Agent, error) {
			return c.Agents.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.Agent) (*terrakube.Agent, error) {
			return c.Agents.Create(ctx, pIDs[0], a)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, a *terrakube.Agent) (*terrakube.Agent, error) {
			return c.Agents.Update(ctx, pIDs[0], a)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Agents.Delete(ctx, pIDs[0], id)
		},
	})
}
