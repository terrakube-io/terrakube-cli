package resource

import (
	"context"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"
)

type testResource struct {
	ID   string  `jsonapi:"primary,test"`
	Name string  `jsonapi:"attr,name"`
	Desc *string `jsonapi:"attr,desc"`
	Flag bool    `jsonapi:"attr,flag"`
}

func testRuntime() Runtime {
	return Runtime{
		NewClient:  func() *terrakube.Client { return nil },
		GetContext:  func() context.Context { return context.Background() },
		GetOutput:  func() string { return "json" },
	}
}

func testConfig() Config[testResource] {
	return Config[testResource]{
		Runtime: testRuntime(),
		Name:    "widget",
		Aliases: []string{"wgt"},
		Parents: []ParentScope{{
			Name:     "organization",
			IDFlag:   "organization-id",
			NameFlag: "organization-name",
		}},
		Fields: []FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: String, Required: true, Description: "widget name"},
			{StructField: "Desc", Flag: "description", Short: "d", Type: String, Description: "widget description"},
			{StructField: "Flag", Flag: "flag", Short: "f", Type: Bool, Description: "widget flag"},
		},
		List:   func(_ context.Context, _ *terrakube.Client, _ []string, _ *terrakube.ListOptions) ([]*testResource, error) { return nil, nil },
		Get:    func(_ context.Context, _ *terrakube.Client, _ []string, _ string) (*testResource, error) { return nil, nil },
		Create: func(_ context.Context, _ *terrakube.Client, _ []string, _ *testResource) (*testResource, error) { return nil, nil },
		Update: func(_ context.Context, _ *terrakube.Client, _ []string, _ *testResource) (*testResource, error) { return nil, nil },
		Delete: func(_ context.Context, _ *terrakube.Client, _ []string, _ string) error { return nil },
	}
}

func TestRegister_CommandTree(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cfg := testConfig()
	Register(root, cfg)

	widgetCmd, _, err := root.Find([]string{"widget"})
	if err != nil {
		t.Fatalf("expected widget command, got error: %v", err)
	}
	if widgetCmd.Use != "widget list|get|create|update|delete [FLAGS]" {
		t.Errorf("unexpected Use: %s", widgetCmd.Use)
	}

	subcmds := widgetCmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subcmds {
		names[sub.Name()] = true
	}

	for _, expected := range []string{"list", "get", "create", "update", "delete"} {
		if !names[expected] {
			t.Errorf("expected subcommand %q, not found", expected)
		}
	}
}

func TestRegister_Aliases(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root, testConfig())

	widgetCmd, _, _ := root.Find([]string{"wgt"})
	if widgetCmd == nil {
		t.Fatal("expected alias 'wgt' to resolve")
	}
}

func TestRegister_OnlyListAndDelete(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cfg := Config[testResource]{
		Runtime: testRuntime(),
		Name:    "partial",
		Fields:  []FieldDef{},
		List:    func(_ context.Context, _ *terrakube.Client, _ []string, _ *terrakube.ListOptions) ([]*testResource, error) { return nil, nil },
		Delete:  func(_ context.Context, _ *terrakube.Client, _ []string, _ string) error { return nil },
	}
	Register(root, cfg)

	partialCmd, _, _ := root.Find([]string{"partial"})
	subcmds := partialCmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subcmds {
		names[sub.Name()] = true
	}

	if !names["list"] {
		t.Error("expected list subcommand")
	}
	if !names["delete"] {
		t.Error("expected delete subcommand")
	}
	if names["get"] {
		t.Error("did not expect get subcommand")
	}
	if names["create"] {
		t.Error("did not expect create subcommand")
	}
	if names["update"] {
		t.Error("did not expect update subcommand")
	}
}

func TestRegister_ParentFlags(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root, testConfig())

	listCmd, _, _ := root.Find([]string{"widget", "list"})
	if listCmd == nil {
		t.Fatal("expected list command")
	}

	orgIDFlag := listCmd.Flags().Lookup("organization-id")
	if orgIDFlag == nil {
		t.Error("expected --organization-id flag on list")
	}

	orgNameFlag := listCmd.Flags().Lookup("organization-name")
	if orgNameFlag == nil {
		t.Error("expected --organization-name flag on list")
	}

	filterFlag := listCmd.Flags().Lookup("filter")
	if filterFlag == nil {
		t.Error("expected --filter flag on list")
	}
}

func TestRegister_CreateRequiredFlags(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root, testConfig())

	createCmd, _, _ := root.Find([]string{"widget", "create"})
	if createCmd == nil {
		t.Fatal("expected create command")
	}

	nameFlag := createCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("expected --name flag on create")
	}

	// Check that description is NOT required on create
	descFlag := createCmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Fatal("expected --description flag on create")
	}
}

func TestRegister_GetAndDeleteRequireID(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root, testConfig())

	for _, subcmd := range []string{"get", "delete"} {
		cmd, _, _ := root.Find([]string{"widget", subcmd})
		if cmd == nil {
			t.Fatalf("expected %s command", subcmd)
		}

		idFlag := cmd.Flags().Lookup("id")
		if idFlag == nil {
			t.Errorf("expected --id flag on %s", subcmd)
		}
	}
}

func TestRegister_UpdateHasIDAndFields(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root, testConfig())

	updateCmd, _, _ := root.Find([]string{"widget", "update"})
	if updateCmd == nil {
		t.Fatal("expected update command")
	}

	if updateCmd.Flags().Lookup("id") == nil {
		t.Error("expected --id flag on update")
	}
	if updateCmd.Flags().Lookup("name") == nil {
		t.Error("expected --name flag on update")
	}
	if updateCmd.Flags().Lookup("description") == nil {
		t.Error("expected --description flag on update")
	}
}

func TestPopulateFields(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("name", "n", "", "")
	cmd.Flags().StringP("description", "d", "", "")
	cmd.Flags().BoolP("flag", "f", false, "")
	_ = cmd.Flags().Set("name", "test-widget")
	_ = cmd.Flags().Set("description", "a desc")
	_ = cmd.Flags().Set("flag", "true")

	fields := []FieldDef{
		{StructField: "Name", Flag: "name", Type: String},
		{StructField: "Desc", Flag: "description", Type: String},
		{StructField: "Flag", Flag: "flag", Type: Bool},
	}

	r := &testResource{}
	if err := populateFields(cmd, fields, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Name != "test-widget" {
		t.Errorf("expected Name 'test-widget', got %q", r.Name)
	}
	if r.Desc == nil || *r.Desc != "a desc" {
		t.Errorf("expected Desc 'a desc', got %v", r.Desc)
	}
	if !r.Flag {
		t.Error("expected Flag true")
	}
}

func TestPopulateFields_EmptyStringLeavesPointerNil(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")

	fields := []FieldDef{
		{StructField: "Name", Flag: "name", Type: String},
		{StructField: "Desc", Flag: "description", Type: String},
	}

	r := &testResource{}
	if err := populateFields(cmd, fields, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Desc != nil {
		t.Errorf("expected nil Desc for empty string, got %v", r.Desc)
	}
}

func TestPopulateChangedFields_OnlyChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	_ = cmd.Flags().Set("name", "updated")
	// description NOT set (not Changed)

	fields := []FieldDef{
		{StructField: "Name", Flag: "name", Type: String},
		{StructField: "Desc", Flag: "description", Type: String},
	}

	r := &testResource{}
	if err := populateChangedFields(cmd, fields, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Name != "updated" {
		t.Errorf("expected Name 'updated', got %q", r.Name)
	}
	if r.Desc != nil {
		t.Errorf("expected Desc to remain nil (unchanged), got %v", r.Desc)
	}
}

func TestSetStructField(t *testing.T) {
	r := &testResource{}
	setStructField(r, "ID", "abc-123")
	if r.ID != "abc-123" {
		t.Errorf("expected ID 'abc-123', got %q", r.ID)
	}
}
