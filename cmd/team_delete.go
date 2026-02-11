package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var TeamDeleteExample string = `Delete a Team
    %[1]v team delete --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

var TeamDeleteId string
var TeamDeleteOrgId string

var deleteTeamCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a team",
	Run: func(cmd *cobra.Command, args []string) {
		deleteTeam()
	},
	Example: fmt.Sprintf(TeamDeleteExample, rootCmd.Use),
}

func init() {
	teamCmd.AddCommand(deleteTeamCmd)
	deleteTeamCmd.Flags().StringVarP(&TeamDeleteOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = deleteTeamCmd.MarkFlagRequired("organization-id")
	deleteTeamCmd.Flags().StringVarP(&TeamDeleteId, "id", "", "", "Id of the Team (required)")
	_ = deleteTeamCmd.MarkFlagRequired("id")
}

func deleteTeam() {
	client := newClient()
	ctx := getContext()

	err := client.Teams.Delete(ctx, TeamDeleteOrgId, TeamDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("deleted")
}
