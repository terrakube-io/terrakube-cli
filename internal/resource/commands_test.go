package resource

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
)

const testOrgID = "11111111-1111-1111-1111-111111111111"

// runCmd executes `widget <args...>` against cfg and returns captured stdout.
func runCmd[T any](t *testing.T, cfg Config[T], args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	Register(root, cfg)
	root.SetArgs(append([]string{cfg.Name}, args...))

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	runErr := root.Execute()
	os.Stdout = orig
	_ = w.Close()
	out, _ := io.ReadAll(r)
	_ = r.Close()
	return string(out), runErr
}

func realClient(t *testing.T) *terrakube.Client {
	t.Helper()
	c, err := terrakube.NewClient(terrakube.WithEndpoint("http://127.0.0.1:1"), terrakube.WithToken("x"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func wantErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q, got nil", want)
	}
	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

func TestCommands_NewClientError(t *testing.T) {
	cfg := testConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return nil, errors.New("no client") }
	for _, args := range [][]string{
		{"list", "--organization", testOrgID},
		{"get", "--organization", testOrgID, "--id", "x"},
		{"create", "--organization", testOrgID, "--name", "n"},
		{"update", "--organization", testOrgID, "--id", "x"},
		{"delete", "--organization", testOrgID, "--id", "x"},
	} {
		t.Run(args[0], func(t *testing.T) {
			_, err := runCmd(t, cfg, args...)
			wantErr(t, err, "no client")
		})
	}
}

func TestCommands_ParentResolveError(t *testing.T) {
	cfg := testConfig() // --organization omitted: parent is required
	for _, args := range [][]string{
		{"list"},
		{"get", "--id", "x"},
		{"create", "--name", "n"},
		{"update", "--id", "x"},
		{"delete", "--id", "x"},
	} {
		t.Run(args[0], func(t *testing.T) {
			_, err := runCmd(t, cfg, args...)
			wantErr(t, err, "--organization is required")
		})
	}
}

func TestCommands_APIErrors(t *testing.T) {
	apiErr := errors.New("api boom")
	cfg := testConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.List = func(context.Context, *terrakube.Client, []string, *terrakube.ListOptions) ([]*testResource, error) {
		return nil, apiErr
	}
	cfg.Get = func(context.Context, *terrakube.Client, []string, string) (*testResource, error) { return nil, apiErr }
	cfg.Create = func(context.Context, *terrakube.Client, []string, *testResource) (*testResource, error) {
		return nil, apiErr
	}
	cfg.Update = func(context.Context, *terrakube.Client, []string, *testResource) (*testResource, error) {
		return nil, apiErr
	}
	cfg.Delete = func(context.Context, *terrakube.Client, []string, string) error { return apiErr }

	for _, args := range [][]string{
		{"list", "--organization", testOrgID},
		{"get", "--organization", testOrgID, "--id", "x"},
		{"create", "--organization", testOrgID, "--name", "n"},
		{"update", "--organization", testOrgID, "--id", "x"},
		{"delete", "--organization", testOrgID, "--id", "x"},
	} {
		t.Run(args[0], func(t *testing.T) {
			out, err := runCmd(t, cfg, args...)
			wantErr(t, err, "api boom")
			if out != "" {
				t.Errorf("expected no output on error, got %q", out)
			}
		})
	}
}

func TestListCmd_FilterPassedToList(t *testing.T) {
	cfg := testConfig()
	var gotOpts *terrakube.ListOptions
	var gotParents []string
	cfg.List = func(_ context.Context, _ *terrakube.Client, parents []string, opts *terrakube.ListOptions) ([]*testResource, error) {
		gotOpts, gotParents = opts, parents
		return []*testResource{{ID: "a", Name: "one"}}, nil
	}
	out, err := runCmd(t, cfg, "list", "--organization", testOrgID, "--filter", "name==one")
	if err != nil {
		t.Fatal(err)
	}
	if gotOpts == nil || gotOpts.Filter != "name==one" {
		t.Errorf("expected filter name==one, got %+v", gotOpts)
	}
	if len(gotParents) != 1 || gotParents[0] != testOrgID {
		t.Errorf("unexpected parents %v", gotParents)
	}
	if !strings.Contains(out, `"id": "a"`) {
		t.Errorf("expected rendered resource, got %q", out)
	}

	if _, err := runCmd(t, cfg, "list", "--organization", testOrgID); err != nil {
		t.Fatal(err)
	}
	if gotOpts != nil {
		t.Errorf("expected nil opts without --filter, got %+v", gotOpts)
	}
}

func TestListCmd_RenderError(t *testing.T) {
	cfg := testConfig()
	cfg.GetOutput = func() string { return "bogus" }
	_, err := runCmd(t, cfg, "list", "--organization", testOrgID)
	wantErr(t, err, `unsupported output format: "bogus"`)
}

func TestDeleteCmd_PrintsConfirmation(t *testing.T) {
	cfg := testConfig()
	var gotID string
	cfg.Delete = func(_ context.Context, _ *terrakube.Client, _ []string, id string) error {
		gotID = id
		return nil
	}
	out, err := runCmd(t, cfg, "delete", "--organization", testOrgID, "--id", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if gotID != "abc" || out != "widget deleted\n" {
		t.Errorf("id=%q out=%q", gotID, out)
	}
}

type richResource struct {
	ID    string     `jsonapi:"primary,rich"`
	Name  string     `jsonapi:"attr,name"`
	Count int        `jsonapi:"attr,count"`
	Limit *int       `jsonapi:"attr,limit"`
	Tags  []string   `jsonapi:"attr,tags"`
	On    *bool      `jsonapi:"attr,on"`
	At    *time.Time `jsonapi:"attr,at"`
}

func richConfig() Config[richResource] {
	return Config[richResource]{
		Runtime: testRuntime(),
		Name:    "rich",
		Parents: []ParentScope{{Name: "organization", Flag: "organization"}},
		Fields: []FieldDef{
			{StructField: "Name", Flag: "name", Type: String},
			{StructField: "Count", Flag: "count", Type: Int},
			{StructField: "Limit", Flag: "limit", Type: Int},
			{StructField: "Tags", Flag: "tags", Type: StringSlice},
			{StructField: "On", Flag: "on", Type: Bool},
			{StructField: "At", Flag: "at", Type: String},
		},
	}
}

func TestCreateCmd_PopulatesTypedFields(t *testing.T) {
	cfg := richConfig()
	var got *richResource
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, r *richResource) (*richResource, error) {
		got = r
		return r, nil
	}
	_, err := runCmd(t, cfg, "create", "--organization", testOrgID,
		"--name", "n", "--count", "3", "--limit", "7", "--tags", "a,b", "--on=true", "--at", "2026-01-02T15:04:05Z")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "n" || got.Count != 3 {
		t.Errorf("name/count: %+v", got)
	}
	if got.Limit == nil || *got.Limit != 7 {
		t.Errorf("expected pointer Limit 7, got %v", got.Limit)
	}
	if !reflect.DeepEqual(got.Tags, []string{"a", "b"}) {
		t.Errorf("tags: %v", got.Tags)
	}
	if got.On == nil || !*got.On {
		t.Errorf("expected On=true, got %v", got.On)
	}
	if got.At == nil || !got.At.Equal(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Errorf("at: %v", got.At)
	}
}

func TestCreateCmd_EmptyStringSliceLeavesNil(t *testing.T) {
	cfg := richConfig()
	var got *richResource
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, r *richResource) (*richResource, error) {
		got = r
		return r, nil
	}
	if _, err := runCmd(t, cfg, "create", "--organization", testOrgID); err != nil {
		t.Fatal(err)
	}
	if got.Tags != nil {
		t.Errorf("expected nil Tags, got %v", got.Tags)
	}
}

func TestCreateCmd_InvalidTimestampError(t *testing.T) {
	cfg := richConfig()
	cfg.Create = func(context.Context, *terrakube.Client, []string, *richResource) (*richResource, error) {
		t.Error("Create must not be called when flag parsing fails")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--at", "tomorrow")
	if err == nil || !strings.HasPrefix(err.Error(), "--at: want an RFC 3339 timestamp") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetFieldFromFlag_UnknownStructField(t *testing.T) {
	cmd := &cobra.Command{}
	f := FieldDef{StructField: "Nope", Flag: "nope", Type: String}
	addFieldFlags(cmd, []FieldDef{f}, false)
	err := setFieldFromFlag(cmd, f, &testResource{})
	wantErr(t, err, `struct has no field "Nope"`)
	wantErr(t, populateFields(cmd, []FieldDef{f}, &testResource{}), `struct has no field "Nope"`)

	_ = cmd.Flags().Set("nope", "x")
	wantErr(t, populateChangedFields(cmd, []FieldDef{f}, &testResource{}), `struct has no field "Nope"`)
}

func TestSetIntField(t *testing.T) {
	r := &richResource{}
	v := reflect.ValueOf(r).Elem()
	setIntField(v.FieldByName("Count"), 5)
	setIntField(v.FieldByName("Limit"), 9)
	if r.Count != 5 {
		t.Errorf("Count=%d", r.Count)
	}
	if r.Limit == nil || *r.Limit != 9 {
		t.Errorf("Limit=%v", r.Limit)
	}
}

func TestSetStringSliceField(t *testing.T) {
	r := &richResource{}
	f := reflect.ValueOf(r).Elem().FieldByName("Tags")
	setStringSliceField(f, nil)
	if r.Tags != nil {
		t.Errorf("empty slice must not set field, got %v", r.Tags)
	}
	setStringSliceField(f, []string{"x", "y"})
	if !reflect.DeepEqual(r.Tags, []string{"x", "y"}) {
		t.Errorf("Tags=%v", r.Tags)
	}
}

func TestSetBoolField_NonPointer(t *testing.T) {
	r := &testResource{}
	setBoolField(reflect.ValueOf(r).Elem().FieldByName("Flag"), true)
	if !r.Flag {
		t.Error("expected Flag true")
	}
}

func TestUpdateCmd_RendersResult(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	var sent *richResource
	cfg.Update = func(_ context.Context, _ *terrakube.Client, _ []string, r *richResource) (*richResource, error) {
		sent = r
		return &richResource{ID: "id-1", Name: "from-api"}, nil
	}
	cfg.Get = func(context.Context, *terrakube.Client, []string, string) (*richResource, error) {
		t.Error("Get must not be called when Update returns a resource")
		return nil, nil
	}
	out, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "id-1", "--name", "new")
	if err != nil {
		t.Fatal(err)
	}
	if sent.ID != "id-1" || sent.Name != "new" || sent.Count != 0 {
		t.Errorf("only changed fields + ID expected, got %+v", sent)
	}
	if !strings.Contains(out, "from-api") {
		t.Errorf("expected API result rendered, got %q", out)
	}
}

func TestUpdateCmd_EmptyResultFallsBackToGet(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, *richResource) (*richResource, error) {
		return nil, nil // 204 No Content
	}
	cfg.Get = func(_ context.Context, _ *terrakube.Client, _ []string, id string) (*richResource, error) {
		return &richResource{ID: id, Name: "stored"}, nil
	}
	out, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "id-2", "--name", "sent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "stored") || strings.Contains(out, "sent") {
		t.Errorf("expected stored state, got %q", out)
	}
}

func TestUpdateCmd_EmptyIDResultFallsBackToSentFieldsWhenGetFails(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, *richResource) (*richResource, error) {
		return &richResource{}, nil
	}
	cfg.Get = func(context.Context, *terrakube.Client, []string, string) (*richResource, error) {
		return nil, errors.New("soft deleted")
	}
	out, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "id-3", "--name", "sent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "sent") || !strings.Contains(out, "id-3") {
		t.Errorf("expected sent fields rendered, got %q", out)
	}
}

func TestUpdateCmd_NoGetConfigured(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, *richResource) (*richResource, error) {
		return nil, nil
	}
	out, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "id-4", "--name", "sent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "sent") {
		t.Errorf("expected sent fields rendered, got %q", out)
	}
}

func TestUpdateCmd_InvalidTimestampError(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, *richResource) (*richResource, error) {
		t.Error("Update must not be called")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "x", "--at", "nope")
	if err == nil || !strings.HasPrefix(err.Error(), "--at: want an RFC 3339 timestamp") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func patchServer(t *testing.T, got *string, method *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got, *method = string(b), r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSparsePatch_RewritesPatchBody(t *testing.T) {
	var body, method string
	srv := patchServer(t, &body, &method)

	hc := &http.Client{Transport: sparsePatch{attrs: map[string]any{"name": "new"}}}
	in := `{"data":{"type":"widget","id":"w1","attributes":{"name":"old","other":"x"},"relationships":{"a":{}}},"included":[1]}`
	req, _ := http.NewRequest(http.MethodPatch, srv.URL, strings.NewReader(in))
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if method != http.MethodPatch {
		t.Errorf("method=%s", method)
	}
	want := `{"data":{"type":"widget","id":"w1","attributes":{"name":"new"}}}`
	if body != want {
		t.Errorf("body=%s want %s", body, want)
	}
}

func TestSparsePatch_NonPatchPassesThrough(t *testing.T) {
	var body, method string
	srv := patchServer(t, &body, &method)

	hc := &http.Client{Transport: sparsePatch{attrs: map[string]any{"name": "new"}}}
	req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("raw"))
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if method != http.MethodPost || body != "raw" {
		t.Errorf("method=%s body=%q", method, body)
	}

	req, _ = http.NewRequest(http.MethodPatch, srv.URL, nil)
	resp, err = hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if method != http.MethodPatch || body != "" {
		t.Errorf("nil-body PATCH: method=%s body=%q", method, body)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read fail") }

func TestSparsePatch_BodyErrors(t *testing.T) {
	s := sparsePatch{attrs: map[string]any{}}

	req, _ := http.NewRequest(http.MethodPatch, "http://127.0.0.1:1", errReader{})
	_, err := s.RoundTrip(req)
	wantErr(t, err, "read fail")

	req, _ = http.NewRequest(http.MethodPatch, "http://127.0.0.1:1", strings.NewReader("{not json"))
	_, err = s.RoundTrip(req)
	if err == nil || !strings.HasPrefix(err.Error(), "rewriting update body: ") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCmd_RendersResultAndRenderError(t *testing.T) {
	cfg := testConfig()
	cfg.Get = func(_ context.Context, _ *terrakube.Client, _ []string, id string) (*testResource, error) {
		return &testResource{ID: id, Name: "got"}, nil
	}
	out, err := runCmd(t, cfg, "get", "--organization", testOrgID, "--id", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"id": "g1"`) {
		t.Errorf("out=%q", out)
	}

	cfg.GetOutput = func() string { return "bogus" }
	_, err = runCmd(t, cfg, "get", "--organization", testOrgID, "--id", "g1")
	wantErr(t, err, `unsupported output format: "bogus"`)
}

func TestSparsePatch_UnmarshalableAttrs(t *testing.T) {
	s := sparsePatch{attrs: map[string]any{"bad": make(chan int)}}
	req, _ := http.NewRequest(http.MethodPatch, "http://127.0.0.1:1", strings.NewReader(`{"data":{"type":"w","id":"1"}}`))
	if _, err := s.RoundTrip(req); err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
