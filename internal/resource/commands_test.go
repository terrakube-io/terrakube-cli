package resource

import (
	"context"
	"errors"
	"io"
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
		{"update", "--organization", testOrgID, "--id", "x", "--name", "n"},
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
	cfg.Create = func(context.Context, *terrakube.Client, []string, terrakube.Patch) (*testResource, error) {
		return nil, apiErr
	}
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*testResource, error) {
		return nil, apiErr
	}
	cfg.Delete = func(context.Context, *terrakube.Client, []string, string) error { return apiErr }

	for _, args := range [][]string{
		{"list", "--organization", testOrgID},
		{"get", "--organization", testOrgID, "--id", "x"},
		{"create", "--organization", testOrgID, "--name", "n"},
		{"update", "--organization", testOrgID, "--id", "x", "--name", "n"},
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
	var got terrakube.Patch
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, p terrakube.Patch) (*richResource, error) {
		got = p
		return &richResource{ID: "r1"}, nil
	}
	_, err := runCmd(t, cfg, "create", "--organization", testOrgID,
		"--name", "n", "--count", "3", "--limit", "7", "--tags", "a,b", "--on=true", "--at", "2026-01-02T15:04:05Z")
	if err != nil {
		t.Fatal(err)
	}
	a := got.Attributes
	if a["name"] != "n" || a["count"] != 3 {
		t.Errorf("name/count: %v", a)
	}
	if limit, ok := a["limit"].(*int); !ok || *limit != 7 {
		t.Errorf("expected pointer limit 7, got %#v", a["limit"])
	}
	if !reflect.DeepEqual(a["tags"], []string{"a", "b"}) {
		t.Errorf("tags: %v", a["tags"])
	}
	if on, ok := a["on"].(*bool); !ok || !*on {
		t.Errorf("expected on=true, got %#v", a["on"])
	}
	if at, ok := a["at"].(*time.Time); !ok || !at.Equal(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Errorf("at: %#v", a["at"])
	}
}

func TestCreateCmd_SendsOnlySetFlags(t *testing.T) {
	cfg := richConfig()
	var got terrakube.Patch
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, p terrakube.Patch) (*richResource, error) {
		got = p
		return &richResource{ID: "r1"}, nil
	}
	if _, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--on=false"); err != nil {
		t.Fatal(err)
	}
	if len(got.Attributes) != 1 {
		t.Fatalf("expected only the set flag, got %v", got.Attributes)
	}
	if on, ok := got.Attributes["on"].(*bool); !ok || *on {
		t.Errorf("expected explicit on=false, got %#v", got.Attributes["on"])
	}
}

func TestCreateCmd_InvalidTimestampError(t *testing.T) {
	cfg := richConfig()
	cfg.Create = func(context.Context, *terrakube.Client, []string, terrakube.Patch) (*richResource, error) {
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
	wantErr(t, setFieldFromFlag(cmd, f, &testResource{}), `struct has no field "Nope"`)

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
	var sentID string
	var sent terrakube.Patch
	cfg.Update = func(_ context.Context, _ *terrakube.Client, _ []string, id string, p terrakube.Patch) (*richResource, error) {
		sentID, sent = id, p
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
	if sentID != "id-1" || len(sent.Attributes) != 1 || sent.Attributes["name"] != "new" {
		t.Errorf("only changed fields + ID expected, got id=%q %v", sentID, sent.Attributes)
	}
	if !strings.Contains(out, "from-api") {
		t.Errorf("expected API result rendered, got %q", out)
	}
}

func TestUpdateCmd_EmptyResultFallsBackToGet(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*richResource, error) {
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
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*richResource, error) {
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
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*richResource, error) {
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
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*richResource, error) {
		t.Error("Update must not be called")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "x", "--at", "nope")
	if err == nil || !strings.HasPrefix(err.Error(), "--at: want an RFC 3339 timestamp") {
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

func TestUpdateCmd_NoFieldFlagsIsAnError(t *testing.T) {
	cfg := richConfig()
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*richResource, error) {
		t.Error("Update must not be called without a field flag")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "x")
	wantErr(t, err, "nothing to update: set at least one field flag, see --help")
}

type linkTarget struct {
	ID string `jsonapi:"primary,target"`
}

type linkedResource struct {
	ID     string      `jsonapi:"primary,linked"`
	Name   string      `jsonapi:"attr,name"`
	Target *linkTarget `jsonapi:"relation,target,omitempty"`
}

func linkedConfig(resolver func(context.Context, *terrakube.Client, []string, string) (string, error)) Config[linkedResource] {
	return Config[linkedResource]{
		Runtime: testRuntime(),
		Name:    "linked",
		Parents: []ParentScope{{Name: "organization", Flag: "organization", IDFlag: "organization-id"}},
		Fields: []FieldDef{
			{StructField: "Name", Flag: "name", Type: String},
			{StructField: "Target", Flag: "target", Type: Relationship, RelType: "target", Resolver: resolver},
		},
	}
}

func TestRelationship_NameWithoutResolverIsAnError(t *testing.T) {
	cfg := linkedConfig(nil)
	cfg.Create = func(context.Context, *terrakube.Client, []string, terrakube.Patch) (*linkedResource, error) {
		t.Error("Create must not be called")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--target", "prod")
	wantErr(t, err, `--target: "prod" is not a UUID`)
}

func TestRelationship_NameGoesThroughResolver(t *testing.T) {
	const resolved = "11111111-2222-4333-8444-555555555555"
	var gotParents []string
	cfg := linkedConfig(func(_ context.Context, _ *terrakube.Client, parents []string, name string) (string, error) {
		gotParents = parents
		if name != "prod" {
			t.Errorf("resolver got %q", name)
		}
		return resolved, nil
	})
	var got terrakube.Patch
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, p terrakube.Patch) (*linkedResource, error) {
		got = p
		return &linkedResource{ID: "l1"}, nil
	}
	if _, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--target", "prod"); err != nil {
		t.Fatal(err)
	}
	rel := got.Relationships["target"]
	if rel.Type != "target" || rel.ID == nil || *rel.ID != resolved {
		t.Errorf("relationship: %+v", rel)
	}
	if len(gotParents) != 1 || gotParents[0] != testOrgID {
		t.Errorf("resolver parents: %v", gotParents)
	}
}

func TestRelationship_ResolverErrorStopsTheWrite(t *testing.T) {
	cfg := linkedConfig(func(context.Context, *terrakube.Client, []string, string) (string, error) {
		return "", errors.New(`target "prod" not found`)
	})
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*linkedResource, error) {
		t.Error("Update must not be called")
		return nil, nil
	}
	_, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "x", "--target", "prod")
	wantErr(t, err, `target "prod" not found`)
}

func TestCheckFields_RelationshipNeedsRelationFieldAndType(t *testing.T) {
	for name, f := range map[string]FieldDef{
		"attr field":  {StructField: "Name", Flag: "name", Type: Relationship, RelType: "x"},
		"no rel type": {StructField: "Target", Flag: "target", Type: Relationship},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected checkFields to panic")
				}
			}()
			checkFields[linkedResource]([]FieldDef{f})
		})
	}
}
