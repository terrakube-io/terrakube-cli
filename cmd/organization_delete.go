package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var OrganizationDeleteId string
var OrganizationDeleteExample string = `Delete an organization using id
    %[1]v organization delete --id 38b6635a-d38e-46f2-a95e-d00a416de4fd`

var deleteOrganizationCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete an organization",
	Run: func(cmd *cobra.Command, args []string) {
		deleteOrganization()
	},
	Example: fmt.Sprintf(OrganizationDeleteExample, rootCmd.Use),
}

func init() {
	organizationCmd.AddCommand(deleteOrganizationCmd)
	deleteOrganizationCmd.Flags().StringVarP(&OrganizationDeleteId, "id", "", "", "Id of the organization (required)")
	_ = deleteOrganizationCmd.MarkFlagRequired("id")
}

func deleteOrganization() {
	client := newTerrakubeClient()
	ctx := getContext()

	err := client.Organizations.Delete(ctx, OrganizationDeleteId)

	if err != nil {
		fmt.Println(err)
		return
	}
}
