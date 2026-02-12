package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var JobCreateExample string = `Create a new job
    %[1]v job create --organization-id e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w e5ad0642-f9b3-48b3-9bf4-35997febe1fb  -c apply`

var JobCreateWorkspaceId string
var JobCreateCommand string
var JobCreateOrgId string

var createJobCmd = &cobra.Command{
	Use:   "create",
	Short: "create a job",
	Run: func(cmd *cobra.Command, args []string) {
		createJob()
	},
	Example: fmt.Sprintf(JobCreateExample, rootCmd.Use),
}

func init() {
	jobCmd.AddCommand(createJobCmd)
	createJobCmd.Flags().StringVarP(&JobCreateCommand, "command", "c", "", "Command to execute: plan,apply,destroy (required)")
	_ = createJobCmd.MarkFlagRequired("command")
	createJobCmd.Flags().StringVarP(&JobCreateOrgId, "organization-id", "", "", "Organization Id (required)")
	_ = createJobCmd.MarkFlagRequired("organization-id")
	createJobCmd.Flags().StringVarP(&JobCreateWorkspaceId, "workspace-id", "w", "", "Workspace Id (required)")
	_ = createJobCmd.MarkFlagRequired("workspace-id")
}

func createJob() {
	client := newClient()
	ctx := getContext()

	job := &terrakube.Job{
		Command:   JobCreateCommand,
		Workspace: &terrakube.Workspace{ID: JobCreateWorkspaceId},
	}

	resp, err := client.Jobs.Create(ctx, JobCreateOrgId, job)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
