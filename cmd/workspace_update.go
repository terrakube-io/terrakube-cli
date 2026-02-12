package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var WorkspaceUpdateExample string = `Update Terraform version in workspace
    %[1]v workspace update -o 312b4415-806b-47a9-9452-b71f0753136e --id 38b6635a-d38e-46f2-a95e-d00a416de4fd -t 0.14.0 `

var WorkspaceUpdateName string
var WorkspaceUpdateSource string
var WorkspaceUpdateBranch string
var WorkspaceUpdateTerraformV string
var WorkspaceUpdateOrgId string
var WorkspaceUpdateId string
var WorkspaceUpdateDescription string
var WorkspaceUpdateFolder string
var WorkspaceUpdateIacType string
var WorkspaceUpdateExecutionMode string

var updateWorkspaceCmd = &cobra.Command{
	Use:   "update",
	Short: "update a workspace",
	Run: func(cmd *cobra.Command, args []string) {
		updateWorkspace()
	},
	Example: fmt.Sprintf(WorkspaceUpdateExample, rootCmd.Use),
}

func init() {
	workspaceCmd.AddCommand(updateWorkspaceCmd)
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateName, "name", "n", "", "Name of the workspace (required)")
	registerOrgFlag(updateWorkspaceCmd, &WorkspaceUpdateOrgId)
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateId, "id", "", "", "Id of the workspace (required)")
	_ = updateWorkspaceCmd.MarkFlagRequired("id")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateBranch, "branch", "b", "", "Branch of the workspace")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateSource, "source", "s", "", "Source of the workspace")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateTerraformV, "iac-version", "v", "", "terraform/tofu Version use in the workspace")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateDescription, "description", "d", "", "Workspace description")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateFolder, "folder", "f", "/", "Workspace folder")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateExecutionMode, "execution-mode", "e", "remote", "Workspace execution mode")
	updateWorkspaceCmd.Flags().StringVarP(&WorkspaceUpdateIacType, "iac-type", "t", "terraform", "Iac type")
}

func updateWorkspace() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, WorkspaceUpdateOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	workspace := &terrakube.Workspace{
		ID:            WorkspaceUpdateId,
		Name:          WorkspaceUpdateName,
		Description:   ptrOrNil(WorkspaceUpdateDescription),
		Folder:        WorkspaceUpdateFolder,
		IaCType:       WorkspaceUpdateIacType,
		ExecutionMode: WorkspaceUpdateExecutionMode,
		Branch:        WorkspaceUpdateBranch,
		Source:        WorkspaceUpdateSource,
		IaCVersion:    WorkspaceUpdateTerraformV,
	}

	_, err = client.Workspaces.Update(ctx, orgID, workspace)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Updated")
}
