package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"testing/iotest"
	"time"

	"github.com/google/jsonapi"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// stateAPI fakes the workspace, history and /tfstate/v1 endpoints of one
// workspace (polWs, named "prod", in polOrg).
type stateAPI struct {
	mu      sync.Mutex
	url     string
	current []byte               // current raw state; nil serves an empty body
	files   map[string][]byte    // state files by name, e.g. "v1.raw.json"
	history []*terrakube.History // Output may be relative to the server URL
	calls   []string             // "METHOD path" of every request
}

const stateDir = "/tfstate/v1/organization/" + polOrg + "/workspace/" + polWs + "/"

func newStateAPI(t *testing.T) *stateAPI {
	t.Helper()
	s := &stateAPI{files: map[string][]byte{}}
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls = append(s.calls, r.Method+" "+r.URL.Path)
		apiWs := "/api/v1/organization/" + polOrg + "/workspace"
		switch p := r.URL.Path; {
		case r.Method == http.MethodGet && p == apiWs:
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, []*terrakube.Workspace{{ID: polWs, Name: "prod"}})
		case r.Method == http.MethodGet && p == apiWs+"/"+polWs+"/history":
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, s.absHistory())
		case r.Method == http.MethodGet && strings.HasPrefix(p, apiWs+"/"+polWs+"/history/"):
			for _, h := range s.absHistory() {
				if h.ID == path.Base(p) {
					w.Header().Set("Content-Type", "application/vnd.api+json")
					_ = jsonapi.MarshalPayload(w, h)
					return
				}
			}
			http.Error(w, `{"errors":[{"detail":"not found"}]}`, http.StatusNotFound)
		case r.Method == http.MethodGet && p == stateDir+"state/terraform.tfstate":
			_, _ = w.Write(s.current)
		case r.Method == http.MethodGet && strings.HasPrefix(p, stateDir+"state/"):
			b, ok := s.files[path.Base(p)]
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			_, _ = w.Write(b)
		case r.Method == http.MethodPut && strings.HasPrefix(p, stateDir+"rollback/"):
			_, _ = w.Write([]byte("rolled back"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, p)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)
	s.url = ts.URL
	return s
}

// absHistory returns the history with relative Output paths made absolute.
func (s *stateAPI) absHistory() []*terrakube.History {
	out := make([]*terrakube.History, 0, len(s.history))
	for _, h := range s.history {
		c := *h
		if strings.HasPrefix(c.Output, "/") {
			c.Output = s.url + c.Output
		}
		out = append(out, &c)
	}
	return out
}

func (s *stateAPI) addVersion(id, name string, created time.Time, raw, showJSON string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, &terrakube.History{ID: id, Output: stateDir + "state/" + name + ".json", CreatedDate: &created})
	s.files[name+".raw.json"] = []byte(raw)
	if showJSON != "" {
		s.files[name+".json"] = []byte(showJSON)
	}
}

func (s *stateAPI) requests() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.calls...)
}

// rawState has the odd spacing, tabs and non-ASCII text a re-encoder would change.
const rawState = "{\n  \"version\": 4,\n\t\"serial\": 7,  \"lineage\": \"é-1\"\n}"

// AC-37: state pull writes the current state byte for byte, by workspace name.
func TestCmdStatePullCurrent(t *testing.T) {
	resetGlobalFlags()
	s := newStateAPI(t)
	s.current = []byte(rawState)
	out, err := executeCommand("state", "pull", "prod", "-o", polOrg)
	if err != nil {
		t.Fatal(err)
	}
	if out != rawState {
		t.Errorf("stdout %q, want %q", out, rawState)
	}
}

// AC-37: --version writes that history version's raw state to the -f file (mode 0600).
func TestCmdStatePullVersionToFile(t *testing.T) {
	resetGlobalFlags()
	s := newStateAPI(t)
	s.addVersion("h1", "8f1c", time.Now(), rawState, `{"values":{}}`)
	file := t.TempDir() + "/prod.tfstate"
	out, err := executeCommand("state", "pull", polWs, "-o", polOrg, "--version", "h1", "-f", file)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(file)
	if string(got) != rawState || out != "" {
		t.Errorf("file %q, stdout %q", got, out)
	}
	if fi, _ := os.Stat(file); fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode %v, want 0600", fi.Mode().Perm())
	}
	if reqs := s.requests(); reqs[len(reqs)-1] != "GET "+stateDir+"state/8f1c.raw.json" {
		t.Errorf("requests %v", reqs)
	}
}

// AC-41: a history URL on another host gets no request (so no token) and the
// command fails with exit code 1; no file is left behind.
func TestCmdStatePullForeignHost(t *testing.T) {
	resetGlobalFlags()
	var foreign int
	evil := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { foreign++ }))
	defer evil.Close()
	s := newStateAPI(t)
	s.history = []*terrakube.History{{ID: "h1", Output: evil.URL + "/bucket/8f1c.json"}}
	file := t.TempDir() + "/out.tfstate"

	_, err := executeCommand("state", "pull", "prod", "-o", polOrg, "--version", "h1", "-f", file)
	if !errors.Is(err, terrakube.ErrForeignHost) || !strings.Contains(err.Error(), evil.URL) {
		t.Fatalf("expected a foreign host error naming %s, got %v", evil.URL, err)
	}
	if foreign != 0 {
		t.Errorf("the foreign host got %d requests", foreign)
	}
	if _, statErr := os.Stat(file); !os.IsNotExist(statErr) {
		t.Errorf("no file should be created, stat: %v", statErr)
	}
}

func TestCmdStatePullErrors(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"no state":        {[]string{"prod"}, "workspace prod has no state"},
		"unknown version": {[]string{"prod", "--version", "nope"}, "not found"},
		"empty version":   {[]string{"prod", "--version", "h1"}, "state version h1 of workspace prod has no state file"},
		"unknown ws":      {[]string{"other"}, `workspace "other" not found`},
		"unwritable file": {[]string{"prod", "-f", "/nonexistent-dir/x"}, "/nonexistent-dir/x"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resetGlobalFlags()
			s := newStateAPI(t)
			s.addVersion("h1", "8f1c", time.Now(), "", "")
			if name == "unwritable file" {
				s.current = []byte(rawState)
			}
			_, err := executeCommand(append([]string{"state", "pull", "-o", polOrg}, c.args...)...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error containing %q, got %v", c.want, err)
			}
		})
	}
}

// A copy that fails halfway leaves no partial state file.
func TestWriteStateFileRemovesPartialFile(t *testing.T) {
	file := t.TempDir() + "/out.tfstate"
	if err := writeStateFile(file, iotest.ErrReader(errors.New("connection reset"))); err == nil || err.Error() != "connection reset" {
		t.Fatalf("expected the read error, got %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("partial file left behind: %v", err)
	}
}

const outputsState = `{"values":{"outputs":{
	"vpc_id":{"value":"vpc-1","type":"string","sensitive":false},
	"db_password":{"value":"hunter2","type":"string","sensitive":true},
	"big":{"value":12345678901234567890,"type":"number","sensitive":false},
	"subnets":{"value":["a","b"],"type":["list","string"],"sensitive":false},
	"sizes":{"value":{"n":[1.5e3,2]},"type":"map","sensitive":false}}}}`

// AC-38: outputs come from the newest state version; sensitive values print
// as (sensitive) unless --show-sensitive.
func TestCmdStateOutputs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
		not  []string
	}{
		{
			name: "json hides sensitive",
			want: []string{`"db_password": {`, `"value": "(sensitive)"`, `"sensitive": true`, `"value": "vpc-1"`, `12345678901234567890`},
			not:  []string{"hunter2"},
		},
		{name: "json shows sensitive", args: []string{"--show-sensitive"}, want: []string{`"value": "hunter2"`}},
		{
			name: "tsv",
			args: []string{"--output", "tsv"},
			want: []string{"big\t12345678901234567890\tfalse\n", "db_password\t(sensitive)\ttrue\n", "subnets\t[\"a\",\"b\"]\tfalse\n", "vpc_id\tvpc-1\tfalse\n"},
			not:  []string{"hunter2"},
		},
		{name: "table", args: []string{"--output", "table"}, want: []string{"NAME", "VALUE", "SENSITIVE", "(sensitive)", "vpc-1"}, not: []string{"hunter2"}},
		{name: "none", args: []string{"--output", "none"}, not: []string{"vpc"}},
		{
			name: "yaml keeps numbers",
			args: []string{"--output", "yaml"},
			want: []string{"value: 12345678901234567890\n", "value: (sensitive)\n", "- 1.5e3\n", "- 2\n"},
			not:  []string{`"12345678901234567890"`, "hunter2"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetGlobalFlags()
			s := newStateAPI(t)
			s.addVersion("h1", "old", time.Now().Add(-time.Hour), "{}", `{"values":{"outputs":{"vpc_id":{"value":"stale"}}}}`)
			s.addVersion("h2", "new", time.Now(), "{}", outputsState)
			out, err := executeCommand(append([]string{"state", "outputs", "prod", "-o", polOrg}, c.args...)...)
			if err != nil {
				t.Fatal(err)
			}
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("output lacks %q:\n%s", w, out)
				}
			}
			for _, n := range append(c.not, "stale") {
				if strings.Contains(out, n) {
					t.Errorf("output contains %q:\n%s", n, out)
				}
			}
		})
	}
}

// AC-38: a workspace without state has no outputs.
func TestCmdStateOutputsEmpty(t *testing.T) {
	resetGlobalFlags()
	newStateAPI(t)
	out, err := executeCommand("state", "outputs", polWs, "-o", polOrg)
	if err != nil || strings.TrimSpace(out) != "{}" {
		t.Errorf("out %q, err %v", out, err)
	}
	if _, err := executeCommand("state", "outputs", "nope", "-o", polOrg); err == nil {
		t.Error("expected an error for an unknown workspace")
	}
}
