package cmd

import (
	"fmt"
	"terrakube/client/models"

	"github.com/spf13/cobra"
)

var TeamUpdateExample string = `Update the permissions of the Team using id
    %[1]v team update --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd --manage-workspace true --manage-module true --manage-provider true" `

var TeamId string
var TeamUpdateName string
var TeamUpdateOrgId string
var TeamUpdateManageProvider bool
var TeamUpdateManageModule bool
var TeamUpdateManageWorkspace bool
var TeamUpdateManageState bool
var TeamUpdateManageCollection bool
var TeamUpdateManageVcs bool
var TeamUpdateManageTemplate bool

var updateTeamCmd = &cobra.Command{
	Use:   "update",
	Short: "update a Team",
	Run: func(cmd *cobra.Command, args []string) {
		updateTeam()
	},
	Example: fmt.Sprintf(TeamUpdateExample, rootCmd.Use),
}

func init() {
	teamCmd.AddCommand(updateTeamCmd)
	updateTeamCmd.AddCommand(updateOrganizationCmd)
	updateTeamCmd.Flags().StringVarP(&TeamId, "id", "", "", "Id of the Team (required)")
	_ = updateTeamCmd.MarkFlagRequired("id")
	updateTeamCmd.Flags().StringVarP(&TeamUpdateName, "name", "n", "", "Name of the Team")
	updateTeamCmd.Flags().StringVarP(&TeamUpdateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = updateTeamCmd.MarkFlagRequired("organization-id")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageProvider, "manage-provider", "", false, "Manage Provider Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageModule, "manage-module", "", false, "Manage Module Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageWorkspace, "manage-workspace", "", false, "Manage Workspaces Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageState, "manage-state", "", false, "Manage State Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageCollection, "manage-collection", "", false, "Manage Collection Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageVcs, "manage-vcs", "", false, "Manage VCS Permissions")
	updateTeamCmd.Flags().BoolVarP(&TeamUpdateManageTemplate, "manage-template", "", false, "Manage Template Permissions")
}

func updateTeam() {
	client := newClient()

	team := models.Team{
		Attributes: &models.TeamAttributes{
			Name:             TeamUpdateName,
			ManageWorkspace:  TeamUpdateManageWorkspace,
			ManageModule:     TeamUpdateManageModule,
			ManageProvider:   TeamUpdateManageProvider,
			ManageState:      TeamUpdateManageState,
			ManageCollection: TeamUpdateManageCollection,
			ManageVcs:        TeamUpdateManageVcs,
			ManageTemplate:   TeamUpdateManageTemplate,
		},
		ID:   TeamId,
		Type: "Team",
	}
	err := client.Team.Update(TeamUpdateOrgId, team)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Updated")

}
