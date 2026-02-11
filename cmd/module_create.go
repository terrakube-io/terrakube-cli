package cmd

import (
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

var ModuleCreateExample string = `Create a new module
    %[1]v module create --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -n myModule -d "module description" -p azurerm -s https://github.com/terrakube-io/terraform-sample-repository.git `

var ModuleCreateName string
var ModuleCreateDescription string
var ModuleCreateOrgId string
var ModuleCreateSource string
var ModuleCreateProvider string
var ModuleCreateTagPrefix string
var ModuleCreateFolder string

var createModuleCmd = &cobra.Command{
	Use:   "create",
	Short: "create a module",
	Run: func(cmd *cobra.Command, args []string) {
		createModule()
	},
	Example: fmt.Sprintf(ModuleCreateExample, rootCmd.Use),
}

func init() {
	moduleCmd.AddCommand(createModuleCmd)
	createModuleCmd.Flags().StringVarP(&ModuleCreateName, "name", "n", "", "Name of the new module (required)")
	_ = createModuleCmd.MarkFlagRequired("name")
	createModuleCmd.Flags().StringVarP(&ModuleCreateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = createModuleCmd.MarkFlagRequired("organization-id")
	createModuleCmd.Flags().StringVarP(&ModuleCreateDescription, "description", "d", "", "Description of the new module")
	createModuleCmd.Flags().StringVarP(&ModuleCreateSource, "source", "s", "", "Source of the new module")
	createModuleCmd.Flags().StringVarP(&ModuleCreateProvider, "provider", "p", "", "Provider of the new module")
	createModuleCmd.Flags().StringVarP(&ModuleCreateTagPrefix, "tag-prefix", "t", "", "Tag prefix of the new module")
	createModuleCmd.Flags().StringVarP(&ModuleCreateFolder, "folder", "f", "", "Folder of the new module")

}

func createModule() {
	client := newTerrakubeClient()
	ctx := getContext()

	module := &terrakube.Module{
		Name:        ModuleCreateName,
		Description: ModuleCreateDescription,
		Source:      ModuleCreateSource,
		Provider:    ModuleCreateProvider,
		TagPrefix:   ptrOrNil(ModuleCreateTagPrefix),
		Folder:      ptrOrNil(ModuleCreateFolder),
	}

	resp, err := client.Modules.Create(ctx, ModuleCreateOrgId, module)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
