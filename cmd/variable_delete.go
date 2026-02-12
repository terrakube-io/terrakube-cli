package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VariableDeleteExample string = `Delete a variable
    %[1]v workspace variable delete -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

var VariableDeleteId string
var VariableDeleteOrgId string
var VariableDeleteWorkspaceId string

var deleteVariableCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a variable",
	Run: func(cmd *cobra.Command, args []string) {
		deleteVariable()
	},
	Example: fmt.Sprintf(VariableDeleteExample, rootCmd.Use),
}

func init() {
	variableCmd.AddCommand(deleteVariableCmd)
	registerOrgFlag(deleteVariableCmd, &VariableDeleteOrgId)
	deleteVariableCmd.Flags().StringVarP(&VariableDeleteId, "id", "", "", "Id of the variable (required)")
	_ = deleteVariableCmd.MarkFlagRequired("id")
	registerWsFlag(deleteVariableCmd, &VariableDeleteWorkspaceId)
}

func deleteVariable() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, VariableDeleteOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsID, err := resolveWs(ctx, client, orgID, VariableDeleteWorkspaceId)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Variables.Delete(ctx, orgID, wsID, VariableDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}
}
