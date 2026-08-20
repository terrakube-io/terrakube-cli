package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.NotificationConfiguration]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "notification-configuration",
		Aliases: []string{"nc", "notification-configurations", "notification", "notifications"},
		Parents: []resource.ParentScope{
			{
				Name:      "organization",
				Flag:      "organization",
				ShortFlag: "o",
				Aliases:   []string{"org"},
				IDFlag:    "organization-id",
				Resolver:  orgResolver,
			},
			{
				Name:      "workspace",
				Flag:      "workspace",
				ShortFlag: "w",
				Aliases:   []string{"ws"},
				IDFlag:    "workspace-id",
				Optional:  true,
				Resolver:  workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Notification configuration name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Notification configuration description"},
			{StructField: "ChannelType", Flag: "channel-type", Type: resource.String, Required: true, Description: "Channel type (SLACK, TEAMS, WEBHOOK)"},
			{StructField: "DestinationURL", Flag: "destination-url", Type: resource.String, Required: true, Description: "Destination URL"},
			{StructField: "SigningSecret", Flag: "signing-secret", Type: resource.String, Description: "Signing secret for payload verification"},
			{StructField: "Active", Flag: "active", Type: resource.Bool, Description: "Whether notification configuration is active"},
			{StructField: "MessageStyle", Flag: "message-style", Type: resource.String, Description: "Message style (DETAILED, SIMPLE)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.NotificationConfiguration, error) {
			if len(pIDs) > 1 && pIDs[1] != "" {
				return c.NotificationConfigurations.ListByWorkspace(ctx, pIDs[0], pIDs[1], opts)
			}
			return c.NotificationConfigurations.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.NotificationConfiguration, error) {
			return c.NotificationConfigurations.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, r *terrakube.NotificationConfiguration) (*terrakube.NotificationConfiguration, error) {
			if len(pIDs) > 1 && pIDs[1] != "" {
				return c.NotificationConfigurations.CreateForWorkspace(ctx, pIDs[0], pIDs[1], r)
			}
			return c.NotificationConfigurations.Create(ctx, pIDs[0], r)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, r *terrakube.NotificationConfiguration) (*terrakube.NotificationConfiguration, error) {
			return c.NotificationConfigurations.Update(ctx, pIDs[0], r)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.NotificationConfigurations.Delete(ctx, pIDs[0], id)
		},
	})
}
