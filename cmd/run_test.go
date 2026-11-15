package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
)

const (
	runOrg  = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	runWs   = "d4e5f6a7-b8c9-0123-defa-234567890123"
	runTpl  = "c3d4e5f6-a7b8-4901-9cde-f12345678901"
	runTpl2 = "c3d4e5f6-a7b8-4901-9cde-f12345678902"
)

type runServer struct {
	mu          sync.Mutex
	defaultTpl  string
	jobStatuses []string
	posts       []map[string]interface{}
	gets        int
}

// newRunServer serves the workspace (with defaultTpl), job creation (id 42,
// pending) and job reads (jobStatuses in order).
func newRunServer(t *testing.T, defaultTpl string, jobStatuses ...string) *runServer {
	t.Helper()
	rs := &runServer{defaultTpl: defaultTpl, jobStatuses: jobStatuses}
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.mu.Lock()
		defer rs.mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/workspace/"+runWs):
			_ = jsonapi.MarshalPayload(w, &terrakube.Workspace{ID: runWs, Name: "prod", TemplateID: rs.defaultTpl})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/organization/"+runOrg+"/job"):
			var payload struct {
				Data map[string]interface{} `json:"data"`
			}
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &payload)
			rs.posts = append(rs.posts, payload.Data)
			w.WriteHeader(http.StatusCreated)
			_ = jsonapi.MarshalPayload(w, &terrakube.Job{ID: "42", Status: "pending"})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/job/42"):
			s := rs.jobStatuses[min(rs.gets, len(rs.jobStatuses)-1)]
			rs.gets++
			_ = jsonapi.MarshalPayload(w, &terrakube.Job{ID: "42", Status: s})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)
	return rs
}

func (rs *runServer) onlyPost(t *testing.T) (attrs, rels map[string]interface{}) {
	t.Helper()
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if len(rs.posts) != 1 {
		t.Fatalf("expected one job created, got %d", len(rs.posts))
	}
	attrs, _ = rs.posts[0]["attributes"].(map[string]interface{})
	rels, _ = rs.posts[0]["relationships"].(map[string]interface{})
	return attrs, rels
}

// AC-24: run apply without -t uses the workspace's default template.
func TestCmdRunApplyUsesDefaultTemplate(t *testing.T) {
	resetGlobalFlags()
	rs := newRunServer(t, runTpl, "pending")

	out, err := executeCommand("run", "apply", "-o", runOrg, "-w", runWs)
	if err != nil {
		t.Fatal(err)
	}
	attrs, rels := rs.onlyPost(t)
	if attrs["templateReference"] != runTpl {
		t.Errorf("templateReference = %v", attrs["templateReference"])
	}
	ws, _ := rels["workspace"].(map[string]interface{})
	if d, _ := ws["data"].(map[string]interface{}); d["id"] != runWs || d["type"] != "workspace" {
		t.Errorf("workspace relationship = %v", rels)
	}
	if !strings.Contains(out, `"id": "42"`) {
		t.Errorf("expected the created job printed, got %q", out)
	}
}

// AC-25: plan and destroy need a template; nothing is created without one.
func TestCmdRunNeedsATemplate(t *testing.T) {
	for _, kind := range []string{"plan", "destroy"} {
		t.Run(kind, func(t *testing.T) {
			resetGlobalFlags()
			rs := newRunServer(t, runTpl, "pending")
			_, err := executeCommand("run", kind, "-o", runOrg, "-w", runWs)
			want := "run " + kind + " needs a template: pass -t <name|ID> or set TERRAKUBE_TEMPLATE_" + strings.ToUpper(kind)
			if err == nil || err.Error() != want {
				t.Fatalf("expected %q, got %v", want, err)
			}
			if len(rs.posts) != 0 {
				t.Errorf("expected no job created, got %d", len(rs.posts))
			}
		})
	}
}

// AC-25: apply without -t and without a default template fails, nothing created.
func TestCmdRunApplyWithoutDefaultTemplate(t *testing.T) {
	resetGlobalFlags()
	rs := newRunServer(t, "", "pending")
	_, err := executeCommand("run", "apply", "-o", runOrg, "-w", runWs)
	want := `workspace "prod" has no default template: pass -t <name|ID> or set TERRAKUBE_TEMPLATE_APPLY`
	if err == nil || err.Error() != want {
		t.Fatalf("expected %q, got %v", want, err)
	}
	if len(rs.posts) != 0 {
		t.Errorf("expected no job created, got %d", len(rs.posts))
	}
}

// The env template applies when -t is not given; -t wins over it.
func TestCmdRunTemplateFromEnv(t *testing.T) {
	resetGlobalFlags()
	t.Setenv("TERRAKUBE_TEMPLATE_DESTROY", runTpl)
	rs := newRunServer(t, "", "pending")
	if _, err := executeCommand("run", "destroy", "-o", runOrg, "-w", runWs); err != nil {
		t.Fatal(err)
	}
	if attrs, _ := rs.onlyPost(t); attrs["templateReference"] != runTpl {
		t.Errorf("env template not used: %v", attrs)
	}

	resetGlobalFlags()
	rs = newRunServer(t, "", "pending")
	if _, err := executeCommand("run", "destroy", "-o", runOrg, "-w", runWs, "-t", runTpl2); err != nil {
		t.Fatal(err)
	}
	if attrs, _ := rs.onlyPost(t); attrs["templateReference"] != runTpl2 {
		t.Errorf("-t did not win over the env: %v", attrs)
	}
}

// AC-26: --target, --replace and --branch reach the job.
func TestCmdRunTargetReplaceBranch(t *testing.T) {
	resetGlobalFlags()
	rs := newRunServer(t, "", "pending")
	if _, err := executeCommand("run", "plan", "-o", runOrg, "-w", runWs, "-t", runTpl,
		"--target", "module.a", "--target", "aws_s3_bucket.b", "--replace", "aws_instance.c", "--branch", "dev"); err != nil {
		t.Fatal(err)
	}
	attrs, _ := rs.onlyPost(t)
	b, _ := json.Marshal(map[string]interface{}{"t": attrs["targetAddrs"], "r": attrs["replaceAddrs"], "b": attrs["overrideBranch"]})
	if string(b) != `{"b":"dev","r":["aws_instance.c"],"t":["module.a","aws_s3_bucket.b"]}` {
		t.Errorf("got %s", b)
	}
}

// AC-23: --wait waits for the job and exits by its status.
func TestCmdRunWait(t *testing.T) {
	resetGlobalFlags()
	fastPolling(t)
	newRunServer(t, runTpl, "running", "failed")

	out, err := executeCommand("run", "apply", "-o", runOrg, "-w", runWs, "--wait")
	if exit.Code(err) != exit.JobFailed {
		t.Fatalf("expected exit %d, got %v", exit.JobFailed, err)
	}
	if strings.Count(out, `"id": "42"`) != 1 || !strings.Contains(out, `"status": "failed"`) {
		t.Errorf("expected only the final job printed, got %q", out)
	}
}

// AC-23: job create --wait behaves like job wait on the new job.
func TestCmdJobCreateWait(t *testing.T) {
	resetGlobalFlags()
	fastPolling(t)
	newRunServer(t, "", "queue", "completed")

	out, err := executeCommand("job", "create", "-o", runOrg, "-w", runWs, "-t", runTpl, "--wait", "--timeout", "1m")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, `"id": "42"`) != 1 || !strings.Contains(out, `"status": "completed"`) {
		t.Errorf("expected only the final job printed, got %q", out)
	}
}
