package exit

import (
	"errors"
	"fmt"
	"testing"
)

func TestCode(t *testing.T) {
	coded := &Error{Code: JobFailed, Err: errors.New("job 7 failed")}
	for _, tt := range []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, OK},
		{"plain error", errors.New("boom"), Failure},
		{"coded", coded, JobFailed},
		{"wrapped", fmt.Errorf("wait: %w", coded), JobFailed},
	} {
		if got := Code(tt.err); got != tt.want {
			t.Errorf("%s: Code = %d, want %d", tt.name, got, tt.want)
		}
	}
	if coded.Error() != "job 7 failed" || !errors.Is(coded, coded.Err) {
		t.Errorf("Error()=%q, unwrap broken", coded.Error())
	}
}
