package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var ModuleFilter string
var ModuleOrgId string
var ModuleListExample string = `List all existing modules
    %[1]v module list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb
List specific organizations applying a filter
    %[1]v module list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --filter name==mymodule `

var listModulesCmd = &cobra.Command{
	Use:   "list",
	Short: "list modules",
	Run: func(cmd *cobra.Command, args []string) {
		listModules()
	},
	Example: fmt.Sprintf(ModuleListExample, rootCmd.Use),
}

func init() {
	moduleCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().StringVarP(&ModuleFilter, "filter", "f", "", "Filter")
	listModulesCmd.Flags().StringVarP(&ModuleOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = listModulesCmd.MarkFlagRequired("organization-id")
}

func listModules() {
	client := newClient()
	ctx := getContext()

	opts := &terrakube.ListOptions{Filter: ModuleFilter}
	resp, err := client.Modules.List(ctx, ModuleOrgId, opts)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
