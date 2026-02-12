package cmd

import (
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
)

var JobCreateExample string = `Create a new job
    %[1]v job create -o e5ad0642-f9b3-48b3-9bf4-35997febe1fb -w e5ad0642-f9b3-48b3-9bf4-35997febe1fb -c apply`

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
	registerOrgFlag(createJobCmd, &JobCreateOrgId)
	registerWsFlag(createJobCmd, &JobCreateWorkspaceId)
}

func createJob() {
	client := newClient()
	ctx := getContext()
	orgID, err := resolveOrg(ctx, client, JobCreateOrgId)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsID, err := resolveWs(ctx, client, orgID, JobCreateWorkspaceId)
	if err != nil {
		fmt.Println(err)
		return
	}

	job := &terrakube.Job{
		Command:   JobCreateCommand,
		Workspace: &terrakube.Workspace{ID: wsID},
	}

	resp, err := client.Jobs.Create(ctx, orgID, job)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderOutput(resp, output)
}
