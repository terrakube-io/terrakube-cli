package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var TeamDeleteExample string = `Delete a Team
    %[1]v team delete -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

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
	registerOrgFlag(deleteTeamCmd, &TeamDeleteOrgId)
	deleteTeamCmd.Flags().StringVarP(&TeamDeleteId, "id", "", "", "Id of the Team (required)")
	_ = deleteTeamCmd.MarkFlagRequired("id")
}

func deleteTeam() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, TeamDeleteOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Teams.Delete(ctx, orgID, TeamDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("deleted")
}
