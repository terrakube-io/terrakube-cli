package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var TeamFilter string
var TeamOrgId string
var TeamListExample string = `List all existing teams
    %[1]v team list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb
List specific team organizations applying a filter
    %[1]v team list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb --filter name==myteam `

var listTeamsCmd = &cobra.Command{
	Use:   "list",
	Short: "list teams",
	Run: func(cmd *cobra.Command, args []string) {
		listTeams()
	},
	Example: fmt.Sprintf(TeamListExample, rootCmd.Use),
}

func init() {
	teamCmd.AddCommand(listTeamsCmd)
	listTeamsCmd.Flags().StringVarP(&TeamFilter, "filter", "f", "", "Filter")
	registerOrgFlag(listTeamsCmd, &TeamOrgId)
}

func listTeams() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, TeamOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Teams.List(ctx, orgID, &terrakube.ListOptions{Filter: TeamFilter})

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
