package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
	"terrakube/internal/jobwait"
	"terrakube/internal/resource"
)

// logReconnectBackoff, when set, replaces the SDK's wait between stream
// reconnects; tests shorten it.
var logReconnectBackoff func(failures int) time.Duration

func newJobLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <job-id>",
		Short: "print a job's logs, step by step",
		Long: `Print the logs of every step of a job in step order, or of one step with --step.

With --follow, the running step's output is printed as it arrives until the job
ends; the exit code then follows the job (see: job wait).`,
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
			orgID, jobID := ids[0], args[0]
			steps, err := client.Steps.List(ctx, orgID, jobID, nil)
			if err != nil {
				return err
			}
			sort.Slice(steps, func(i, j int) bool { return steps[i].StepNumber < steps[j].StepNumber })

			headings := true
			if cmd.Flags().Changed("step") {
				n, _ := cmd.Flags().GetInt("step")
				step, err := findStep(steps, jobID, n)
				if err != nil {
					return err
				}
				steps, headings = []*terrakube.Step{step}, false
			}

			follow, _ := cmd.Flags().GetBool("follow")
			out := cmd.OutOrStdout()
			for _, step := range steps {
				if headings {
					if _, err := fmt.Fprintf(out, "==> Step %d: %s (%s)\n", step.StepNumber, step.Name, step.Status); err != nil {
						return err
					}
				}
				if follow && stepActive(step.Status) {
					err = followStep(ctx, client, orgID, jobID, step, out, cmd.ErrOrStderr())
				} else {
					err = printStepLog(ctx, client, orgID, jobID, step.ID, out)
				}
				if err != nil {
					return err
				}
			}
			if !follow {
				return nil
			}
			return jobResult(ctx, client, orgID, jobID)
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	cmd.Flags().Int("step", 0, "print only this step (its step number, e.g. 100)")
	cmd.Flags().Bool("follow", false, "stream the running step until the job ends")
	return cmd
}

func findStep(steps []*terrakube.Step, jobID string, n int) (*terrakube.Step, error) {
	numbers := make([]string, 0, len(steps))
	for _, s := range steps {
		if s.StepNumber == n {
			return s, nil
		}
		numbers = append(numbers, strconv.Itoa(s.StepNumber))
	}
	return nil, fmt.Errorf("step %d not found: job %s has steps %s", n, jobID, strings.Join(numbers, ", "))
}

// stepActive reports whether a step can still produce output.
func stepActive(status string) bool {
	switch status {
	case "pending", "waitingApproval", "approved", "queue", "running":
		return true
	}
	return false
}

func printStepLog(ctx context.Context, c *terrakube.Client, orgID, jobID, stepID string, out io.Writer) error {
	log, err := c.StepLogs.Get(ctx, orgID, jobID, stepID, nil)
	if terrakube.IsNotFound(err) {
		return nil // a step that never ran (notExecuted) has no log
	}
	if err != nil {
		return err
	}
	defer func() { _ = log.Body.Close() }()
	_, err = io.Copy(out, log.Body)
	return err
}

// followStep streams a step's output. The server ends the stream only when the
// whole job ends, so for an earlier step it stops once the step is no longer
// active and its stream has been quiet for two poll intervals. Whatever the
// stream did not deliver (a step that ended before it was followed, a tail cut
// off) is then printed from the archived log.
func followStep(ctx context.Context, c *terrakube.Client, orgID, jobID string, step *terrakube.Step, out, errOut io.Writer) error {
	streamed, err := streamStep(ctx, c, orgID, jobID, step, out, errOut)
	if err != nil {
		return err
	}
	return printArchivedRest(ctx, c, orgID, jobID, step.ID, streamed, out)
}

// streamStep follows the step's stream and returns how many lines it printed.
func streamStep(ctx context.Context, c *terrakube.Client, orgID, jobID string, step *terrakube.Step, out, errOut io.Writer) (int, error) {
	fctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var mu sync.Mutex
	lastEvent := time.Now()
	lines := 0
	count := func() int {
		mu.Lock()
		defer mu.Unlock()
		return lines
	}
	done := make(chan error, 1)
	go func() {
		opts := &terrakube.StepLogFollowOptions{Backoff: logReconnectBackoff}
		done <- c.StepLogs.Follow(fctx, orgID, jobID, step.ID, opts, func(e terrakube.StepLogEvent) error {
			if e.Gap != nil {
				_, err := fmt.Fprintf(errOut, "[some output of step %d was not kept for streaming; run 'job logs %s --step %d' once it ends]\n",
					step.StepNumber, jobID, step.StepNumber)
				return err
			}
			line := e.Data
			if !strings.HasSuffix(line, "\n") {
				line += "\n"
			}
			mu.Lock()
			lastEvent = time.Now()
			lines += strings.Count(line, "\n")
			mu.Unlock()
			_, err := io.WriteString(out, line)
			return err
		})
	}()

	for {
		select {
		case err := <-done:
			return count(), err
		case <-time.After(jobPollInterval):
		}
		current, err := c.Steps.Get(ctx, orgID, jobID, step.ID)
		if err != nil {
			return count(), err
		}
		mu.Lock()
		quiet := time.Since(lastEvent) > 2*jobPollInterval
		mu.Unlock()
		if !stepActive(current.Status) && quiet {
			cancel()
			if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
				return count(), err
			}
			return count(), nil
		}
	}
}

// printArchivedRest prints the archived log of a step from line skip on.
// Lines are counted rather than bytes: the stream and the archive can differ in
// whitespace (the stream loses one leading space of indented lines).
func printArchivedRest(ctx context.Context, c *terrakube.Client, orgID, jobID, stepID string, skip int, out io.Writer) error {
	log, err := c.StepLogs.Get(ctx, orgID, jobID, stepID, nil)
	if terrakube.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { _ = log.Body.Close() }()
	r := bufio.NewReader(log.Body)
	for n := 0; ; n++ {
		line, err := r.ReadString('\n')
		if n >= skip && line != "" {
			if _, werr := io.WriteString(out, line); werr != nil {
				return werr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// jobResult waits for the job to end (without printing it) and returns its exit
// code as an *exit.Error, or nil for completed and noChanges.
func jobResult(ctx context.Context, c *terrakube.Client, orgID, jobID string) error {
	status, err := jobwait.Wait(ctx, func(ctx context.Context) (string, error) {
		j, err := c.Jobs.Get(ctx, orgID, jobID)
		if err != nil {
			return "", err
		}
		return j.Status, nil
	}, jobwait.Options{Interval: jobPollInterval})
	if err != nil {
		return err
	}
	if code := jobwait.ExitCode(status); code != exit.OK {
		return &exit.Error{Code: code, Err: fmt.Errorf("job %s ended %s", jobID, status)}
	}
	return nil
}
