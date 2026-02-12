package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var teamLong = `
This command consists of multiple subcommands to interact with modules.
It can be used to create, update, delete and list teams.
`

var teamCmd = &cobra.Command{
	Use:     "team create|update|delete|list [ARGS]",
	Short:   "create, update, delete and list teams",
	Long:    teamLong,
	Aliases: []string{"teams"},
}

func init() {
	rootCmd.AddCommand(teamCmd)
	_ = viper.BindEnv("organization", "TERRAKUBE_ORGANIZATION_ID")
}
