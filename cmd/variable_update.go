package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var VariableUpdateExample string = `Update the value of the variable using id
    %[1]v workspace variable update --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd --id 38b6635a-d38e-46f2-a95e-d00a416de4fd -v "new value" `

var VariableId string
var VariableUpdateKey string
var VariableUpdateValue string
var VariableUpdateDescription string
var VariableUpdateCategory string
var VariableUpdateSensitive bool
var VariableUpdateHcl bool
var VariableUpdateOrgId string
var VariableUpdateWorkspaceId string

var updateVariableCmd = &cobra.Command{
	Use:   "update",
	Short: "update a variable",
	Run: func(cmd *cobra.Command, args []string) {
		updateVariable()
	},
	Example: fmt.Sprintf(VariableUpdateExample, rootCmd.Use),
}

func init() {
	variableCmd.AddCommand(updateVariableCmd)
	updateVariableCmd.Flags().StringVarP(&VariableId, "id", "", "", "Id of the variable (required)")
	_ = updateVariableCmd.MarkFlagRequired("id")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateKey, "key", "k", "", "Key of the variable")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateValue, "value", "v", "", "Value of the variable")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = updateVariableCmd.MarkFlagRequired("organization-id")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateWorkspaceId, "workspace-id", "w", "", "Workspace Id (required)")
	_ = updateVariableCmd.MarkFlagRequired("workspace-id")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateDescription, "description", "d", "", "Description of the variable")
	updateVariableCmd.Flags().StringVarP(&VariableUpdateCategory, "category", "c", "", "Category of the variable. Valid values are TERRAFORM or ENV")
	updateVariableCmd.Flags().BoolVarP(&VariableUpdateSensitive, "sensitive", "s", false, "Whether the value is sensitive. If true then the variable is written once and not visible thereafter.")
	updateVariableCmd.Flags().BoolVarP(&VariableUpdateHcl, "hcl", "", false, "Whether to evaluate the value of the variable as a string of HCL code. Has no effect for environment variables.")

}

func updateVariable() {
	client := newClient()
	ctx := getContext()

	variable := &terrakube.Variable{
		ID:          VariableId,
		Key:         VariableUpdateKey,
		Value:       VariableUpdateValue,
		Description: VariableUpdateDescription,
		Sensitive:   VariableUpdateSensitive,
		Hcl:         VariableUpdateHcl,
		Category:    VariableUpdateCategory,
	}
	_, err := client.Variables.Update(ctx, VariableUpdateOrgId, VariableUpdateWorkspaceId, variable)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Updated")

}
