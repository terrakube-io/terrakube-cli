package cmd

import (
	"fmt"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

var JobFilter string
var JobOrgId string
var JobListExample string = `List all existing jobs
    %[1]v job list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb
List specific jobs applying a filter
    %[1]v job list --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb --filter id==jobid `

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
	listJobsCmd.Flags().StringVarP(&JobOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = listJobsCmd.MarkFlagRequired("organization-id")
}

func listJobs() {
	client := newTerrakubeClient()
	ctx := getContext()

	resp, err := client.Jobs.List(ctx, JobOrgId, &terrakube.ListOptions{Filter: JobFilter})

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
