package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
)

// logsServer serves job 42 with steps, archived logs per step ID, an SSE
// script per step ID (one string per connection), and the job's final status.
type logsServer struct {
	mu        sync.Mutex
	steps     []*terrakube.Step
	logs      map[string]string
	streams   map[string][]string
	jobStatus string
	lastIDs   []string
	// stepNow overrides a step's status when it is read on its own (it moved on
	// after the list was read).
	stepNow map[string]string
}

func newLogsServer(t *testing.T, ls *logsServer) *logsServer {
	t.Helper()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ls.mu.Lock()
		defer ls.mu.Unlock()
		p := r.URL.Path
		switch {
		case strings.HasSuffix(p, "/job/42/step"):
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, ls.steps)
		case strings.Contains(p, "/job/42/step/") && strings.HasPrefix(p, "/api/v1/"):
			id := p[strings.LastIndex(p, "/")+1:]
			for _, s := range ls.steps {
				if s.ID == id {
					cur := *s
					if st, ok := ls.stepNow[id]; ok {
						cur.Status = st
					}
					w.Header().Set("Content-Type", "application/vnd.api+json")
					_ = jsonapi.MarshalPayload(w, &cur)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(p, "/job/42"):
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, &terrakube.Job{ID: "42", Status: ls.jobStatus})
		case strings.HasSuffix(p, "/stream"):
			id := strings.TrimSuffix(p[:len(p)-len("/stream")], "/")
			id = id[strings.LastIndex(id, "/")+1:]
			ls.lastIDs = append(ls.lastIDs, r.Header.Get("Last-Event-ID"))
			script := ls.streams[id]
			if len(script) == 0 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			body := script[0]
			ls.streams[id] = script[1:]
			w.Header().Set("Content-Type", "text/event-stream")
			// "hold:" keeps the connection open after the body, like a live stream.
			hold := strings.HasPrefix(body, "hold:")
			_, _ = fmt.Fprint(w, strings.TrimPrefix(body, "hold:"))
			if hold {
				w.(http.Flusher).Flush()
				ls.mu.Unlock()
				<-r.Context().Done()
				ls.mu.Lock()
			}
		case strings.Contains(p, "/tfoutput/v1/"):
			id := p[strings.LastIndex(p, "/")+1:]
			body, ok := ls.logs[id]
			if !ok {
				// A step that never ran (notExecuted) has no log.
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = fmt.Fprint(w, body)
		default:
			t.Errorf("unexpected request %s %s", r.Method, p)
		}
	}))
	t.Cleanup(ts.Close)
	return ls
}

func twoDoneSteps() []*terrakube.Step {
	return []*terrakube.Step{
		{ID: "s-200", Name: "Apply", StepNumber: 200, Status: "completed"},
		{ID: "s-100", Name: "Plan", StepNumber: 100, Status: "completed"},
	}
}

// AC-15: all steps, in step order, each under a heading.
func TestCmdJobLogsAllSteps(t *testing.T) {
	resetGlobalFlags()
	newLogsServer(t, &logsServer{steps: twoDoneSteps(), logs: map[string]string{"s-100": "plan out\n", "s-200": "apply out\n"}})

	out, err := executeCommand("job", "logs", "42", "-o", runOrg)
	if err != nil {
		t.Fatal(err)
	}
	want := "==> Step 100: Plan (completed)\nplan out\n==> Step 200: Apply (completed)\napply out\n"
	if out != want {
		t.Errorf("got %q\nwant %q", out, want)
	}
}

// AC-16: --step prints only that step, without a heading.
func TestCmdJobLogsOneStep(t *testing.T) {
	resetGlobalFlags()
	newLogsServer(t, &logsServer{steps: twoDoneSteps(), logs: map[string]string{"s-100": "plan out\n", "s-200": "apply out\n"}})

	out, err := executeCommand("job", "logs", "42", "-o", runOrg, "--step", "200")
	if err != nil {
		t.Fatal(err)
	}
	if out != "apply out\n" {
		t.Errorf("got %q", out)
	}
}

// AC-17: an unknown step lists the job's steps.
func TestCmdJobLogsUnknownStep(t *testing.T) {
	resetGlobalFlags()
	newLogsServer(t, &logsServer{steps: twoDoneSteps()})

	_, err := executeCommand("job", "logs", "42", "-o", runOrg, "--step", "300")
	if err == nil || err.Error() != "step 300 not found: job 42 has steps 100, 200" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC-18, AC-19: --follow streams the running step, resumes after a drop
// without repeating or skipping, and exits by the job's final status.
func TestCmdJobLogsFollow(t *testing.T) {
	for _, tt := range []struct {
		jobStatus string
		code      int
	}{{"completed", exit.OK}, {"failed", exit.JobFailed}} {
		t.Run(tt.jobStatus, func(t *testing.T) {
			resetGlobalFlags()
			fastPolling(t)
			ls := newLogsServer(t, &logsServer{
				steps: []*terrakube.Step{
					{ID: "s-100", Name: "Plan", StepNumber: 100, Status: "completed"},
					{ID: "s-200", Name: "Apply", StepNumber: 200, Status: "running"},
				},
				logs: map[string]string{"s-100": "plan out\n"},
				streams: map[string][]string{"s-200": {
					"id: 1-0\ndata: line one\n\n", // the connection drops here
					"id: 1-0\ndata: line one\n\nid: 2-0\ndata: line two\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n",
				}},
				jobStatus: tt.jobStatus,
			})

			out, err := executeCommand("job", "logs", "42", "-o", runOrg, "--follow")
			if exit.Code(err) != tt.code {
				t.Fatalf("exit %d, want %d (err %v)", exit.Code(err), tt.code, err)
			}
			want := "==> Step 100: Plan (completed)\nplan out\n==> Step 200: Apply (running)\nline one\nline two\n"
			if out != want {
				t.Errorf("got %q\nwant %q", out, want)
			}
			ls.mu.Lock()
			defer ls.mu.Unlock()
			if len(ls.lastIDs) != 2 || ls.lastIDs[0] != "" || ls.lastIDs[1] != "1-0" {
				t.Errorf("Last-Event-ID per connection = %q, want [\"\" \"1-0\"]", ls.lastIDs)
			}
		})
	}
}

// An earlier step's stream stays open until the job ends; follow moves on once
// that step is done and its stream has gone quiet.
func TestCmdJobLogsFollowMovesToTheNextStep(t *testing.T) {
	resetGlobalFlags()
	fastPolling(t)
	ls := &logsServer{
		steps: []*terrakube.Step{
			{ID: "s-100", Name: "Plan", StepNumber: 100, Status: "running"},
			{ID: "s-200", Name: "Apply", StepNumber: 200, Status: "running"},
		},
		streams: map[string][]string{
			"s-100": {"hold:id: 1-0\ndata: planning\n\n"},
			"s-200": {"id: 2-0\ndata: applying\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"},
		},
		jobStatus: "completed",
		stepNow:   map[string]string{"s-100": "completed"},
	}
	newLogsServer(t, ls)

	out, err := executeCommand("job", "logs", "42", "-o", runOrg, "--follow")
	if err != nil {
		t.Fatal(err)
	}
	want := "==> Step 100: Plan (running)\nplanning\n==> Step 200: Apply (running)\napplying\n"
	if out != want {
		t.Errorf("got %q\nwant %q", out, want)
	}
}

// A step whose stream ended before (or while) it was followed gets the rest of
// its output from the archived log, without repeating what was streamed.
func TestCmdJobLogsFollowFillsInFromTheArchive(t *testing.T) {
	resetGlobalFlags()
	fastPolling(t)
	newLogsServer(t, &logsServer{
		steps: []*terrakube.Step{
			{ID: "s-100", Name: "Plan", StepNumber: 100, Status: "running"},
			{ID: "s-200", Name: "Apply", StepNumber: 200, Status: "pending"},
		},
		logs: map[string]string{"s-100": "p1\np2\np3\n", "s-200": "a1\na2\n"},
		streams: map[string][]string{
			"s-100": {"id: 1-0\ndata: p1\n\nevent: end\ndata: {\"reason\":\"complete\"}\n\n"},
			"s-200": {"event: end\ndata: {\"reason\":\"complete\"}\n\n"},
		},
		jobStatus: "completed",
		stepNow:   map[string]string{"s-100": "completed", "s-200": "completed"},
	})

	out, err := executeCommand("job", "logs", "42", "-o", runOrg, "--follow")
	if err != nil {
		t.Fatal(err)
	}
	want := "==> Step 100: Plan (running)\np1\np2\np3\n==> Step 200: Apply (pending)\na1\na2\n"
	if out != want {
		t.Errorf("got %q\nwant %q", out, want)
	}
}
