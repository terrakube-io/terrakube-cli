package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Webhook]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "webhook",
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:     "workspace",
				IDFlag:   "workspace-id",
				NameFlag: "workspace-name",
				Resolver: workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Path", Flag: "path", Type: resource.String, Description: "Webhook path"},
			{StructField: "Branch", Flag: "branch", Type: resource.String, Description: "Branch to watch"},
			{StructField: "TemplateID", Flag: "template-id", Type: resource.String, Description: "Template ID"},
			{StructField: "RemoteHookID", Flag: "remote-hook-id", Type: resource.String, Description: "Remote hook ID"},
			{StructField: "Event", Flag: "event", Type: resource.String, Description: "Event type"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Webhook, error) {
			return c.Webhooks.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Webhook, error) {
			return c.Webhooks.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, w *terrakube.Webhook) (*terrakube.Webhook, error) {
			return c.Webhooks.Create(ctx, pIDs[0], pIDs[1], w)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, w *terrakube.Webhook) (*terrakube.Webhook, error) {
			return c.Webhooks.Update(ctx, pIDs[0], pIDs[1], w)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Webhooks.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
