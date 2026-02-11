package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var WorkspaceDeleteExample string = `Delete a workspace
    %[1]v workspace delete --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

var WorkspaceDeleteId string
var WorkspaceDeleteOrgId string
var deleteWorkspaceCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a workspace",
	Run: func(cmd *cobra.Command, args []string) {
		deleteWorkspace()
	},
	Example: fmt.Sprintf(WorkspaceDeleteExample, rootCmd.Use),
}

func init() {
	workspaceCmd.AddCommand(deleteWorkspaceCmd)
	deleteWorkspaceCmd.Flags().StringVarP(&WorkspaceDeleteOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = deleteWorkspaceCmd.MarkFlagRequired("organization-id")
	deleteWorkspaceCmd.Flags().StringVarP(&WorkspaceDeleteId, "id", "", "", "Id of the workspace (required)")
	_ = deleteWorkspaceCmd.MarkFlagRequired("id")
}

func deleteWorkspace() {
	client := newClient()
	ctx := getContext()

	err := client.Workspaces.Delete(ctx, WorkspaceDeleteOrgId, WorkspaceDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}
}
