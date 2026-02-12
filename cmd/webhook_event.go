package cmd

import (
	"context"

	terrakube "github.com/denniswebb/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.WebhookEvent]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "webhook-event",
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
			{
				Name:   "webhook",
				IDFlag: "webhook-id",
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Branch", Flag: "branch", Type: resource.String, Description: "Branch to watch"},
			{StructField: "Event", Flag: "event", Type: resource.String, Description: "Event type"},
			{StructField: "Path", Flag: "path", Type: resource.String, Description: "Webhook event path"},
			{StructField: "Priority", Flag: "priority", Type: resource.Int, Description: "Priority"},
			{StructField: "TemplateID", Flag: "template-id", Type: resource.String, Description: "Template ID"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.WebhookEvent, error) {
			return c.WebhookEvents.List(ctx, pIDs[0], pIDs[1], pIDs[2], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.WebhookEvent, error) {
			return c.WebhookEvents.Get(ctx, pIDs[0], pIDs[1], pIDs[2], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, e *terrakube.WebhookEvent) (*terrakube.WebhookEvent, error) {
			return c.WebhookEvents.Create(ctx, pIDs[0], pIDs[1], pIDs[2], e)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, e *terrakube.WebhookEvent) (*terrakube.WebhookEvent, error) {
			return c.WebhookEvents.Update(ctx, pIDs[0], pIDs[1], pIDs[2], e)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.WebhookEvents.Delete(ctx, pIDs[0], pIDs[1], pIDs[2], id)
		},
	})
}
