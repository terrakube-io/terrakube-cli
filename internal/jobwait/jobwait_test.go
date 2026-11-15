package jobwait

import (
	"context"
	"errors"
	"testing"
	"time"

	"terrakube/internal/exit"
)

func TestFinalAndExitCode(t *testing.T) {
	for status, want := range map[string]int{"completed": exit.OK, "noChanges": exit.OK, "failed": exit.JobFailed, "rejected": exit.JobFailed, "cancelled": exit.JobFailed} {
		if !Final(status) {
			t.Errorf("%s should be final", status)
		}
		if got := ExitCode(status); got != want {
			t.Errorf("ExitCode(%s) = %d, want %d", status, got, want)
		}
	}
	for _, status := range []string{"pending", "waitingApproval", "approved", "queue", "running", "unknown", ""} {
		if Final(status) {
			t.Errorf("%s should not be final", status)
		}
	}
}

// fakeClock advances only when Wait sleeps.
type fakeClock struct{ now time.Time }

func (c *fakeClock) options(timeout time.Duration) Options {
	return Options{
		Interval: time.Second,
		Timeout:  timeout,
		Now:      func() time.Time { return c.now },
		Sleep:    func(_ context.Context, d time.Duration) error { c.now = c.now.Add(d); return nil },
	}
}

func statuses(list ...string) (func(context.Context) (string, error), *int) {
	calls := 0
	return func(context.Context) (string, error) {
		s := list[min(calls, len(list)-1)]
		calls++
		return s, nil
	}, &calls
}

func TestWait_ReturnsTheFinalStatus(t *testing.T) {
	get, calls := statuses("queue", "running", "failed")
	clock := &fakeClock{now: time.Unix(0, 0)}
	status, err := Wait(context.Background(), get, clock.options(0))
	if err != nil || status != "failed" || *calls != 3 {
		t.Fatalf("status=%q err=%v calls=%d", status, err, *calls)
	}
}

func TestWait_Timeout(t *testing.T) {
	get, _ := statuses("running")
	clock := &fakeClock{now: time.Unix(0, 0)}
	status, err := Wait(context.Background(), get, clock.options(3*time.Second))
	if status != "running" || exit.Code(err) != exit.Timeout {
		t.Fatalf("status=%q err=%v code=%d", status, err, exit.Code(err))
	}
	if err.Error() != "timed out after 3s, job still running" {
		t.Errorf("message: %q", err.Error())
	}
}

func TestWait_FinalWithinTimeout(t *testing.T) {
	get, _ := statuses("running", "running", "completed")
	clock := &fakeClock{now: time.Unix(0, 0)}
	if status, err := Wait(context.Background(), get, clock.options(5*time.Second)); err != nil || status != "completed" {
		t.Fatalf("status=%q err=%v", status, err)
	}
}

func TestWait_Errors(t *testing.T) {
	boom := errors.New("boom")
	clock := &fakeClock{now: time.Unix(0, 0)}
	if _, err := Wait(context.Background(), func(context.Context) (string, error) { return "", boom }, clock.options(0)); !errors.Is(err, boom) {
		t.Errorf("get error: %v", err)
	}
	opts := clock.options(0)
	opts.Sleep = func(context.Context, time.Duration) error { return context.Canceled }
	get, _ := statuses("running")
	if _, err := Wait(context.Background(), get, opts); !errors.Is(err, context.Canceled) {
		t.Errorf("sleep error: %v", err)
	}
	if _, err := Wait(context.Background(), get, Options{Timeout: -time.Second}); err == nil || err.Error() != "--timeout must not be negative" {
		t.Errorf("negative timeout: %v", err)
	}
}

func TestSleepContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SleepContext(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled sleep: %v", err)
	}
	if err := SleepContext(context.Background(), time.Millisecond); err != nil {
		t.Errorf("short sleep: %v", err)
	}
}

func TestWait_DefaultsToTheRealClock(t *testing.T) {
	get, calls := statuses("completed")
	if status, err := Wait(context.Background(), get, Options{}); err != nil || status != "completed" || *calls != 1 {
		t.Fatalf("status=%q err=%v calls=%d", status, err, *calls)
	}
}
