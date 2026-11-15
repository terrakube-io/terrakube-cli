package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
	"terrakube/internal/jobwait"
	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

// jobPollInterval is how often job wait reads the job; tests shorten it.
var jobPollInterval = 5 * time.Second

// orgScope is the organization flag of hand-written job commands.
var orgScope = []resource.ParentScope{{
	Name:      "organization",
	Flag:      "organization",
	ShortFlag: "o",
	Aliases:   []string{"org"},
	IDFlag:    "organization-id",
	Resolver:  orgResolver,
}}

func newJobWaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait <job-id>",
		Short: "wait until a job reaches a final status",
		Long: `Wait until the job is completed, noChanges, failed, rejected or cancelled, then print it.

Exit codes: 0 completed or noChanges, 1 error, 2 failed, rejected or cancelled, 3 timeout.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout < 0 {
				return errors.New("--timeout must not be negative")
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			ctx := getContext()
			ids, err := resource.ResolveScope(ctx, client, cmd, orgScope)
			if err != nil {
				return err
			}
			return waitForJob(ctx, client, ids[0], args[0], timeout)
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	cmd.Flags().Duration("timeout", 0, "give up after this long, e.g. 30m (0 waits without a limit)")
	return cmd
}

// waitForJob polls the job until it is final, prints its last state and
// returns an *exit.Error for a failed, rejected or cancelled job or a timeout.
func waitForJob(ctx context.Context, c *terrakube.Client, orgID, jobID string, timeout time.Duration) error {
	var job *terrakube.Job
	status, err := jobwait.Wait(ctx, func(ctx context.Context) (string, error) {
		j, err := c.Jobs.Get(ctx, orgID, jobID)
		if err != nil {
			return "", err
		}
		job = j
		return j.Status, nil
	}, jobwait.Options{Interval: jobPollInterval, Timeout: timeout})
	if job != nil {
		if renderErr := outputpkg.Render(os.Stdout, job, output); renderErr != nil {
			return renderErr
		}
	}
	if err != nil {
		return err
	}
	if code := jobwait.ExitCode(status); code != exit.OK {
		return &exit.Error{Code: code, Err: fmt.Errorf("job %s ended %s", jobID, status)}
	}
	return nil
}
