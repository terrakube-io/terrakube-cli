package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var VariableFilter string
var VariableOrgId string
var VariableWorkspaceId string
var VariableListExample string = `List all existing variables for a workspace
    %[1]v workspace variable list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd
List specific variable applying a filter
    %[1]v workspace variable list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd --filter key==myvariable `

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
	registerOrgFlag(listVariablesCmd, &VariableOrgId)
	registerWsFlag(listVariablesCmd, &VariableWorkspaceId)
}

func listVariables() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, VariableOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsID, err := resolveWs(ctx, client, orgID, VariableWorkspaceId)
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := &terrakube.ListOptions{Filter: VariableFilter}
	resp, err := client.Variables.List(ctx, orgID, wsID, opts)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
