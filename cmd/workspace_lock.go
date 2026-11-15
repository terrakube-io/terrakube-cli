package cmd

import (
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

// newWorkspaceLockCmd returns workspace lock (lock=true) or unlock (lock=false).
func newWorkspaceLockCmd(lock bool) *cobra.Command {
	name, short := "unlock", "unlock a workspace so new runs can start"
	if lock {
		name, short = "lock", "lock a workspace so no new run starts until it is unlocked"
	}
	cmd := &cobra.Command{
		Use:          name + " <workspace>",
		Short:        short,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			ctx := getContext()
			ids, err := resource.ResolveScope(ctx, client, cmd, orgScope)
			if err != nil {
				return err
			}
			wsID := args[0]
			if !resource.IsUUID(wsID) {
				if wsID, err = workspaceResolver(ctx, client, ids, wsID); err != nil {
					return err
				}
			}
			p := terrakube.Patch{Attributes: map[string]any{"locked": lock}}
			if reason, _ := cmd.Flags().GetString("reason"); lock && reason != "" {
				p.Attributes["lockDescription"] = reason
			}
			ws, err := client.Workspaces.UpdateFields(ctx, ids[0], wsID, p)
			if err != nil {
				return err
			}
			return outputpkg.Render(os.Stdout, ws, output)
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	if lock {
		cmd.Flags().String("reason", "", "why the workspace is locked (shown in the UI)")
	}
	return cmd
}
