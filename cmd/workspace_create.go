package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var WorkspaceCreateExample string = `Create a new workspace
    %[1]v workspace create -o 312b4415-806b-47a9-9452-b71f0753136e -n myWorkspace -s https://github.com/terrakube-io/terraform-sample-repository.git -b master -t 0.15.0`

var WorkspaceCreateName string
var WorkspaceDescription string
var WorkspaceCreateIacType string
var WorkspaceCreateFolder string
var WorkspaceExecutionMode string
var WorkspaceCreateSource string
var WorkspaceCreateBranch string
var WorkspaceCreateCli bool
var WorkspaceCreateIacV string
var WorkspaceCreateOrgId string
var createWorkspaceCmd = &cobra.Command{
	Use:   "create",
	Short: "create a workspace",
	Run: func(cmd *cobra.Command, args []string) {
		// Adjust branch/source based on CLI flag after flags are parsed
		if WorkspaceCreateCli {
			fmt.Println("Creating cli workspace")
			WorkspaceCreateBranch = "remote-content"
			WorkspaceCreateSource = "empty"
		} else {
			fmt.Println("Creating vcs workspace")
		}
		createWorkspace()
	},
	Example: fmt.Sprintf(WorkspaceCreateExample, rootCmd.Use),
}

func init() {
	workspaceCmd.AddCommand(createWorkspaceCmd)
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateName, "name", "n", "", "Name of the new workspace (required)")
	_ = createWorkspaceCmd.MarkFlagRequired("name")
	registerOrgFlag(createWorkspaceCmd, &WorkspaceCreateOrgId)
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateBranch, "branch", "b", "", "Branch of the new workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateSource, "source", "s", "", "Source of the new workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateIacV, "iac-version", "v", "", "Terraform/tofu Version use in the new workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateFolder, "folder", "f", "/", "Folder of the new workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceExecutionMode, "execution-mode", "e", "remote", "Execution mode for workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceCreateIacType, "iac-type", "t", "terraform", "IAC Type for workspace")
	createWorkspaceCmd.Flags().StringVarP(&WorkspaceDescription, "description", "d", "", "Description of the new workspace")
	createWorkspaceCmd.Flags().BoolVarP(&WorkspaceCreateCli, "cli", "c", false, "Create a CLI workspace")
}

func createWorkspace() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, WorkspaceCreateOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	workspace := &terrakube.Workspace{
		Name:          WorkspaceCreateName,
		Description:   ptrOrNil(WorkspaceDescription),
		Folder:        WorkspaceCreateFolder,
		Source:        WorkspaceCreateSource,
		Branch:        WorkspaceCreateBranch,
		IaCType:       WorkspaceCreateIacType,
		ExecutionMode: WorkspaceExecutionMode,
		IaCVersion:    WorkspaceCreateIacV,
	}

	resp, err := client.Workspaces.Create(ctx, orgID, workspace)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
