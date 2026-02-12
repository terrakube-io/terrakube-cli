package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"terrakube/internal/resource"
)

// setFlagAliases registers flag name aliases via pflag's normalize function.
// Aliases map old/short names to the canonical flag name. Multiple calls
// on the same command chain correctly.
func setFlagAliases(cmd *cobra.Command, aliases map[string]string) {
	prev := cmd.Flags().GetNormalizeFunc()
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if mapped, ok := aliases[name]; ok {
			return pflag.NormalizedName(mapped)
		}
		return prev(f, name)
	})
}

// registerOrgFlag registers --organization/-o and maps --organization-id and --org to it.
func registerOrgFlag(cmd *cobra.Command, dst *string) {
	cmd.Flags().StringVarP(dst, "organization", "o", "", "Organization ID or name (required)")
	_ = cmd.MarkFlagRequired("organization")
	setFlagAliases(cmd, map[string]string{
		"organization-id": "organization",
		"org":             "organization",
	})
}

// registerWsFlag registers --workspace/-w and maps --workspace-id and --ws to it.
func registerWsFlag(cmd *cobra.Command, dst *string) {
	cmd.Flags().StringVarP(dst, "workspace", "w", "", "Workspace ID or name (required)")
	_ = cmd.MarkFlagRequired("workspace")
	setFlagAliases(cmd, map[string]string{
		"workspace-id": "workspace",
		"ws":           "workspace",
	})
}

// resolveOrg resolves an organization value that may be a UUID or a name.
func resolveOrg(ctx context.Context, c *terrakube.Client, value string) (string, error) {
	if resource.IsUUID(value) {
		return value, nil
	}
	return orgResolver(ctx, c, nil, value)
}

// resolveWs resolves a workspace value that may be a UUID or a name.
func resolveWs(ctx context.Context, c *terrakube.Client, orgID, value string) (string, error) {
	if resource.IsUUID(value) {
		return value, nil
	}
	return workspaceResolver(ctx, c, []string{orgID}, value)
}
