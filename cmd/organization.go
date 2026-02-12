package cmd

import (
	"github.com/spf13/cobra"
)

var organizationLong = `
This command consists of multiple subcommands to interact with organizations.
It can be used to create, update, delete and list organizations.
`

var organizationCmd = &cobra.Command{
	Use:     "organization create|update|delete|list [ARGS]",
	Short:   "create, update, delete and list organizations",
	Long:    organizationLong,
	Aliases: []string{"org", "orgs", "organizations"},
}

func init() {
	rootCmd.AddCommand(organizationCmd)
}
