package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var WorkspaceFilter string
var WorkspaceOrgId string
var WorkspaceListExample string = `List all existing workspaces
    %[1]v workspace list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb
List specific organizations applying a filter
    %[1]v workspace list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb --filter name==mymodule `
var listWorkspacesCmd = &cobra.Command{
	Use:   "list",
	Short: "list workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		listWorkspaces()
	},
	Example: fmt.Sprintf(WorkspaceListExample, rootCmd.Use),
}

func init() {
	workspaceCmd.AddCommand(listWorkspacesCmd)
	listWorkspacesCmd.Flags().StringVarP(&WorkspaceFilter, "filter", "f", "", "Filter")
	registerOrgFlag(listWorkspacesCmd, &WorkspaceOrgId)
}

func listWorkspaces() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, WorkspaceOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := client.Workspaces.List(ctx, orgID, &terrakube.ListOptions{Filter: WorkspaceFilter})

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
