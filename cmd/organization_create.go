package cmd

import (
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

var OrganizationCreateExample string = `Create a new organization
    %[1]v organization create -n myorg -d "org description" `

var OrganizationCreateName string
var OrganizationCreateDescription string
var OrganizationCreateExecutionMode string
var OrganizationCreateIcon string
var createOrganizationCmd = &cobra.Command{
	Use:   "create",
	Short: "create an organization",
	Run: func(cmd *cobra.Command, args []string) {
		createOrganization()
	},
	Example: fmt.Sprintf(OrganizationCreateExample, rootCmd.Use),
}

func init() {
	organizationCmd.AddCommand(createOrganizationCmd)
	createOrganizationCmd.Flags().StringVarP(&OrganizationCreateName, "name", "n", "", "Name of the new organization (required)")
	_ = createOrganizationCmd.MarkFlagRequired("name")
	createOrganizationCmd.Flags().StringVarP(&OrganizationCreateDescription, "description", "d", "", "Description of the new organization")
	createOrganizationCmd.Flags().StringVarP(&OrganizationCreateExecutionMode, "executionMode", "e", "", "Default execution mode for the organization")
	createOrganizationCmd.Flags().StringVarP(&OrganizationCreateIcon, "icon", "i", "", "Organization icon")

	// Add validation for execution mode
	createOrganizationCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if OrganizationCreateExecutionMode != "" && OrganizationCreateExecutionMode != "remote" && OrganizationCreateExecutionMode != "local" {
			return fmt.Errorf("executionMode must be either 'remote' or 'local'")
		}
		return nil
	}

}

func createOrganization() {
	client := newTerrakubeClient()
	ctx := getContext()

	organization := &terrakube.Organization{
		Name:          OrganizationCreateName,
		Description:   ptrOrNil(OrganizationCreateDescription),
		ExecutionMode: OrganizationCreateExecutionMode,
		Icon:          ptrOrNil(OrganizationCreateIcon),
	}
	resp, err := client.Organizations.Create(ctx, organization)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
