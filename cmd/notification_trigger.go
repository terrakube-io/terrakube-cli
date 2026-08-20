package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.NotificationTrigger]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "notification-trigger",
		Aliases: []string{"nt", "notification-triggers", "trigger", "triggers"},
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
				Name:      "notification-configuration",
				Flag:      "notification-configuration",
				ShortFlag: "n",
				Aliases:   []string{"nc", "notification"},
				IDFlag:    "notification-configuration-id",
				Resolver:  notificationConfigurationResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "JobStatus", Flag: "job-status", Short: "s", Type: resource.String, Required: true, Description: "Job status triggering notification (completed, failed, etc.)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.NotificationTrigger, error) {
			return c.NotificationTriggers.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.NotificationTrigger, error) {
			return c.NotificationTriggers.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, r *terrakube.NotificationTrigger) (*terrakube.NotificationTrigger, error) {
			return c.NotificationTriggers.Create(ctx, pIDs[0], pIDs[1], r)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, r *terrakube.NotificationTrigger) (*terrakube.NotificationTrigger, error) {
			return c.NotificationTriggers.Update(ctx, pIDs[0], pIDs[1], r)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.NotificationTriggers.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
