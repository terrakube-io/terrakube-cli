package cmd

import (
	"context"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	cmd := resource.Register(rootCmd, resource.Config[terrakube.PolicyAttachment]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-attachment",
		Aliases: []string{"policy-attachments"},
		Parents: policySetScope,
		Fields: []resource.FieldDef{
			{StructField: "Workspace", Flag: "workspace", Short: "w", Type: resource.Relationship, RelType: "workspace", Resolver: workspaceResolver, Description: "Attach to this workspace, by name or ID"},
			{StructField: "Project", Flag: "project", Type: resource.Relationship, RelType: "project", Resolver: projectResolver, Description: "Attach to every workspace of this project, by name or ID"},
			{StructField: "Tag", Flag: "tag", Type: resource.Relationship, RelType: "tag", Resolver: tagResolver, Description: "Attach to every workspace with this tag, by name or ID"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.PolicyAttachment, error) {
			return c.PolicyAttachments.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicyAttachment, error) {
			return c.PolicyAttachments.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.PolicyAttachment, error) {
			return c.PolicyAttachments.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.PolicyAttachment, error) {
			return c.PolicyAttachments.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.PolicyAttachments.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
	oneTarget(cmd, true, "workspace", "project", "tag")
}

// oneTarget lets create and update take at most one of the target flags, and
// with required, makes create need one. Cobra checks this before RunE, so a
// wrong combination fails before any request.
func oneTarget(resCmd *cobra.Command, required bool, flags ...string) {
	for _, sub := range resCmd.Commands() {
		switch sub.Name() {
		case "create":
			if required {
				sub.MarkFlagsOneRequired(flags...)
			}
			sub.MarkFlagsMutuallyExclusive(flags...)
		case "update":
			sub.MarkFlagsMutuallyExclusive(flags...)
		}
	}
}
