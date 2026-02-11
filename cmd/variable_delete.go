package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VariableDeleteExample string = `Delete a variable
    %[1]v workspace variable delete --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb  -w 38b6635a-d38e-46f2-a95e-d00a416de4fd --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

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
	deleteVariableCmd.Flags().StringVarP(&VariableDeleteOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = deleteVariableCmd.MarkFlagRequired("organization-id")
	deleteVariableCmd.Flags().StringVarP(&VariableDeleteId, "id", "", "", "Id of the variable (required)")
	_ = deleteVariableCmd.MarkFlagRequired("id")
	deleteVariableCmd.Flags().StringVarP(&VariableDeleteWorkspaceId, "workspace-id", "w", "", "Workspace Id (required)")
	_ = deleteVariableCmd.MarkFlagRequired("workspace-id")
}

func deleteVariable() {
	client := newClient()
	ctx := getContext()

	err := client.Variables.Delete(ctx, VariableDeleteOrgId, VariableDeleteWorkspaceId, VariableDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}
}
