package cmd

import (
	"github.com/spf13/cobra"
)

var variableLong = `
This command consists of multiple subcommands to interact with variables.
It can be used to create, update, delete and list variables within a workspace.
`

var variableCmd = &cobra.Command{
	Use:     "variable create|update|delete|list [ARGS]",
	Short:   "create, update, delete and list variables",
	Long:    variableLong,
	Aliases: []string{"var", "vars", "variables"},
}

func init() {
	workspaceCmd.AddCommand(variableCmd)
}
