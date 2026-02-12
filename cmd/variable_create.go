package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var VariableCreateExample string = `Create a new variable
    %[1]v workspace variable create --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w 38b6635a-d38e-46f2-a95e-d00a416de4fd -k tag_name -v "Hola mundo" --hcl=false --sensitive=false --category TERRAFORM `

var VariableCreateKey string
var VariableCreateValue string
var VariableCreateDescription string
var VariableCreateCategory string
var VariableCreateSensitive bool
var VariableCreateHcl bool
var VariableCreateOrgId string
var VariableCreateWorkspaceId string

var createVariableCmd = &cobra.Command{
	Use:   "create",
	Short: "create a variable",
	Run: func(cmd *cobra.Command, args []string) {
		createVariable()
	},
	Example: fmt.Sprintf(VariableCreateExample, rootCmd.Use),
}

func init() {
	variableCmd.AddCommand(createVariableCmd)
	createVariableCmd.Flags().StringVarP(&VariableCreateKey, "key", "k", "", "Key of the new variable (required)")
	_ = createVariableCmd.MarkFlagRequired("key")
	createVariableCmd.Flags().StringVarP(&VariableCreateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = createVariableCmd.MarkFlagRequired("organization-id")
	createVariableCmd.Flags().StringVarP(&VariableCreateValue, "value", "v", "", "Value for the new variable")
	_ = createVariableCmd.MarkFlagRequired("value")
	createVariableCmd.Flags().StringVarP(&VariableCreateDescription, "description", "d", "", "Description of the new variable")
	createVariableCmd.Flags().StringVarP(&VariableCreateCategory, "category", "c", "", "Category of the new variable. Valid values are TERRAFORM or ENV")
	_ = createVariableCmd.MarkFlagRequired("category")
	createVariableCmd.Flags().BoolVarP(&VariableCreateSensitive, "sensitive", "s", false, "Whether the value is sensitive. If true then the variable is written once and not visible thereafter.")
	_ = createVariableCmd.MarkFlagRequired("sensitive")
	createVariableCmd.Flags().BoolVarP(&VariableCreateHcl, "hcl", "", false, "Whether to evaluate the value of the variable as a string of HCL code. Has no effect for environment variables.")
	_ = createVariableCmd.MarkFlagRequired("hcl")
	createVariableCmd.Flags().StringVarP(&VariableCreateWorkspaceId, "workspace-id", "w", "", "Workspace Id (required)")
	_ = createVariableCmd.MarkFlagRequired("workspace-id")
}

func createVariable() {
	client := newClient()
	ctx := getContext()

	variable := &terrakube.Variable{
		Key:         VariableCreateKey,
		Value:       VariableCreateValue,
		Description: VariableCreateDescription,
		Sensitive:   VariableCreateSensitive,
		Hcl:         VariableCreateHcl,
		Category:    VariableCreateCategory,
	}

	resp, err := client.Variables.Create(ctx, VariableCreateOrgId, VariableCreateWorkspaceId, variable)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
