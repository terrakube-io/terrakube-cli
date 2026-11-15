package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.PolicySet]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-set",
		Aliases: []string{"policy-sets"},
		Parents: orgScope,
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Policy set name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Policy set description"},
			{StructField: "EnforcementLevel", Flag: "enforcement-level", Type: resource.String, Description: "ADVISORY, SOFT_MANDATORY or HARD_MANDATORY (server default)"},
			{StructField: "ShadowEnforcementLevel", Flag: "shadow-enforcement-level", Type: resource.String, Description: "Enforcement level evaluated in shadow mode (reported, never blocks)"},
			{StructField: "OverrideTeam", Flag: "override-team", Type: resource.String, Description: "Team allowed to override soft-mandatory failures"},
			{StructField: "Global", Flag: "global", Type: resource.Bool, Description: "Apply to every workspace in the organization"},
			{StructField: "Repository", Flag: "repository", Type: resource.String, Description: "Git repository with the policies"},
			{StructField: "Branch", Flag: "branch", Short: "b", Type: resource.String, Description: "Repository branch (server default: main)"},
			{StructField: "Folder", Flag: "folder", Short: "f", Type: resource.String, Description: "Folder with the policies (server default: /)"},
			{StructField: "OpaVersion", Flag: "opa-version", Type: resource.String, Description: "OPA version"},
			{StructField: "VCS", Flag: "vcs", Type: resource.Relationship, RelType: "vcs", Resolver: vcsResolver, Description: "VCS connection for a private repository, by name or ID (\"\" clears it)"},
			{StructField: "NotificationConfiguration", Flag: "notification-configuration", Type: resource.Relationship, RelType: "notification_configuration", Resolver: notificationConfigurationResolver, Description: "Notification configuration, by name or ID (\"\" clears it)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.PolicySet, error) {
			return c.PolicySets.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicySet, error) {
			return c.PolicySets.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.PolicySet, error) {
			return c.PolicySets.CreateFields(ctx, pIDs[0], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.PolicySet, error) {
			return c.PolicySets.UpdateFields(ctx, pIDs[0], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.PolicySets.Delete(ctx, pIDs[0], id)
		},
	})
}
