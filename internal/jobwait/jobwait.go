// Package jobwait polls a job until it reaches a final status.
package jobwait

import (
	"context"
	"errors"
	"fmt"
	"time"

	"terrakube/internal/exit"
)

// Final statuses, as in the server's JobStatus: the job will not change again.
var finalStatuses = map[string]int{
	"completed": exit.OK,
	"noChanges": exit.OK,
	"failed":    exit.JobFailed,
	"rejected":  exit.JobFailed,
	"cancelled": exit.JobFailed,
}

// Final reports whether status is a final job status.
func Final(status string) bool {
	_, ok := finalStatuses[status]
	return ok
}

// ExitCode returns the exit code for a final status: exit.OK for completed and
// noChanges, exit.JobFailed for failed, rejected and cancelled.
func ExitCode(status string) int {
	return finalStatuses[status]
}

// Options controls polling. Now and Sleep default to the real clock.
type Options struct {
	Interval time.Duration // between polls; default 5s
	Timeout  time.Duration // 0 waits without a limit
	Now      func() time.Time
	Sleep    func(ctx context.Context, d time.Duration) error
}

// Wait calls get until it returns a final status and returns that status. When
// the timeout passes first, it returns the last status and an *exit.Error with
// exit.Timeout.
func Wait(ctx context.Context, get func(context.Context) (string, error), opts Options) (string, error) {
	if opts.Timeout < 0 {
		return "", errors.New("--timeout must not be negative")
	}
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Sleep == nil {
		opts.Sleep = SleepContext
	}

	deadline := opts.Now().Add(opts.Timeout)
	for {
		status, err := get(ctx)
		if err != nil {
			return "", err
		}
		if Final(status) {
			return status, nil
		}
		if opts.Timeout > 0 && !opts.Now().Before(deadline) {
			return status, &exit.Error{Code: exit.Timeout, Err: fmt.Errorf("timed out after %s, job still %s", opts.Timeout, status)}
		}
		if err := opts.Sleep(ctx, opts.Interval); err != nil {
			return status, err
		}
	}
}

// SleepContext sleeps for d or until ctx is done.
func SleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
