package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

// whoamiResult is what whoami prints.
type whoamiResult struct {
	Teams        []string               `json:"teams"`
	Organization string                 `json:"organization,omitempty"`
	Workspace    string                 `json:"workspace,omitempty"`
	Permissions  *terrakube.Permissions `json:"permissions,omitempty"`
}

func init() {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "show your teams and, with --org (and --workspace), your permissions there",
		Long: `Show the identity provider groups (teams) of the logged-in user. With --org, also
show the user's effective permissions in that organization; with --workspace as well,
the permissions on that workspace. Both take a name or an ID.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			org, _ := cmd.Flags().GetString("org")
			ws, _ := cmd.Flags().GetString("workspace")
			if ws != "" && org == "" {
				return errors.New("--workspace needs --org")
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			ctx := getContext()
			teams, err := client.Access.CurrentTeams(ctx)
			if err != nil {
				return err
			}
			res := whoamiResult{Teams: teams}
			if org != "" {
				if res.Organization, err = idOrName(org, func(name string) (string, error) {
					return orgResolver(ctx, client, nil, name)
				}); err != nil {
					return err
				}
				if ws == "" {
					res.Permissions, err = client.Access.Permissions(ctx, res.Organization)
				} else {
					if res.Workspace, err = idOrName(ws, func(name string) (string, error) {
						return workspaceResolver(ctx, client, []string{res.Organization}, name)
					}); err != nil {
						return err
					}
					res.Permissions, err = client.Access.WorkspacePermissions(ctx, res.Organization, res.Workspace)
				}
				if err != nil {
					return err
				}
			}
			return outputpkg.Render(os.Stdout, res, output)
		},
	}
	cmd.Flags().String("org", "", "organization name or ID")
	cmd.Flags().String("workspace", "", "workspace name or ID (needs --org)")
	rootCmd.AddCommand(cmd)
}

// idOrName returns v when it is a UUID, otherwise what resolve finds for it.
func idOrName(v string, resolve func(string) (string, error)) (string, error) {
	if resource.IsUUID(v) {
		return v, nil
	}
	return resolve(v)
}
