// Package exit maps command errors to process exit codes.
package exit

import "errors"

// Exit codes. Every error without a code exits with Failure.
const (
	OK        = 0 // success; for a job: completed or noChanges
	Failure   = 1 // CLI or API error
	JobFailed = 2 // the job ended failed, rejected or cancelled
	Timeout   = 3 // the wait timed out before the job ended
)

// Error is an error that ends the process with Code.
type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }

func (e *Error) Unwrap() error { return e.Err }

// Code returns the exit code for err: OK for nil, the code of a wrapped
// *Error, and Failure otherwise.
func Code(err error) int {
	if err == nil {
		return OK
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Failure
}
