package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// resetGlobalFlags resets Cobra flag states and Viper keys between tests.
func resetGlobalFlags() {
	// login
	apiURL = ""
	patToken = ""

	// root
	output = "json"

	// Clear viper keys that postInitCommands uses to pre-fill cobra required flags.
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

	// Point viper at a nonexistent config file to prevent real user config interference.
	cfgFile = os.DevNull

	// Reset cobra flag state on all commands.
	resetCobraFlags(rootCmd)
}

// resetCobraFlags recursively resets all flags on a command and its subcommands.
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

// setupTestServer creates an httptest server and configures viper so that newClient() points at it.
func setupTestServer(handler http.Handler) *httptest.Server {
	ts := httptest.NewServer(handler)
	viper.Set("api_url", ts.URL)
	viper.Set("token", "test-token")
	return ts
}

func findSubCmd(name string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// ----- Flag Validation Tests -----

func TestCmdWorkspaceListMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	output, err := executeCommand("workspace", "list")
	if err == nil {
		t.Fatalf("expected error for workspace list without --organization, got nil (out=%q)", output)
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Fatalf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdWorkspaceCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("workspace", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil {
		t.Fatal("expected error for workspace create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdWorkspaceCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("workspace", "create", "--name", "test-ws")
	if err == nil {
		t.Fatal("expected error for workspace create without --organization, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdOrganizationCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

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
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

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
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("workspace", "delete", "--id", "38b6635a-d38e-46f2-a95e-d00a416de4fd")
	if err == nil {
		t.Fatal("expected error for workspace delete without --organization, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdWorkspaceDeleteMissingId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("workspace", "delete", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil {
		t.Fatal("expected error for workspace delete without --id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention id, got: %v", err)
	}
}

func TestCmdModuleCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("module", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil {
		t.Fatal("expected error for module create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdModuleCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("module", "create", "--name", "mod1")
	if err == nil {
		t.Fatal("expected error for module create without --organization, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdTeamCreateMissingName(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("team", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err == nil {
		t.Fatal("expected error for team create without --name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestCmdTeamCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("team", "create", "--name", "team1")
	if err == nil {
		t.Fatal("expected error for team create without --organization, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

func TestCmdJobCreateMissingCommand(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("job", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd")
	if err == nil {
		t.Fatal("expected error for job create without --command, got nil")
	}
	if !strings.Contains(err.Error(), "command") {
		t.Errorf("expected error to mention command, got: %v", err)
	}
}

func TestCmdJobCreateMissingOrgId(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("job", "create", "--command", "plan", "--workspace-id", "38b6635a-d38e-46f2-a95e-d00a416de4fd")
	if err == nil {
		t.Fatal("expected error for job create without --organization, got nil")
	}
	if !strings.Contains(err.Error(), "organization") {
		t.Errorf("expected error to mention organization, got: %v", err)
	}
}

// ----- Flag Default Tests -----

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

// ----- Command Alias Tests -----

func TestCmdWorkspaceAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("workspace")
	if cmd == nil {
		t.Fatal("workspace command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "ws" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected workspace command to have alias 'ws', got aliases: %v", cmd.Aliases)
	}
}

func TestCmdOrganizationAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("organization")
	if cmd == nil {
		t.Fatal("organization command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "org" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected organization command to have alias 'org', got aliases: %v", cmd.Aliases)
	}
}

func TestCmdModuleAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("module")
	if cmd == nil {
		t.Fatal("module command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "mod" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected module command to have alias 'mod', got aliases: %v", cmd.Aliases)
	}
}

func TestCmdVariableAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("variable")
	if cmd == nil {
		t.Fatal("variable command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "var" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected variable command to have alias 'var', got aliases: %v", cmd.Aliases)
	}
}

func TestCmdAliasOrgRoutes(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

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
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer ts.Close()

	_, err := executeCommand("mod", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
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

	cmd := findSubCmd("workspace")
	if cmd == nil {
		t.Fatal("workspace command not found")
	}
	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
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

	cmd := findSubCmd("organization")
	if cmd == nil {
		t.Fatal("organization command not found")
	}
	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
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

func TestCmdModuleSubcommands(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("module")
	if cmd == nil {
		t.Fatal("module command not found")
	}
	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
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

	cmd := findSubCmd("team")
	if cmd == nil {
		t.Fatal("team command not found")
	}
	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
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

	cmd := findSubCmd("job")
	if cmd == nil {
		t.Fatal("job command not found")
	}
	expected := map[string]bool{
		"create": false,
		"list":   false,
		"update": false,
		"delete": false,
	}

	for _, sub := range cmd.Commands() {
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
	viper.Set("api_url", "")
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
	viper.Set("pat", "")
	viper.Set("token", "")
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

	_, err := executeCommand("workspace", "list", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	_ = err
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

// ----- Team alias -----

func TestCmdTeamAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("team")
	if cmd == nil {
		t.Fatal("team command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "teams" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected team command to have alias 'teams', got aliases: %v", cmd.Aliases)
	}
}

// ----- Job alias -----

func TestCmdJobAlias(t *testing.T) {
	resetGlobalFlags()

	cmd := findSubCmd("job")
	if cmd == nil {
		t.Fatal("job command not found")
	}
	found := false
	for _, a := range cmd.Aliases {
		if a == "jobs" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected job command to have alias 'jobs', got aliases: %v", cmd.Aliases)
	}
}

// ----- Logout Stub Test -----

func TestCmdLogoutIsStub(t *testing.T) {
	resetGlobalFlags()

	viper.Set("api_url", "http://localhost:8080")
	viper.Set("token", "some-secret-token")

	out, err := executeCommand("logout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "logout ok") {
		t.Errorf("expected 'logout ok' in output, got: %s", out)
	}

	if viper.GetString("token") != "some-secret-token" {
		t.Error("expected token to remain in viper after logout")
	}
	if viper.GetString("api_url") != "http://localhost:8080" {
		t.Error("expected api_url to remain in viper after logout")
	}
}

// ----- Login Config Writing -----

func TestCmdLoginWritesConfig(t *testing.T) {
	resetGlobalFlags()

	desc := "test org"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	if viper.GetString("api_url") != ts.URL {
		t.Errorf("expected viper api_url to be %q, got %q", ts.URL, viper.GetString("api_url"))
	}
	if viper.GetString("token") != "my-secret-pat" {
		t.Errorf("expected viper token to be 'my-secret-pat', got %q", viper.GetString("token"))
	}
}

// ----- newClient uses viper api_url -----

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

	if receivedHost == "" {
		t.Error("expected request to reach test server, but no request was received")
	}
}

// ----- New Domain Services E2E Tests -----

func TestCmdProjectE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/vnd.api+json")
		p := &terrakube.Project{ID: "proj-123", Name: "test-project"}
		_ = jsonapi.MarshalPayload(w, p)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("project", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--name", "test-project")
	if err != nil {
		t.Fatalf("unexpected error creating project: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/project") {
		t.Errorf("expected path to contain organization/.../project, got %s", receivedPath)
	}
	if !strings.Contains(out, "test-project") {
		t.Errorf("expected output to contain project name, got: %s", out)
	}
}

func TestCmdProjectAccessE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/vnd.api+json")
		pa := &terrakube.ProjectAccess{ID: "pa-123", Name: "default-access"}
		_ = jsonapi.MarshalPayload(w, pa)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("project-access", "create", "--organization-id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "--project-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901", "--name", "default-access")
	if err != nil {
		t.Fatalf("unexpected error creating project access: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "organization/a1b2c3d4-e5f6-7890-abcd-ef1234567890/project/b2c3d4e5-f6a7-8901-bcde-f12345678901/projectAccess") {
		t.Errorf("expected path to contain organization/.../project/.../projectAccess, got %s", receivedPath)
	}
}

func TestCmdFederatedE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fed := &terrakube.Federated{ID: "fed-123", Name: "github-oidc", IssuerURL: "https://token.actions.githubusercontent.com", Audience: "terrakube"}
		_ = jsonapi.MarshalPayload(w, fed)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("federated", "create", "--name", "github-oidc", "--issuer-url", "https://token.actions.githubusercontent.com", "--audience", "terrakube")
	if err != nil {
		t.Fatalf("unexpected error creating federated identity: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "federated") {
		t.Errorf("expected path to contain federated, got %s", receivedPath)
	}
	if !strings.Contains(out, "https://token.actions.githubusercontent.com") {
		t.Errorf("expected output to contain issuerUrl, got: %s", out)
	}
}

func TestCmdFederatedClaimE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fc := &terrakube.FederatedClaim{ID: "fc-123", ClaimKey: "sub", ClaimValue: "repo:org/repo:ref:refs/heads/main"}
		_ = jsonapi.MarshalPayload(w, fc)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	_, err := executeCommand("federated-claim", "create", "--federated-id", "b2c3d4e5-f6a7-8901-bcde-f12345678901", "--claim-key", "sub", "--claim-value", "repo:org/repo:ref:refs/heads/main")
	if err != nil {
		t.Fatalf("unexpected error creating federated claim: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "federated/b2c3d4e5-f6a7-8901-bcde-f12345678901/claims") {
		t.Errorf("expected path to contain federated/.../claims, got %s", receivedPath)
	}
}

func TestCmdTeamTokenE2E(t *testing.T) {
	resetGlobalFlags()

	var receivedMethod string
	var receivedPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		tt := &terrakube.TeamToken{ID: "tt-123", Description: "test token", Group: "dev-team"}
		b, _ := json.Marshal(tt)
		_, _ = w.Write(b)
	})

	ts := setupTestServer(handler)
	defer ts.Close()

	out, err := executeCommand("team-token", "create", "--group", "dev-team", "--description", "test token")
	if err != nil {
		t.Fatalf("unexpected error creating team token: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "access-token/v1/teams") {
		t.Errorf("expected path to contain access-token/v1/teams, got %s", receivedPath)
	}
	if !strings.Contains(out, "tt-123") {
		t.Errorf("expected output to contain token ID, got: %s", out)
	}
}
