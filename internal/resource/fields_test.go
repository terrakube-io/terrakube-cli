package resource

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	terrakube "github.com/terrakube-io/terrakube-go"
)

type extraResource struct {
	ID      string         `jsonapi:"primary,extra"`
	Body    string         `jsonapi:"attr,body"`
	Vars    map[string]any `jsonapi:"attr,vars"`
	Targets []*linkTarget  `jsonapi:"relation,targets,omitempty"`
}

func extraConfig(got *terrakube.Patch) Config[extraResource] {
	write := func(p terrakube.Patch) (*extraResource, error) {
		*got = p
		return &extraResource{ID: "x1"}, nil
	}
	return Config[extraResource]{
		Runtime: testRuntime(),
		Name:    "extra",
		Parents: []ParentScope{{Name: "organization", Flag: "organization"}},
		Fields: []FieldDef{
			{StructField: "Body", Flag: "body", Type: String, Required: true, FileFlag: "body-file"},
			{StructField: "Vars", Flag: "vars", Type: StringMap},
			{
				StructField: "Targets", Flag: "targets", Type: RelationshipMany, RelType: "target",
				Resolver: func(_ context.Context, _ *terrakube.Client, _ []string, name string) (string, error) {
					if name == "missing" {
						return "", errors.New(`target "missing" not found`)
					}
					return "id-" + name, nil
				},
			},
		},
		Create: func(_ context.Context, _ *terrakube.Client, _ []string, p terrakube.Patch) (*extraResource, error) {
			return write(p)
		},
		Update: func(_ context.Context, _ *terrakube.Client, _ []string, _ string, p terrakube.Patch) (*extraResource, error) {
			return write(p)
		},
	}
}

func TestFileFlag_ReadsTheValueFromAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "body.yaml")
	if err := os.WriteFile(path, []byte("flow:\n  - type: terraformPlan\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got terrakube.Patch
	if _, err := runCmd(t, extraConfig(&got), "create", "--organization", testOrgID, "--body-file", path); err != nil {
		t.Fatal(err)
	}
	if got.Attributes["body"] != "flow:\n  - type: terraformPlan\n" || len(got.Attributes) != 1 {
		t.Errorf("attributes %v", got.Attributes)
	}
	if _, err := runCmd(t, extraConfig(&got), "update", "--organization", testOrgID, "--id", "x1", "--body-file", path); err != nil {
		t.Fatal(err)
	}
	if got.Attributes["body"] != "flow:\n  - type: terraformPlan\n" {
		t.Errorf("update attributes %v", got.Attributes)
	}
}

func TestFileFlag_Errors(t *testing.T) {
	var got terrakube.Patch
	cases := map[string]struct {
		args []string
		want string
	}{
		"neither":      {[]string{"create"}, "at least one of the flags in the group [body body-file] is required"},
		"both":         {[]string{"create", "--body", "b", "--body-file", "f"}, "if any flags in the group [body body-file] are set none of the others can be"},
		"missing file": {[]string{"create", "--body-file", filepath.Join(t.TempDir(), "nope")}, "--body-file: open "},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got = terrakube.Patch{}
			_, err := runCmd(t, extraConfig(&got), append(c.args, "--organization", testOrgID)...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected error containing %q, got %v", c.want, err)
			}
			if got.Attributes != nil {
				t.Errorf("nothing should be written, got %v", got)
			}
		})
	}
}

func TestStringMap_ParsesKeyValuePairs(t *testing.T) {
	var got terrakube.Patch
	if _, err := runCmd(t, extraConfig(&got), "create", "--organization", testOrgID, "--body", "b",
		"--vars", "region=eu-west-1", "--vars", "tags=a=b,c"); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"region": "eu-west-1", "tags": "a=b,c"}
	if !reflect.DeepEqual(got.Attributes["vars"], want) {
		t.Errorf("vars %#v, want %#v", got.Attributes["vars"], want)
	}

	got = terrakube.Patch{}
	_, err := runCmd(t, extraConfig(&got), "create", "--organization", testOrgID, "--body", "b", "--vars", "region")
	wantErr(t, err, `--vars: "region" is not KEY=VALUE`)
	if got.Attributes != nil {
		t.Errorf("nothing should be written, got %v", got)
	}
}

func TestRelationshipMany_IDsNamesAndClear(t *testing.T) {
	const id = "11111111-2222-4333-8444-555555555555"
	var got terrakube.Patch
	if _, err := runCmd(t, extraConfig(&got), "update", "--organization", testOrgID, "--id", "x1",
		"--targets", id+",plan", "--targets", "apply"); err != nil {
		t.Fatal(err)
	}
	rel := got.Relationships["targets"]
	if !rel.Many || rel.Type != "target" || !reflect.DeepEqual(rel.IDs, []string{id, "id-plan", "id-apply"}) {
		t.Errorf("relationship %+v", rel)
	}

	if _, err := runCmd(t, extraConfig(&got), "update", "--organization", testOrgID, "--id", "x1", "--targets", ""); err != nil {
		t.Fatal(err)
	}
	if rel := got.Relationships["targets"]; !rel.Many || rel.Type != "target" || len(rel.IDs) != 0 {
		t.Errorf("clear: relationship %+v", rel)
	}

	got = terrakube.Patch{}
	_, err := runCmd(t, extraConfig(&got), "update", "--organization", testOrgID, "--id", "x1", "--targets", "plan,missing")
	wantErr(t, err, `target "missing" not found`)
	if got.Relationships != nil {
		t.Errorf("nothing should be written, got %v", got)
	}

	cfg := extraConfig(&got)
	cfg.Fields[2].Resolver = nil
	_, err = runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "x1", "--targets", "plan")
	wantErr(t, err, `--targets: "plan" is not a UUID`)
}

func TestCheckFields_NewTypes(t *testing.T) {
	checkFields[extraResource](extraConfig(new(terrakube.Patch)).Fields)
	for name, f := range map[string]FieldDef{
		"map on a string":       {StructField: "Body", Flag: "b", Type: StringMap},
		"many on an attribute":  {StructField: "Vars", Flag: "v", Type: RelationshipMany, RelType: "x"},
		"many without rel type": {StructField: "Targets", Flag: "t", Type: RelationshipMany},
		"file flag on a map":    {StructField: "Vars", Flag: "v", Type: StringMap, FileFlag: "v-file"},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected checkFields to panic")
				}
			}()
			checkFields[extraResource]([]FieldDef{f})
		})
	}
}
