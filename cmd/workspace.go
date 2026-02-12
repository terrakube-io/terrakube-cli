package cmd

import (
	"github.com/spf13/cobra"
)

var workspaceLong = `
This command consists of multiple subcommands to interact with workspaces.
It can be used to create, update, delete and list workspaces.
`

var workspaceCmd = &cobra.Command{
	Use:     "workspace create|update|delete|list [ARGS]",
	Short:   "create, update, delete and list workspaces",
	Long:    workspaceLong,
	Aliases: []string{"wrk", "ws", "workspaces"},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
