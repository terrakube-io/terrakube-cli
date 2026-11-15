package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/jobwait"
	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

// jobAction describes a job command that sets the job's status.
type jobAction struct {
	name, sets, short string
	// allowed returns an error when the job's current status does not allow the action.
	allowed func(jobID, status string) error
}

func needsWaitingApproval(action string) func(jobID, status string) error {
	return func(jobID, status string) error {
		if status != "waitingApproval" {
			return fmt.Errorf("job %s is %s: %s needs waitingApproval", jobID, status, action)
		}
		return nil
	}
}

func notFinal(jobID, status string) error {
	if jobwait.Final(status) {
		return fmt.Errorf("job %s already ended %s", jobID, status)
	}
	return nil
}

var jobActions = []jobAction{
	{name: "approve", sets: "approved", short: "approve a job waiting for approval", allowed: needsWaitingApproval("approve")},
	{name: "reject", sets: "rejected", short: "reject (discard) a job that has not ended", allowed: notFinal},
	{name: "cancel", sets: "cancelled", short: "cancel a job that has not ended", allowed: notFinal},
}

func newJobActionCmd(a jobAction) *cobra.Command {
	cmd := &cobra.Command{
		Use:          a.name + " <job-id>",
		Short:        a.short,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			ctx := getContext()
			ids, err := resource.ResolveScope(ctx, client, cmd, orgScope)
			if err != nil {
				return err
			}
			jobID := args[0]
			job, err := client.Jobs.Get(ctx, ids[0], jobID)
			if err != nil {
				return err
			}
			if err := a.allowed(jobID, job.Status); err != nil {
				return err
			}
			updated, err := client.Jobs.UpdateFields(ctx, ids[0], jobID, terrakube.Patch{Attributes: map[string]any{"status": a.sets}})
			if err != nil {
				return err
			}
			return outputpkg.Render(os.Stdout, updated, output)
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	return cmd
}
