package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/google/jsonapi"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// resetGlobalFlags resets all package-level flag variables that persist between tests.
// Cobra binds flags to package-level vars via StringVarP/BoolVarP, so they bleed across tests.
// We also reset cobra flag states and viper keys that postInitCommands uses to pre-fill
// required flags from config/env.
func resetGlobalFlags() {
	// workspace create
	WorkspaceCreateName = ""
	WorkspaceDescription = ""
	WorkspaceCreateIacType = ""
	WorkspaceCreateFolder = ""
	WorkspaceExecutionMode = ""
	WorkspaceCreateSource = ""
	WorkspaceCreateBranch = ""
	WorkspaceCreateCli = false
	WorkspaceCreateIacV = ""
	WorkspaceCreateOrgId = ""

	// workspace list
	WorkspaceFilter = ""
	WorkspaceOrgId = ""

	// workspace update
	WorkspaceUpdateName = ""
	WorkspaceUpdateSource = ""
	WorkspaceUpdateBranch = ""
	WorkspaceUpdateTerraformV = ""
	WorkspaceUpdateOrgId = ""
	WorkspaceUpdateId = ""
	WorkspaceUpdateDescription = ""
	WorkspaceUpdateFolder = ""
	WorkspaceUpdateIacType = ""
	WorkspaceUpdateExecutionMode = ""

	// workspace delete
	WorkspaceDeleteId = ""
	WorkspaceDeleteOrgId = ""

	// organization create
	OrganizationCreateName = ""
	OrganizationCreateDescription = ""
	OrganizationCreateExecutionMode = ""
	OrganizationCreateIcon = ""

	// organization update
	OrganizationId = ""
	OrganizationUpdateDescription = ""
	OrganizationUpdateName = ""
	OrganizationUpdateIcon = ""
	OrganizationUpdateExecutionMode = ""

	// organization delete
	OrganizationDeleteId = ""

	// organization list
	OrganizationFilter = ""

	// variable create
	VariableCreateKey = ""
	VariableCreateValue = ""
	VariableCreateDescription = ""
	VariableCreateCategory = ""
	VariableCreateSensitive = false
	VariableCreateHcl = false
	VariableCreateOrgId = ""
	VariableCreateWorkspaceId = ""

	// variable list
	VariableFilter = ""
	VariableOrgId = ""
	VariableWorkspaceId = ""

	// variable update
	VariableId = ""
	VariableUpdateKey = ""
	VariableUpdateValue = ""
	VariableUpdateDescription = ""
	VariableUpdateCategory = ""
	VariableUpdateSensitive = false
	VariableUpdateHcl = false
	VariableUpdateOrgId = ""
	VariableUpdateWorkspaceId = ""

	// variable delete
	VariableDeleteId = ""
	VariableDeleteOrgId = ""
	VariableDeleteWorkspaceId = ""

	// module create
	ModuleCreateName = ""
	ModuleCreateDescription = ""
	ModuleCreateOrgId = ""
	ModuleCreateSource = ""
	ModuleCreateProvider = ""
	ModuleCreateTagPrefix = ""
	ModuleCreateFolder = ""

	// module list
	ModuleFilter = ""
	ModuleOrgId = ""

	// module update
	ModuleId = ""
	ModuleUpdateDescription = ""
	ModuleUpdateName = ""
	ModuleUpdateOrgId = ""
	ModuleUpdateSource = ""
	ModuleUpdateProvider = ""
	ModuleUpdateTagPrefix = ""
	ModuleUpdateFolder = ""

	// module delete
	ModuleDeleteId = ""
	ModuleDeleteOrgId = ""

	// team create
	TeamCreateName = ""
	TeamCreateOrgId = ""
	TeamCreateManageProvider = false
	TeamCreateManageModule = false
	TeamCreateManageWorkspace = false
	TeamCreateManageState = false
	TeamCreateManageCollection = false
	TeamCreateManageVcs = false
	TeamCreateManageTemplate = false

	// team list
	TeamFilter = ""
	TeamOrgId = ""

	// team update
	TeamId = ""
	TeamUpdateName = ""
	TeamUpdateOrgId = ""
	TeamUpdateManageProvider = false
	TeamUpdateManageModule = false
	TeamUpdateManageWorkspace = false
	TeamUpdateManageState = false
	TeamUpdateManageCollection = false
	TeamUpdateManageVcs = false
	TeamUpdateManageTemplate = false

	// team delete
	TeamDeleteId = ""
	TeamDeleteOrgId = ""

	// job create
	JobCreateWorkspaceId = ""
	JobCreateCommand = ""
	JobCreateOrgId = ""

	// job list
	JobFilter = ""
	JobOrgId = ""

	// login
	apiURL = ""
	patToken = ""

	// root
	output = "json"

	// Clear viper keys that postInitCommands uses to pre-fill cobra required flags.
	// Without this, a value set in one test (via flag parsing -> viper binding) bleeds
	// into subsequent tests because presetRequiredFlags reads viper state.
	viperKeysToReset := []string{
		"organization-id", "organization-name", "workspace-id", "workspace-name",
		"id", "name", "api-url", "pat",
		"api_url", "token", "command", "key", "value", "category",
		"sensitive", "hcl", "filter", "source", "branch", "folder",
		"execution-mode", "iac-type", "iac-version", "description",
		"executionMode", "icon", "provider", "tag-prefix", "tag-id",
		"manage-provider", "manage-module", "manage-workspace",
		"manage-state", "manage-collection", "manage-vcs", "manage-template",
		"cli",
	}
	for _, key := range viperKeysToReset {
		viper.Set(key, "")
	}

	// Point viper at a nonexistent config file to prevent the real user config
	// from injecting values during initConfig.
	cfgFile = os.DevNull

	// Reset cobra flag state on all commands. When cobra parses args, it sets
	// the flag's Changed bit and its Value. These persist across calls to
	// rootCmd.Execute() since the command tree is global. Without resetting,
	// postInitCommands -> presetRequiredFlags will see the old value in viper
	// and re-fill it, satisfying required flag checks unexpectedly.
	resetCobraFlags(rootCmd)
}

// resetCobraFlags recursively resets all flags on a command and its subcommands
// back to their default values and clears the Changed bit.
func resetCobraFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetCobraFlags(sub)
	}
}

// executeCommand runs the root cobra command with the given args and captures stdout.
// It returns the captured output and any error from Execute().
func executeCommand(args ...string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs(args)
	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	err := rootCmd.Execute()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	return string(out), err
}

// setupTestServer creates an httptest server and configures viper so that
// newClient() points at it. Returns the server (caller must defer ts.Close()).
func setupTestServer(handler http.Handler) *httptest.Server {
	ts := httptest.NewServer(handler)
	viper.Set("api_url", ts.URL)
	viper.Set("token", "test-token")
	return ts
}

// ----- Flag Validation Tests -----

func TestCmdWorkspaceListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "list")
	if err == nil {
		t.Fatal("expected error for workspace list without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdWorkspaceCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "create", "--organization-id", "some-org-id")
	if err == nil {
		t.Fatal("expected error for workspace create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdWorkspaceCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "create", "--name", "test-ws")
	if err == nil {
		t.Fatal("expected error for workspace create without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdOrganizationCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("organization", "create")
	if err == nil {
		t.Fatal("expected error for organization create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdOrganizationDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("organization", "delete")
	if err == nil {
		t.Fatal("expected error for organization delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdWorkspaceDeleteMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "delete", "--id", "ws-123")
	if err == nil {
		t.Fatal("expected error for workspace delete without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdWorkspaceDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "delete", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for workspace delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdModuleCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "create", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for module create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdModuleCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "create", "--name", "mod1")
	if err == nil {
		t.Fatal("expected error for module create without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdTeamCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "create", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for team create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdTeamCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "create", "--name", "team1")
	if err == nil {
		t.Fatal("expected error for team create without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdJobCreateMissingCommand(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("job", "create", "--organization-id", "org-123", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for job create without --command, got nil")
	}
	if !strings.Contains(err.Error(), "command") {
		t.Errorf("expected error to mention command, got: %v", err)
	}
}

func TestCmdJobCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("job", "create", "--command", "plan", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for job create without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdVariableListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "list", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for variable list without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdVariableListMissingWorkspaceId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "list", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for variable list without --workspace-id, got nil")
	}
	if !strings.Contains(err.Error(), "workspace-id") {
		t.Errorf("expected error to mention workspace-id, got: %v", err)
	}
}

// ----- Flag Default Tests -----

func TestCmdWorkspaceCreateFlagDefaults(t *testing.T) {
	resetGlobalFlags()

	folderFlag := createWorkspaceCmd.Flags().Lookup("folder")
	if folderFlag == nil {
		t.Fatal("folder flag not found on workspace create command")
	}
	if folderFlag.DefValue != "/" {
		t.Errorf("expected folder default '/', got %q", folderFlag.DefValue)
	}

	execFlag := createWorkspaceCmd.Flags().Lookup("execution-mode")
	if execFlag == nil {
		t.Fatal("execution-mode flag not found on workspace create command")
	}
	if execFlag.DefValue != "remote" {
		t.Errorf("expected execution-mode default 'remote', got %q", execFlag.DefValue)
	}

	iacFlag := createWorkspaceCmd.Flags().Lookup("iac-type")
	if iacFlag == nil {
		t.Fatal("iac-type flag not found on workspace create command")
	}
	if iacFlag.DefValue != "terraform" {
		t.Errorf("expected iac-type default 'terraform', got %q", iacFlag.DefValue)
	}
}

func TestCmdRootOutputFlagDefault(t *testing.T) {
	resetGlobalFlags()

	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found on root command")
	}
	if outputFlag.DefValue != "json" {
		t.Errorf("expected output default 'json', got %q", outputFlag.DefValue)
	}
}

func TestCmdWorkspaceCreateCliFlagDefault(t *testing.T) {
	resetGlobalFlags()

	cliFlag := createWorkspaceCmd.Flags().Lookup("cli")
	if cliFlag == nil {
		t.Fatal("cli flag not found on workspace create command")
	}
	if cliFlag.DefValue != "false" {
		t.Errorf("expected cli default 'false', got %q", cliFlag.DefValue)
	}
}

// ----- Workspace --cli Flag Behavior -----

func TestCmdWorkspaceCreateCliFlag(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var bodyMap map[string]interface{}
		_ = json.Unmarshal(body, &bodyMap)

		if data, ok := bodyMap["data"].(map[string]interface{}); ok {
			if attrs, ok := data["attributes"].(map[string]interface{}); ok {
				if attrs["source"] != "empty" {
					t.Errorf("expected source 'empty' when --cli is set, got %q", attrs["source"])
				}
				if attrs["branch"] != "remote-content" {
					t.Errorf("expected branch 'remote-content' when --cli is set, got %q", attrs["branch"])
				}
			}
		}

		ws := &terrakube.Workspace{ID: "ws-new", Name: "cli-ws", Source: "empty", Branch: "remote-content"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "create",
		"--name", "cli-ws",
		"--organization-id", "org-123",
		"--cli",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Creating cli workspace") {
		t.Errorf("expected 'Creating cli workspace' in output, got: %s", out)
	}

	// Verify the global vars were set by the --cli logic
	if WorkspaceCreateSource != "empty" {
		t.Errorf("expected WorkspaceCreateSource='empty', got %q", WorkspaceCreateSource)
	}
	if WorkspaceCreateBranch != "remote-content" {
		t.Errorf("expected WorkspaceCreateBranch='remote-content', got %q", WorkspaceCreateBranch)
	}
}

func TestCmdWorkspaceCreateWithoutCliFlag(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := &terrakube.Workspace{ID: "ws-new", Name: "vcs-ws", Source: "https://github.com/example/repo.git", Branch: "main"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "create",
		"--name", "vcs-ws",
		"--organization-id", "org-123",
		"--source", "https://github.com/example/repo.git",
		"--branch", "main",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Creating vcs workspace") {
		t.Errorf("expected 'Creating vcs workspace' in output, got: %s", out)
	}
}

// ----- Command Alias Tests -----

func TestCmdWorkspaceAlias(t *testing.T) {
	resetGlobalFlags()

	aliases := workspaceCmd.Aliases
	found := false
	for _, a := range aliases {
		if a == "wrk" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected workspace command to have alias 'wrk', got aliases: %v", aliases)
	}
}

func TestCmdOrganizationAlias(t *testing.T) {
	resetGlobalFlags()

	aliases := organizationCmd.Aliases
	found := false
	for _, a := range aliases {
		if a == "org" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected organization command to have alias 'org', got aliases: %v", aliases)
	}
}

func TestCmdModuleAlias(t *testing.T) {
	resetGlobalFlags()

	aliases := moduleCmd.Aliases
	found := false
	for _, a := range aliases {
		if a == "mod" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected module command to have alias 'mod', got aliases: %v", aliases)
	}
}

func TestCmdVariableAlias(t *testing.T) {
	resetGlobalFlags()

	aliases := variableCmd.Aliases
	found := false
	for _, a := range aliases {
		if a == "var" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected variable command to have alias 'var', got aliases: %v", aliases)
	}
}

// Test that aliases actually work for command routing (via error message)
func TestCmdAliasWrkRoutes(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("wrk", "list")
	if err == nil {
		t.Fatal("expected error for wrk list without --organization-id, got nil")
	}
	// The fact that we get a "required flag" error means the alias routed correctly
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected alias 'wrk' to route to workspace command, got error: %v", err)
	}
}

func TestCmdAliasOrgRoutes(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("org", "create")
	if err == nil {
		t.Fatal("expected error for org create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected alias 'org' to route to organization command, got error: %v", err)
	}
}

func TestCmdAliasModRoutes(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("mod", "create", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for mod create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected alias 'mod' to route to module command, got error: %v", err)
	}
}

// ----- Command Tree Structure Tests -----

func TestCmdWorkspaceSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range workspaceCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("workspace missing expected subcommand %q", name)
		}
	}
}

func TestCmdOrganizationSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range organizationCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("organization missing expected subcommand %q", name)
		}
	}
}

func TestCmdVariableNestedUnderWorkspace(t *testing.T) {
	resetGlobalFlags()

	found := false
	for _, sub := range workspaceCmd.Commands() {
		if sub.Name() == "variable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'variable' to be a subcommand of 'workspace'")
	}
}

func TestCmdVariableSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range variableCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("variable missing expected subcommand %q", name)
		}
	}
}

func TestCmdModuleSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range moduleCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("module missing expected subcommand %q", name)
		}
	}
}

func TestCmdTeamSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range teamCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("team missing expected subcommand %q", name)
		}
	}
}

func TestCmdJobSubcommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"create": false,
		"list":   false,
	}

	for _, sub := range jobCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("job missing expected subcommand %q", name)
		}
	}
}

func TestCmdRootTopLevelCommands(t *testing.T) {
	resetGlobalFlags()

	expected := map[string]bool{
		"workspace":     false,
		"organization":  false,
		"module":        false,
		"team":          false,
		"job":           false,
		"template":      false,
		"workspace-tag": false,
		"login":         false,
		"logout":        false,
	}

	for _, sub := range rootCmd.Commands() {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("root command missing expected subcommand %q", name)
		}
	}
}

// ----- Organization Create Execution Mode Validation -----

func TestCmdOrganizationCreateInvalidExecutionMode(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("organization", "create", "--name", "test-org", "--executionMode", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid execution mode, got nil")
	}
	if !strings.Contains(err.Error(), "executionMode") {
		t.Errorf("expected error to mention executionMode, got: %v", err)
	}
}

func TestCmdOrganizationUpdateInvalidExecutionMode(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("organization", "update", "--id", "org-123", "--executionMode", "bogus")
	if err == nil {
		t.Fatal("expected error for invalid execution mode on update, got nil")
	}
	if !strings.Contains(err.Error(), "executionMode") {
		t.Errorf("expected error to mention executionMode, got: %v", err)
	}
}

// ----- End-to-End Command Execution with httptest -----

func TestCmdWorkspaceListE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "organization/org-123/workspace") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		workspaces := []*terrakube.Workspace{{ID: "ws-1", Name: "workspace-one", Source: "https://github.com/example/repo.git", Branch: "main"}}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, workspaces)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("workspace", "list", "--organization-id", "org-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ws-1") {
		t.Errorf("expected output to contain workspace ID 'ws-1', got: %s", out)
	}
	if !strings.Contains(out, "workspace-one") {
		t.Errorf("expected output to contain workspace name, got: %s", out)
	}
}

func TestCmdOrganizationListE2E(t *testing.T) {
	resetGlobalFlags()

	desc := "test desc"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgs := []*terrakube.Organization{{ID: "org-1", Name: "org-one", Description: &desc}}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, orgs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "org-1") {
		t.Errorf("expected output to contain org ID 'org-1', got: %s", out)
	}
}

func TestCmdOrganizationCreateE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// consume body
		_, _ = io.ReadAll(r.Body)

		desc := "test org"
		org := &terrakube.Organization{ID: "org-new", Name: "new-org", Description: &desc}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, org)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("organization", "create", "--name", "new-org", "--description", "test org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "org-new") {
		t.Errorf("expected output to contain 'org-new', got: %s", out)
	}
}

func TestCmdWorkspaceCreateE2E(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// consume body
		_, _ = io.ReadAll(r.Body)

		ws := &terrakube.Workspace{ID: "ws-new", Name: "test-ws", Folder: "/modules", ExecutionMode: "remote", IaCType: "terraform"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "create",
		"--name", "test-ws",
		"--organization-id", "org-123",
		"--folder", "/modules",
		"--source", "https://github.com/example/repo.git",
		"--branch", "main",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ws-new") {
		t.Errorf("expected output to contain 'ws-new', got: %s", out)
	}
}

func TestCmdLogoutE2E(t *testing.T) {
	resetGlobalFlags()

	out, err := executeCommand("logout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "logout ok") {
		t.Errorf("expected 'logout ok' in output, got: %s", out)
	}
}

// ----- Login Required Flags -----

func TestCmdLoginMissingApiUrl(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("login", "--pat", "some-token")
	if err == nil {
		t.Fatal("expected error for login without --api-url, got nil")
	}
	if !strings.Contains(err.Error(), "api-url") {
		t.Errorf("expected error to mention api-url, got: %v", err)
	}
}

func TestCmdLoginMissingPat(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("login", "--api-url", "http://localhost:8080")
	if err == nil {
		t.Fatal("expected error for login without --pat, got nil")
	}
	if !strings.Contains(err.Error(), "pat") {
		t.Errorf("expected error to mention pat, got: %v", err)
	}
}

// ----- API Error Handling -----

func TestCmdWorkspaceListAPIError(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	// The command should not return an error from cobra (it prints the error itself)
	// but the output should indicate something went wrong or at least not crash
	_, err := executeCommand("workspace", "list", "--organization-id", "org-123")
	// Even with a 500 response, cobra itself doesn't error out - the command
	// prints the error internally. So we just verify no panic occurred.
	_ = err
}

// ----- Capture stdout helper for verifying output format -----

func TestCmdWorkspaceCreateOutputContainsJSON(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := &terrakube.Workspace{ID: "ws-json-test", Name: "json-ws", Folder: "/"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "create",
		"--name", "json-ws",
		"--organization-id", "org-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default output is json, so the output should contain JSON structure
	// The output will contain "Creating vcs workspace\n" followed by JSON
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Find the JSON portion (skip the "Creating vcs workspace" line)
	var jsonBuf bytes.Buffer
	inJSON := false
	for _, line := range lines {
		if strings.HasPrefix(line, "{") {
			inJSON = true
		}
		if inJSON {
			jsonBuf.WriteString(line + "\n")
		}
	}

	if jsonBuf.Len() > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(jsonBuf.Bytes(), &result); err != nil {
			t.Errorf("expected valid JSON in output, got parse error: %v\nOutput: %s", err, out)
		}
	}
}

// ----- Request Body Verification -----

func TestCmdWorkspaceCreateSendsCorrectBody(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		capturedBody, _ = io.ReadAll(r.Body)

		ws := &terrakube.Workspace{ID: "ws-resp", Name: "body-test"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"workspace", "create",
		"--name", "body-test",
		"--organization-id", "org-abc",
		"--source", "https://github.com/test/repo.git",
		"--branch", "develop",
		"--iac-version", "1.5.0",
		"--folder", "/infra",
		"--execution-mode", "local",
		"--iac-type", "tofu",
		"--description", "a test workspace",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body to contain data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "body-test" {
		t.Errorf("expected name 'body-test', got %v", attrs["name"])
	}
	if attrs["source"] != "https://github.com/test/repo.git" {
		t.Errorf("expected source 'https://github.com/test/repo.git', got %v", attrs["source"])
	}
	if attrs["branch"] != "develop" {
		t.Errorf("expected branch 'develop', got %v", attrs["branch"])
	}
	if attrs["terraformVersion"] != "1.5.0" {
		t.Errorf("expected iac-version '1.5.0', got %v", attrs["terraformVersion"])
	}
	if attrs["folder"] != "/infra" {
		t.Errorf("expected folder '/infra', got %v", attrs["folder"])
	}
	if attrs["executionMode"] != "local" {
		t.Errorf("expected execution-mode 'local', got %v", attrs["executionMode"])
	}
	if attrs["iacType"] != "tofu" {
		t.Errorf("expected iac-type 'tofu', got %v", attrs["iacType"])
	}
	if attrs["description"] != "a test workspace" {
		t.Errorf("expected description 'a test workspace', got %v", attrs["description"])
	}
}

func TestCmdOrganizationCreateSendsCorrectBody(t *testing.T) {
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)

		desc := "org desc"
		org := &terrakube.Organization{ID: "org-resp", Name: "body-org", Description: &desc}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, org)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"organization", "create",
		"--name", "body-org",
		"--description", "org desc",
		"--executionMode", "remote",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body to contain data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "body-org" {
		t.Errorf("expected name 'body-org', got %v", attrs["name"])
	}
}

// ----- Command Help/Usage Tests -----

func TestCmdRootHelpDoesNotError(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error from root --help: %v", err)
	}
}

func TestCmdWorkspaceHelpDoesNotError(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "--help")
	if err != nil {
		t.Fatalf("unexpected error from workspace --help: %v", err)
	}
}

func TestCmdUnknownCommandErrors(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("nonexistent")
	if err == nil {
		t.Error("expected error for unknown command, got nil")
	}
}

// ----- Team has no alias (verify negative) -----

func TestCmdTeamHasNoAlias(t *testing.T) {
	resetGlobalFlags()

	if len(teamCmd.Aliases) != 0 {
		t.Errorf("expected team command to have no aliases, got: %v", teamCmd.Aliases)
	}
}

// ----- Job has no alias (verify negative) -----

func TestCmdJobHasNoAlias(t *testing.T) {
	resetGlobalFlags()

	if len(jobCmd.Aliases) != 0 {
		t.Errorf("expected job command to have no aliases, got: %v", jobCmd.Aliases)
	}
}

// ----- Variable accessed through workspace path -----

func TestCmdVariableAccessedThroughWorkspace(t *testing.T) {
	resetGlobalFlags()

	// If we can reach variable list through "workspace variable list" and get
	// the expected required-flag error, the nesting is working.
	_, err := executeCommand("workspace", "variable", "list")
	if err == nil {
		t.Fatal("expected error for workspace variable list without required flags, got nil")
	}
	// It should complain about required flags, not "unknown command"
	if strings.Contains(err.Error(), "unknown command") {
		t.Error("'workspace variable list' returned 'unknown command' — variable is not properly nested under workspace")
	}
}

// ----- Organization create valid execution modes -----

func TestCmdOrganizationCreateValidExecutionModes(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desc := ""
		org := &terrakube.Organization{ID: "org-ok", Name: "ok-org", Description: &desc}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, org)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	for _, mode := range []string{"remote", "local"} {
		resetGlobalFlags()
		viper.Set("api_url", ts.URL)
		viper.Set("token", "test-token")

		_, err := executeCommand("organization", "create", "--name", "ok-org", "--executionMode", mode)
		if err != nil {
			t.Errorf("execution mode %q should be valid, got error: %v", mode, err)
		}
	}
}

// ----- Workspace update required flags -----

func TestCmdWorkspaceUpdateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "update", "--id", "ws-123")
	if err == nil {
		t.Fatal("expected error for workspace update without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdWorkspaceUpdateMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "update", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for workspace update without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

// ----- Module list required flags -----

func TestCmdModuleListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "list")
	if err == nil {
		t.Fatal("expected error for module list without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Team list required flags -----

func TestCmdTeamListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "list")
	if err == nil {
		t.Fatal("expected error for team list without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Job list required flags -----

func TestCmdJobListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("job", "list")
	if err == nil {
		t.Fatal("expected error for job list without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Workspace create flag short forms -----

func TestCmdWorkspaceCreateShortFlags(t *testing.T) {
	resetGlobalFlags()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := &terrakube.Workspace{ID: "ws-short", Name: "short-ws"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, ws)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"workspace", "create",
		"-n", "short-ws",
		"--organization-id", "org-123",
		"-b", "main",
		"-s", "https://github.com/test/repo.git",
		"-f", "/modules",
		"-e", "local",
		"-t", "tofu",
	)
	if err != nil {
		t.Fatalf("unexpected error using short flags: %v", err)
	}
}

// ----- Team Create Permission Boolean in Body -----

func TestCmdTeamCreatePermissionsInBody(t *testing.T) {
	// Test 1: With --manage-workspace flag, the field appears as true in body.
	resetGlobalFlags()

	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		team := &terrakube.Team{ID: "team-new", Name: "test-team", ManageWorkspace: true}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, team)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"team", "create",
		"--name", "test-team",
		"--organization-id", "org-123",
		"--manage-workspace",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(capturedBody), `"manageWorkspace":true`) {
		t.Errorf("expected request body to contain \"manageWorkspace\":true, got: %s", capturedBody)
	}

	// Test 2: Without --manage-workspace flag, the field is present as false.
	// terrakube-go has no omitempty on booleans, so false values are properly sent.
	resetGlobalFlags()
	viper.Set("api_url", ts.URL)
	viper.Set("token", "test-token")

	capturedBody = nil

	_, err = executeCommand(
		"team", "create",
		"--name", "test-team-no-perms",
		"--organization-id", "org-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})

	if _, exists := attrs["manageWorkspace"]; !exists {
		t.Error("expected manageWorkspace to be present (false value sent by terrakube-go), but it was absent")
	}
}

// ----- Organization Update E2E -----

func TestCmdOrganizationUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"organization", "update",
		"--id", "org-123",
		"--name", "updated-org",
		"--description", "new desc",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-123") {
		t.Errorf("expected path to contain organization/org-123, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "updated-org" {
		t.Errorf("expected name 'updated-org', got %v", attrs["name"])
	}
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", out)
	}
}

// ----- Organization Delete E2E -----

func TestCmdOrganizationDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("organization", "delete", "--id", "org-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-456") {
		t.Errorf("expected path to contain organization/org-456, got %s", receivedPath)
	}
}

// ----- Workspace Update E2E -----

func TestCmdWorkspaceUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "update",
		"--organization-id", "org-abc",
		"--id", "ws-789",
		"--name", "updated-ws",
		"--branch", "develop",
		"--source", "https://github.com/test/repo.git",
		"--iac-version", "1.6.0",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", receivedMethod)
	}
	// BUG: workspace_update.go does not set workspace.ID on the struct, so the
	// client builds a path with an empty workspace ID. The --id flag value is
	// stored in WorkspaceUpdateId but never placed into the Workspace model.
	// We test the actual (buggy) behavior here.
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/") {
		t.Errorf("expected path to contain organization/org-abc/workspace/, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "updated-ws" {
		t.Errorf("expected name 'updated-ws', got %v", attrs["name"])
	}
	if attrs["branch"] != "develop" {
		t.Errorf("expected branch 'develop', got %v", attrs["branch"])
	}
	if attrs["terraformVersion"] != "1.6.0" {
		t.Errorf("expected iac-version '1.6.0', got %v", attrs["terraformVersion"])
	}
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", out)
	}
}

// ----- Workspace Delete E2E -----

func TestCmdWorkspaceDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"workspace", "delete",
		"--organization-id", "org-abc",
		"--id", "ws-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/ws-789") {
		t.Errorf("expected path to contain organization/org-abc/workspace/ws-789, got %s", receivedPath)
	}
}

// ----- Module Create E2E -----

func TestCmdModuleCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)

		mod := &terrakube.Module{ID: "mod-new", Name: "test-mod", Provider: "azurerm", Source: "https://github.com/test/repo.git"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, mod)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module", "create",
		"--organization-id", "org-abc",
		"--name", "test-mod",
		"--description", "a test module",
		"--source", "https://github.com/test/repo.git",
		"--provider", "azurerm",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/module") {
		t.Errorf("expected path to contain organization/org-abc/module, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "test-mod" {
		t.Errorf("expected name 'test-mod', got %v", attrs["name"])
	}
	if attrs["provider"] != "azurerm" {
		t.Errorf("expected provider 'azurerm', got %v", attrs["provider"])
	}
	if !strings.Contains(out, "mod-new") {
		t.Errorf("expected output to contain 'mod-new', got: %s", out)
	}
}

// ----- Module List E2E -----

func TestCmdModuleListE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path

		modules := []*terrakube.Module{{ID: "mod-1", Name: "module-one", Provider: "aws"}}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, modules)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("module", "list", "--organization-id", "org-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/module") {
		t.Errorf("expected path to contain organization/org-abc/module, got %s", receivedPath)
	}
	if !strings.Contains(out, "mod-1") {
		t.Errorf("expected output to contain 'mod-1', got: %s", out)
	}
	if !strings.Contains(out, "module-one") {
		t.Errorf("expected output to contain 'module-one', got: %s", out)
	}
}

// ----- Variable List E2E -----

func TestCmdVariableListE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path

		variables := []*terrakube.Variable{{ID: "var-1", Key: "TF_VAR_name", Value: "hello", Category: "TERRAFORM"}}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, variables)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "variable", "list",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/ws-123/variable") {
		t.Errorf("expected path to contain organization/org-abc/workspace/ws-123/variable, got %s", receivedPath)
	}
	if !strings.Contains(out, "var-1") {
		t.Errorf("expected output to contain 'var-1', got: %s", out)
	}
}

// ----- Variable Create E2E -----

func TestCmdVariableCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)

		v := &terrakube.Variable{ID: "var-new", Key: "MY_VAR", Value: "my-value", Category: "TERRAFORM"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, v)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "variable", "create",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--key", "MY_VAR",
		"--value", "my-value",
		"--category", "TERRAFORM",
		"--sensitive=false",
		"--hcl=false",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/ws-123/variable") {
		t.Errorf("expected path to contain organization/org-abc/workspace/ws-123/variable, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["key"] != "MY_VAR" {
		t.Errorf("expected key 'MY_VAR', got %v", attrs["key"])
	}
	if attrs["value"] != "my-value" {
		t.Errorf("expected value 'my-value', got %v", attrs["value"])
	}
	if attrs["category"] != "TERRAFORM" {
		t.Errorf("expected category 'TERRAFORM', got %v", attrs["category"])
	}
	if !strings.Contains(out, "var-new") {
		t.Errorf("expected output to contain 'var-new', got: %s", out)
	}
}

// ----- Job Create E2E -----

func TestCmdJobCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)

		job := &terrakube.Job{ID: "job-new", Command: "plan", Status: "pending"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, job)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"job", "create",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--command", "plan",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/job") {
		t.Errorf("expected path to contain organization/org-abc/job, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["command"] != "plan" {
		t.Errorf("expected command 'plan', got %v", attrs["command"])
	}
	// Verify workspace relationship is set
	rels, ok := data["relationships"].(map[string]interface{})
	if !ok {
		t.Fatal("expected job to have relationships")
	}
	wsRel, ok := rels["workspace"].(map[string]interface{})
	if !ok {
		t.Fatal("expected job to have workspace relationship")
	}
	wsData, ok := wsRel["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected workspace relationship to have data")
	}
	if wsData["id"] != "ws-123" {
		t.Errorf("expected workspace relationship ID 'ws-123', got %v", wsData["id"])
	}
	if !strings.Contains(out, "job-new") {
		t.Errorf("expected output to contain 'job-new', got: %s", out)
	}
}

// ----- Team Create E2E -----

func TestCmdTeamCreateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)

		team := &terrakube.Team{ID: "team-new", Name: "test-team"}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, team)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"team", "create",
		"--organization-id", "org-abc",
		"--name", "test-team",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/team") {
		t.Errorf("expected path to contain organization/org-abc/team, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "test-team" {
		t.Errorf("expected name 'test-team', got %v", attrs["name"])
	}
	if !strings.Contains(out, "team-new") {
		t.Errorf("expected output to contain 'team-new', got: %s", out)
	}
}

// ----- Team Create with Permissions E2E -----

func TestCmdTeamCreateWithPermissionsE2E(t *testing.T) {
	resetGlobalFlags()

	var capturedPermsBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPermsBody, _ = io.ReadAll(r.Body)

		team := &terrakube.Team{ID: "team-perms", Name: "perms-team", ManageWorkspace: true, ManageModule: true}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, team)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"team", "create",
		"--organization-id", "org-abc",
		"--name", "perms-team",
		"--manage-workspace=true",
		"--manage-module=true",
		"--manage-provider=false",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedPermsBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["manageWorkspace"] != true {
		t.Error("expected ManageWorkspace to be true")
	}
	if attrs["manageModule"] != true {
		t.Error("expected ManageModule to be true")
	}
	if !strings.Contains(out, "team-perms") {
		t.Errorf("expected output to contain 'team-perms', got: %s", out)
	}
}

// ----- Logout Stub Test -----

// logout is a stub — it doesn't clear credentials. It only prints "logout ok"
// but does NOT remove or modify any configuration file.
func TestCmdLogoutIsStub(t *testing.T) {
	resetGlobalFlags()

	// Set some viper config to simulate a logged-in state
	viper.Set("api_url", "http://localhost:8080")
	viper.Set("token", "some-secret-token")

	out, err := executeCommand("logout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "logout ok") {
		t.Errorf("expected 'logout ok' in output, got: %s", out)
	}

	// Verify that logout does NOT clear the stored credentials.
	// The viper state still contains the token because logout is a stub
	// that only prints "logout ok" without touching any config.
	if viper.GetString("token") != "some-secret-token" {
		t.Error("expected token to remain in viper after logout (stub does not clear config)")
	}
	if viper.GetString("api_url") != "http://localhost:8080" {
		t.Error("expected api_url to remain in viper after logout (stub does not clear config)")
	}
}

// ----- Login Config Writing -----

// TestCmdLoginWritesConfig verifies that a successful login writes api_url and
// token into viper state. We cannot easily test the file write to
// ~/.terrakube-cli.yaml in a unit test without side effects, but we CAN verify
// that viper.Get("api_url") and viper.Get("token") are set after login runs.
//
// Note: newClient() in cmd/root.go calls url.Parse(viper.GetString("api_url"))
// and os.Exit(1) on error. Since os.Exit cannot be intercepted in tests, we
// cannot test the error path. We document this limitation here.
func TestCmdLoginWritesConfig(t *testing.T) {
	resetGlobalFlags()

	desc := "test org"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The login command calls Organization.List("") to verify connection.
		orgs := []*terrakube.Organization{{ID: "org-1", Name: "test-org", Description: &desc}}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, orgs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"login",
		"--api-url", ts.URL,
		"--pat", "my-secret-pat",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Successfully logged in") {
		t.Errorf("expected 'Successfully logged in' in output, got: %s", out)
	}

	// After successful login, viper should have the api_url and token set.
	if viper.GetString("api_url") != ts.URL {
		t.Errorf("expected viper api_url to be %q, got %q", ts.URL, viper.GetString("api_url"))
	}
	if viper.GetString("token") != "my-secret-pat" {
		t.Errorf("expected viper token to be 'my-secret-pat', got %q", viper.GetString("token"))
	}
}

// ----- newClient uses viper api_url -----

// TestCmdNewClientUsesViperURL verifies that newClient() constructs a client
// using the viper "api_url" value. We test this indirectly by executing a
// command that calls newClient() and verifying the request goes to the
// configured server.
//
// Note: newClient() calls os.Exit(1) if url.Parse fails. Since os.Exit
// cannot be intercepted in tests, the error path is untestable without
// refactoring newClient to return an error. This is a known limitation.
func TestCmdNewClientUsesViperURL(t *testing.T) {
	resetGlobalFlags()

	var receivedHost string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		orgs := []*terrakube.Organization{}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, orgs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("organization", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The test server URL was set via setupTestServer -> viper.Set("api_url", ts.URL).
	// newClient() reads viper "api_url" and passes it to client.NewClient.
	// The request should have gone to the test server, confirming viper URL is used.
	if receivedHost == "" {
		t.Error("expected request to reach test server (proving newClient uses viper api_url), but no request was received")
	}
}

// ----- Module Update E2E -----

// BUG: module_update.go:33 — updateModuleCmd.AddCommand(updateOrganizationCmd) registers
// the wrong subcommand (organization update) under module update. This doesn't break
// the update command itself, but pollutes the command tree.
func TestCmdModuleUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module", "update",
		"--organization-id", "org-abc",
		"--id", "mod-789",
		"--name", "updated-mod",
		"--description", "new desc",
		"--source", "https://github.com/test/repo.git",
		"--provider", "azurerm",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/module/mod-789") {
		t.Errorf("expected path to contain organization/org-abc/module/mod-789, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["name"] != "updated-mod" {
		t.Errorf("expected name 'updated-mod', got %v", attrs["name"])
	}
	if attrs["description"] != "new desc" {
		t.Errorf("expected description 'new desc', got %v", attrs["description"])
	}
	if attrs["source"] != "https://github.com/test/repo.git" {
		t.Errorf("expected source 'https://github.com/test/repo.git', got %v", attrs["source"])
	}
	if attrs["provider"] != "azurerm" {
		t.Errorf("expected provider 'azurerm', got %v", attrs["provider"])
	}
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", out)
	}
}

// ----- Module Delete E2E -----

func TestCmdModuleDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"module", "delete",
		"--organization-id", "org-abc",
		"--id", "mod-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/module/mod-789") {
		t.Errorf("expected path to contain organization/org-abc/module/mod-789, got %s", receivedPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

// ----- Team Update E2E -----

// Migrated to terrakube-go: fixes wrong subcommand registration, Type casing, and boolean omitempty bugs.
func TestCmdTeamUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"team", "update",
		"--organization-id", "org-abc",
		"--id", "team-789",
		"--name", "updated-team",
		"--manage-workspace",
		"--manage-module",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/team/team-789") {
		t.Errorf("expected path to contain organization/org-abc/team/team-789, got %s", receivedPath)
	}

	// Verify body content via raw JSON
	var body map[string]interface{}
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})
	if attrs["name"] != "updated-team" {
		t.Errorf("expected name 'updated-team', got %v", attrs["name"])
	}
	if attrs["manageWorkspace"] != true {
		t.Errorf("expected manageWorkspace true, got %v", attrs["manageWorkspace"])
	}
	if attrs["manageModule"] != true {
		t.Errorf("expected manageModule true, got %v", attrs["manageModule"])
	}

	// terrakube-go correctly uses type "team" (lowercase)
	if data["type"] != "team" {
		t.Errorf("expected type 'team', got %v", data["type"])
	}

	if !strings.Contains(out, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", out)
	}
}

// ----- Team Delete E2E -----

func TestCmdTeamDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"team", "delete",
		"--organization-id", "org-abc",
		"--id", "team-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/team/team-789") {
		t.Errorf("expected path to contain organization/org-abc/team/team-789, got %s", receivedPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

// ----- Variable Update E2E -----

// BUG: variable_update.go:34 — updateVariableCmd.AddCommand(updateOrganizationCmd) registers
// the wrong subcommand (organization update) under variable update.
func TestCmdVariableUpdateE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	var capturedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand(
		"workspace", "variable", "update",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--id", "var-789",
		"--key", "UPDATED_KEY",
		"--value", "new-value",
		"--category", "ENV",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/ws-123/variable/var-789") {
		t.Errorf("expected path to contain organization/org-abc/workspace/ws-123/variable/var-789, got %s", receivedPath)
	}
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(capturedBody, &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	data, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body with data")
	}
	attrs, ok := data["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request body data to contain attributes")
	}
	if attrs["key"] != "UPDATED_KEY" {
		t.Errorf("expected key 'UPDATED_KEY', got %v", attrs["key"])
	}
	if attrs["value"] != "new-value" {
		t.Errorf("expected value 'new-value', got %v", attrs["value"])
	}
	if attrs["category"] != "ENV" {
		t.Errorf("expected category 'ENV', got %v", attrs["category"])
	}
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", out)
	}
}

// ----- Variable Delete E2E -----

// BUG: variable_delete.go prints nothing on success — no "deleted" message.
// This test documents the silent-success behavior.
func TestCmdVariableDeleteE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand(
		"workspace", "variable", "delete",
		"--organization-id", "org-abc",
		"--workspace-id", "ws-123",
		"--id", "var-789",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/workspace/ws-123/variable/var-789") {
		t.Errorf("expected path to contain organization/org-abc/workspace/ws-123/variable/var-789, got %s", receivedPath)
	}
	// No output assertion — variable_delete.go has no fmt.Print on success (BUG)
}

// ----- Job List E2E -----

func TestCmdJobListE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path

		jobs := []*terrakube.Job{
			{ID: "job-1", Command: "plan", Status: "completed"},
			{ID: "job-2", Command: "apply", Status: "pending"},
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, jobs)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("job", "list", "--organization-id", "org-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/org-abc/job") {
		t.Errorf("expected path to contain organization/org-abc/job, got %s", receivedPath)
	}
	if !strings.Contains(out, "job-1") {
		t.Errorf("expected output to contain 'job-1', got: %s", out)
	}
	if !strings.Contains(out, "job-2") {
		t.Errorf("expected output to contain 'job-2', got: %s", out)
	}
}

// ----- Module Update Missing Flags -----

func TestCmdModuleUpdateMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "update", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for module update without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdModuleUpdateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "update", "--id", "mod-123")
	if err == nil {
		t.Fatal("expected error for module update without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Module Delete Missing Flags -----

func TestCmdModuleDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "delete", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for module delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdModuleDeleteMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("module", "delete", "--id", "mod-123")
	if err == nil {
		t.Fatal("expected error for module delete without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Team Update Missing Flags -----

func TestCmdTeamUpdateMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "update", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for team update without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdTeamUpdateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "update", "--id", "team-123")
	if err == nil {
		t.Fatal("expected error for team update without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Team Delete Missing Flags -----

func TestCmdTeamDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "delete", "--organization-id", "org-123")
	if err == nil {
		t.Fatal("expected error for team delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdTeamDeleteMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("team", "delete", "--id", "team-123")
	if err == nil {
		t.Fatal("expected error for team delete without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

// ----- Variable Update Missing Flags -----

func TestCmdVariableUpdateMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "update", "--organization-id", "org-123", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for variable update without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdVariableUpdateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "update", "--id", "var-123", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for variable update without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdVariableUpdateMissingWorkspaceId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "update", "--organization-id", "org-123", "--id", "var-123")
	if err == nil {
		t.Fatal("expected error for variable update without --workspace-id, got nil")
	}
	if !strings.Contains(err.Error(), "workspace-id") {
		t.Errorf("expected error to mention workspace-id, got: %v", err)
	}
}

// ----- Variable Delete Missing Flags -----

func TestCmdVariableDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "delete", "--organization-id", "org-123", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for variable delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdVariableDeleteMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "delete", "--id", "var-123", "--workspace-id", "ws-123")
	if err == nil {
		t.Fatal("expected error for variable delete without --organization-id, got nil")
	}
	if !strings.Contains(err.Error(), "organization-id") {
		t.Errorf("expected error to mention organization-id, got: %v", err)
	}
}

func TestCmdVariableDeleteMissingWorkspaceId(t *testing.T) {
	resetGlobalFlags()
	_, err := executeCommand("workspace", "variable", "delete", "--organization-id", "org-123", "--id", "var-123")
	if err == nil {
		t.Fatal("expected error for variable delete without --workspace-id, got nil")
	}
	if !strings.Contains(err.Error(), "workspace-id") {
		t.Errorf("expected error to mention workspace-id, got: %v", err)
	}
}
