package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ModuleDeleteExample string = `Delete a module
    %[1]v module delete --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd `

var ModuleDeleteId string
var ModuleDeleteOrgId string

var deleteModuleCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a module",
	Run: func(cmd *cobra.Command, args []string) {
		deleteModule()
	},
	Example: fmt.Sprintf(ModuleDeleteExample, rootCmd.Use),
}

func init() {
	moduleCmd.AddCommand(deleteModuleCmd)
	deleteModuleCmd.Flags().StringVarP(&ModuleDeleteOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = deleteModuleCmd.MarkFlagRequired("organization-id")
	deleteModuleCmd.Flags().StringVarP(&ModuleDeleteId, "id", "", "", "Id of the module (required)")
	_ = deleteModuleCmd.MarkFlagRequired("id")
}

func deleteModule() {
	client := newTerrakubeClient()
	ctx := getContext()

	err := client.Modules.Delete(ctx, ModuleDeleteOrgId, ModuleDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("deleted")
}
