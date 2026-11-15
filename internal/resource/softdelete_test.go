package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
)

func fakeTTY(t *testing.T, tty bool, input string) {
	t.Helper()
	oldTTY, oldIn := IsTerminal, Stdin
	IsTerminal = func() bool { return tty }
	Stdin = strings.NewReader(input)
	t.Cleanup(func() { IsTerminal, Stdin = oldTTY, oldIn })
}

// softConfig records soft (Update) and hard (Delete) calls.
func softConfig(soft *SoftDelete, calls *[]string) Config[testResource] {
	cfg := testConfig()
	cfg.SoftDelete = soft
	cfg.Get = func(_ context.Context, _ *terrakube.Client, _ []string, id string) (*testResource, error) {
		*calls = append(*calls, "get")
		return &testResource{ID: id, Name: "prod"}, nil
	}
	cfg.Update = func(_ context.Context, _ *terrakube.Client, _ []string, id string, p terrakube.Patch) (*testResource, error) {
		*calls = append(*calls, "update "+id)
		for k, v := range p.Attributes {
			*calls = append(*calls, fmt.Sprintf("%s=%v", k, v))
		}
		return &testResource{ID: id}, nil
	}
	cfg.Delete = func(_ context.Context, _ *terrakube.Client, _ []string, id string) error {
		*calls = append(*calls, "delete "+id)
		return nil
	}
	return cfg
}

func TestSoftDelete_SetsAttribute(t *testing.T) {
	var calls []string
	out, err := runCmd(t, softConfig(&SoftDelete{Attribute: "deleted", Done: "marked as deleted", Hard: true}, &calls),
		"delete", "--organization", testOrgID, "--id", "w1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "update w1,deleted=true" || !strings.Contains(out, "widget marked as deleted") {
		t.Errorf("calls=%v out=%q", calls, out)
	}
}

func TestSoftDelete_HardSendsDelete(t *testing.T) {
	var calls []string
	out, err := runCmd(t, softConfig(&SoftDelete{Attribute: "deleted", Done: "marked as deleted", Hard: true}, &calls),
		"delete", "--organization", testOrgID, "--id", "w1", "--hard")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "delete w1" || !strings.Contains(out, "widget deleted") {
		t.Errorf("calls=%v out=%q", calls, out)
	}
}

func TestSoftDelete_Confirmation(t *testing.T) {
	soft := &SoftDelete{Attribute: "disabled", Done: "disabled", Confirm: true}
	tests := []struct {
		name, input, wantErr, wantCalls string
		tty                             bool
		args                            []string
	}{
		{name: "yes skips the prompt", args: []string{"--yes"}, wantCalls: "update w1,disabled=true"},
		{name: "no tty, no yes", wantErr: "widget delete needs confirmation: run it in a terminal or pass --yes"},
		{name: "name typed back", tty: true, input: "prod\n", wantCalls: "get,update w1,disabled=true"},
		{name: "wrong name", tty: true, input: "dev\n", wantErr: `confirmation did not match "prod": nothing changed`, wantCalls: "get"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTTY(t, tt.tty, tt.input)
			var calls []string
			_, err := runCmd(t, softConfig(soft, &calls), append([]string{"delete", "--organization", testOrgID, "--id", "w1"}, tt.args...)...)
			if tt.wantErr != "" {
				wantErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(calls, ","); got != tt.wantCalls {
				t.Errorf("calls = %q, want %q", got, tt.wantCalls)
			}
		})
	}
}

func TestSoftDelete_Errors(t *testing.T) {
	boom := errors.New("boom")
	soft := &SoftDelete{Attribute: "disabled", Done: "disabled", Confirm: true}

	var calls []string
	cfg := softConfig(soft, &calls)
	cfg.Get = func(context.Context, *terrakube.Client, []string, string) (*testResource, error) { return nil, boom }
	fakeTTY(t, true, "prod\n")
	_, err := runCmd(t, cfg, "delete", "--organization", testOrgID, "--id", "w1")
	wantErr(t, err, "boom")

	cfg = softConfig(soft, &calls)
	cfg.Update = func(context.Context, *terrakube.Client, []string, string, terrakube.Patch) (*testResource, error) {
		return nil, boom
	}
	_, err = runCmd(t, cfg, "delete", "--organization", testOrgID, "--id", "w1", "--yes")
	wantErr(t, err, "boom")

	cfg = softConfig(&SoftDelete{Attribute: "deleted", Hard: true}, &calls)
	cfg.Delete = func(context.Context, *terrakube.Client, []string, string) error { return boom }
	_, err = runCmd(t, cfg, "delete", "--organization", testOrgID, "--id", "w1", "--hard")
	wantErr(t, err, "boom")

	cfg = softConfig(soft, &calls)
	cfg.NewClient = func() (*terrakube.Client, error) { return nil, boom }
	_, err = runCmd(t, cfg, "delete", "--organization", testOrgID, "--id", "w1", "--yes")
	wantErr(t, err, "boom")
}

func TestStringResolver_ResolvesNamesOnly(t *testing.T) {
	const id = "11111111-2222-4333-8444-555555555555"
	resolver := func(_ context.Context, _ *terrakube.Client, _ []string, name string) (string, error) {
		if name == "bad" {
			return "", errors.New(`thing "bad" not found`)
		}
		return id, nil
	}
	cfg := testConfig()
	cfg.Fields = []FieldDef{{StructField: "Name", Flag: "name", Type: String, Resolver: resolver}}
	var got terrakube.Patch
	cfg.Create = func(_ context.Context, _ *terrakube.Client, _ []string, p terrakube.Patch) (*testResource, error) {
		got = p
		return &testResource{ID: "w"}, nil
	}

	for _, tt := range []struct{ in, want string }{{"prod", id}, {id, id}} {
		if _, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--name", tt.in); err != nil {
			t.Fatal(err)
		}
		if got.Attributes["name"] != tt.want {
			t.Errorf("%q: name = %v, want %s", tt.in, got.Attributes["name"], tt.want)
		}
	}
	_, err := runCmd(t, cfg, "create", "--organization", testOrgID, "--name", "bad")
	wantErr(t, err, `thing "bad" not found`)
}

func TestRelationship_UUIDAndClear(t *testing.T) {
	const id = "11111111-2222-4333-8444-555555555555"
	cfg := linkedConfig(nil)
	var got terrakube.Patch
	cfg.NewClient = func() (*terrakube.Client, error) { return realClient(t), nil }
	cfg.Update = func(_ context.Context, _ *terrakube.Client, _ []string, _ string, p terrakube.Patch) (*linkedResource, error) {
		got = p
		return &linkedResource{ID: "l"}, nil
	}
	if _, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "l", "--target", id); err != nil {
		t.Fatal(err)
	}
	if r := got.Relationships["target"]; r.ID == nil || *r.ID != id || r.Type != "target" {
		t.Errorf("uuid: %+v", r)
	}
	if _, err := runCmd(t, cfg, "update", "--organization", testOrgID, "--id", "l", "--target", ""); err != nil {
		t.Fatal(err)
	}
	if r, ok := got.Relationships["target"]; !ok || r.ID != nil || r.Type != "" {
		t.Errorf("clear: %+v (present=%v)", r, ok)
	}
}

func TestHiddenVerbsAndFields(t *testing.T) {
	cfg := testConfig()
	cfg.HiddenVerbs = []string{"delete"}
	cfg.Fields = append(cfg.Fields, FieldDef{StructField: "Desc", Flag: "secret-desc", Type: String, Hidden: true})
	root := &cobra.Command{Use: "test"}
	parent := Register(root, cfg)
	for _, sub := range parent.Commands() {
		if sub.Hidden != (sub.Name() == "delete") {
			t.Errorf("%s hidden=%v", sub.Name(), sub.Hidden)
		}
		if sub.Name() == "create" {
			if f := sub.Flags().Lookup("secret-desc"); f == nil || !f.Hidden {
				t.Errorf("secret-desc flag: %v", f)
			}
		}
	}
}
