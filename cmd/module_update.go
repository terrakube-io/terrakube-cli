package cmd

import (
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

var ModuleUpdateExample string = `Update the description of the module using id
    %[1]v module update --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --id 38b6635a-d38e-46f2-a95e-d00a416de4fd -d "new description" `

var ModuleId string
var ModuleUpdateDescription string
var ModuleUpdateName string
var ModuleUpdateOrgId string
var ModuleUpdateSource string
var ModuleUpdateProvider string
var ModuleUpdateTagPrefix string
var ModuleUpdateFolder string

var updateModuleCmd = &cobra.Command{
	Use:   "update",
	Short: "update a module",
	Run: func(cmd *cobra.Command, args []string) {
		updateModule()
	},
	Example: fmt.Sprintf(ModuleUpdateExample, rootCmd.Use),
}

func init() {
	moduleCmd.AddCommand(updateModuleCmd)
	updateModuleCmd.Flags().StringVarP(&ModuleId, "id", "", "", "Id of the module (required)")
	_ = updateModuleCmd.MarkFlagRequired("id")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = updateModuleCmd.MarkFlagRequired("organization-id")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateName, "name", "n", "", "Name of the module")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateDescription, "description", "d", "", "Description of the module")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateSource, "source", "s", "", "Source of the module")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateProvider, "provider", "p", "", "Provider of the module")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateTagPrefix, "tag-prefix", "t", "", "Tag prefix of the module")
	updateModuleCmd.Flags().StringVarP(&ModuleUpdateFolder, "folder", "f", "", "Folder of the module")
}

func updateModule() {
	client := newTerrakubeClient()
	ctx := getContext()

	module := &terrakube.Module{
		ID:          ModuleId,
		Name:        ModuleUpdateName,
		Description: ModuleUpdateDescription,
		Source:      ModuleUpdateSource,
		Provider:    ModuleUpdateProvider,
		TagPrefix:   ptrOrNil(ModuleUpdateTagPrefix),
		Folder:      ptrOrNil(ModuleUpdateFolder),
	}
	_, err := client.Modules.Update(ctx, ModuleUpdateOrgId, module)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Updated")
}
