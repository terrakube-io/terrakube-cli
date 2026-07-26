package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
)

var teamTokenCmd = &cobra.Command{
	Use:     "token",
	Short:   "manage team tokens",
	Aliases: []string{"tokens"},
}

var createTeamTokenCmd = &cobra.Command{
	Use:   "create",
	Short: "generate a new team token",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()

		desc, _ := cmd.Flags().GetString("description")
		group, _ := cmd.Flags().GetString("group")
		days, _ := cmd.Flags().GetInt32("days")
		hours, _ := cmd.Flags().GetInt32("hours")
		minutes, _ := cmd.Flags().GetInt32("minutes")

		token := &terrakube.TeamToken{
			Description: desc,
			Group:       group,
			Days:        days,
			Hours:       hours,
			Minutes:     minutes,
		}

		res, err := client.TeamTokens.Create(ctx, token)
		if err != nil {
			return err
		}

		renderOutput(res, output)
		return nil
	},
}

var listTeamTokensCmd = &cobra.Command{
	Use:   "list",
	Short: "list team tokens",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()

		tokens, err := client.TeamTokens.List(ctx)
		if err != nil {
			return err
		}

		renderOutput(tokens, output)
		return nil
	},
}

var deleteTeamTokenCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a team token",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()

		id, _ := cmd.Flags().GetString("id")
		if err := client.TeamTokens.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Println("Team token deleted")
		return nil
	},
}

func init() {
	createTeamTokenCmd.Flags().StringP("description", "d", "", "Token description")
	createTeamTokenCmd.Flags().StringP("group", "g", "", "Team/Group name")
	createTeamTokenCmd.Flags().Int32("days", 0, "Token validity in days")
	createTeamTokenCmd.Flags().Int32("hours", 0, "Token validity in hours")
	createTeamTokenCmd.Flags().Int32("minutes", 0, "Token validity in minutes")
	_ = createTeamTokenCmd.MarkFlagRequired("group")

	deleteTeamTokenCmd.Flags().String("id", "", "Token ID")
	_ = deleteTeamTokenCmd.MarkFlagRequired("id")

	teamTokenCmd.AddCommand(createTeamTokenCmd)
	teamTokenCmd.AddCommand(listTeamTokensCmd)
	teamTokenCmd.AddCommand(deleteTeamTokenCmd)

	teamCmd.AddCommand(teamTokenCmd)
}
