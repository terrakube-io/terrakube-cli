package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var JobFilter string
var JobOrgId string
var JobListExample string = `List all existing jobs
    %[1]v job list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb
List specific jobs applying a filter
    %[1]v job list -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb --filter id==jobid `

var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "list jobs",
	Run: func(cmd *cobra.Command, args []string) {
		listJobs()
	},
	Example: fmt.Sprintf(JobListExample, rootCmd.Use),
}

func init() {
	jobCmd.AddCommand(listJobsCmd)
	listJobsCmd.Flags().StringVarP(&JobFilter, "filter", "f", "", "Filter")
	registerOrgFlag(listJobsCmd, &JobOrgId)
}

func listJobs() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, JobOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Jobs.List(ctx, orgID, &terrakube.ListOptions{Filter: JobFilter})

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
