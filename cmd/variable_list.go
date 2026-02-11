package cmd

import (
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

var VariableFilter string
var VariableOrgId string
var VariableWorkspaceId string
var VariableListExample string = `List all existing variables for a workspace
    %[1]v workspace variable list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd
List specific variable applying a filter
    %[1]v workspace variable list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd --filter key==myvariable `

var listVariablesCmd = &cobra.Command{
	Use:   "list",
	Short: "list variables",
	Run: func(cmd *cobra.Command, args []string) {
		listVariables()
	},
	Example: fmt.Sprintf(VariableListExample, rootCmd.Use),
}

func init() {
	variableCmd.AddCommand(listVariablesCmd)
	listVariablesCmd.Flags().StringVarP(&VariableFilter, "filter", "f", "", "Filter")
	listVariablesCmd.Flags().StringVarP(&VariableOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = listVariablesCmd.MarkFlagRequired("organization-id")
	listVariablesCmd.Flags().StringVarP(&VariableWorkspaceId, "workspace-id", "w", "", "Workspace Id (required)")
	_ = listVariablesCmd.MarkFlagRequired("workspace-id")
}

func listVariables() {
	client := newTerrakubeClient()
	ctx := getContext()

	opts := &terrakube.ListOptions{Filter: VariableFilter}
	resp, err := client.Variables.List(ctx, VariableOrgId, VariableWorkspaceId, opts)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
