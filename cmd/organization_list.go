package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var OrganizationFilter string
var listOrganizationsCmd = &cobra.Command{
	Use:   "list",
	Short: "list organizations",
	Run: func(cmd *cobra.Command, args []string) {
		listOrganizations()
	},
	Example: fmt.Sprintf(OrganizationListExample, rootCmd.Use),
}

var OrganizationListExample string = `List all existing organizations
    %[1]v organization list
List specific organizations applying a filter
    %[1]v organization list --filter name==myorg `

func init() {
	organizationCmd.AddCommand(listOrganizationsCmd)
	listOrganizationsCmd.Flags().StringVarP(&OrganizationFilter, "filter", "f", "", "Filter")
}

func listOrganizations() {
	client := newClient()
	ctx := getContext()

	opts := &terrakube.ListOptions{Filter: OrganizationFilter}
	organizations, err := client.Organizations.List(ctx, opts)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(organizations, output)
}
