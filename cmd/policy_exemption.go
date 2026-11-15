package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	cmd := resource.Register(rootCmd, resource.Config[terrakube.PolicyExemption]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-exemption",
		Aliases: []string{"policy-exemptions"},
		Parents: orgScope,
		Fields: []resource.FieldDef{
			{StructField: "PolicySet", Flag: "policy-set", Type: resource.Relationship, RelType: "policy_set", Resolver: policySetResolver, Required: true, Description: "Policy set the rule belongs to, by name or ID"},
			{StructField: "RuleID", Flag: "rule", Type: resource.String, Required: true, Description: "ID of the exempted policy rule"},
			{StructField: "Justification", Flag: "justification", Type: resource.String, Required: true, Description: "Why the rule does not apply"},
			{StructField: "ExpiresAt", Flag: "expires-at", Type: resource.String, Required: true, Description: "When the exemption ends, RFC 3339 (2026-01-02T15:04:05Z)"},
			{StructField: "TicketReference", Flag: "ticket", Type: resource.String, Required: true, Description: "Ticket that tracks the exemption (required by the server)"},
			{StructField: "Workspace", Flag: "workspace", Short: "w", Type: resource.Relationship, RelType: "workspace", Resolver: workspaceResolver, Description: "Exempt only this workspace, by name or ID"},
			{StructField: "Project", Flag: "project", Type: resource.Relationship, RelType: "project", Resolver: projectResolver, Description: "Exempt only this project's workspaces, by name or ID"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.PolicyExemption, error) {
			return c.PolicyExemptions.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicyExemption, error) {
			return c.PolicyExemptions.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.PolicyExemption, error) {
			return c.PolicyExemptions.CreateFields(ctx, pIDs[0], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.PolicyExemption, error) {
			return c.PolicyExemptions.UpdateFields(ctx, pIDs[0], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.PolicyExemptions.Delete(ctx, pIDs[0], id)
		},
	})
	// Without --workspace or --project the exemption covers the whole organization.
	oneTarget(cmd, false, "workspace", "project")
}
