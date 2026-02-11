package models

import (
	"encoding/json"
	"strings"
	"testing"
)

// helpers

func strPtr(s string) *string { return &s }

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	return b
}

func jsonContainsKey(t *testing.T, data []byte, key string) {
	t.Helper()
	if !strings.Contains(string(data), `"`+key+`"`) {
		t.Errorf("expected JSON key %q in output: %s", key, data)
	}
}

func jsonLacksKey(t *testing.T, data []byte, key string) {
	t.Helper()
	if strings.Contains(string(data), `"`+key+`"`) {
		t.Errorf("did not expect JSON key %q in output: %s", key, data)
	}
}

// ---------------------------------------------------------------------------
// Organization
// ---------------------------------------------------------------------------

func fullOrganization() *Organization {
	return &Organization{
		ID:   "org-123",
		Type: "organization",
		Attributes: &OrganizationAttributes{
			Name:          "my-org",
			Description:   strPtr("Org description"),
			ExecutionMode: strPtr("remote"),
			Icon:          strPtr("https://example.com/icon.png"),
		},
		Relationships: &OrganizationRelationships{
			Job:       &OrganizationRelationshipsJob{ID: "job-1", Type: "job"},
			Module:    &OrganizationRelationshipsModule{ID: "mod-1", Type: "module"},
			Workspace: &OrganizationRelationshipsWorkspace{ID: "ws-1", Type: "workspace"},
		},
	}
}

func TestOrganization_RoundTrip(t *testing.T) {
	orig := fullOrganization()
	data := mustMarshal(t, orig)

	var got Organization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	if got.Attributes.Name != orig.Attributes.Name {
		t.Errorf("Name: got %q, want %q", got.Attributes.Name, orig.Attributes.Name)
	}
	if *got.Attributes.Description != *orig.Attributes.Description {
		t.Errorf("Description: got %q, want %q", *got.Attributes.Description, *orig.Attributes.Description)
	}
	if *got.Attributes.ExecutionMode != *orig.Attributes.ExecutionMode {
		t.Errorf("ExecutionMode: got %q, want %q", *got.Attributes.ExecutionMode, *orig.Attributes.ExecutionMode)
	}
	if *got.Attributes.Icon != *orig.Attributes.Icon {
		t.Errorf("Icon: got %q, want %q", *got.Attributes.Icon, *orig.Attributes.Icon)
	}
	if got.Relationships.Job.ID != orig.Relationships.Job.ID {
		t.Errorf("Job.ID: got %q, want %q", got.Relationships.Job.ID, orig.Relationships.Job.ID)
	}
	if got.Relationships.Module.ID != orig.Relationships.Module.ID {
		t.Errorf("Module.ID: got %q, want %q", got.Relationships.Module.ID, orig.Relationships.Module.ID)
	}
	if got.Relationships.Workspace.ID != orig.Relationships.Workspace.ID {
		t.Errorf("Workspace.ID: got %q, want %q", got.Relationships.Workspace.ID, orig.Relationships.Workspace.ID)
	}
}

func TestOrganization_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullOrganization())
	for _, key := range []string{
		"id", "type", "attributes", "relationships",
		"name", "description", "executionMode", "icon",
		"job", "module", "workspace",
	} {
		jsonContainsKey(t, data, key)
	}
	for _, bad := range []string{"execution_mode", "ExecutionMode", "Name", "Description", "Icon"} {
		jsonLacksKey(t, data, bad)
	}
}

func TestOrganization_OmitEmpty(t *testing.T) {
	org := &Organization{ID: "org-1"}
	data := mustMarshal(t, org)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "relationships")
	jsonLacksKey(t, data, "type")
}

func TestOrganizationAttributes_OmitEmpty(t *testing.T) {
	attrs := &OrganizationAttributes{Name: "only-name"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "name")
	jsonLacksKey(t, data, "description")
	jsonLacksKey(t, data, "executionMode")
	jsonLacksKey(t, data, "icon")
}

func TestOrganization_GetBody(t *testing.T) {
	body := GetBodyOrganization{
		Data: []*Organization{fullOrganization(), {ID: "org-456", Type: "organization"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyOrganization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 organizations, got %d", len(got.Data))
	}
	if got.Data[0].ID != "org-123" {
		t.Errorf("first org ID: got %q, want %q", got.Data[0].ID, "org-123")
	}
	if got.Data[1].ID != "org-456" {
		t.Errorf("second org ID: got %q, want %q", got.Data[1].ID, "org-456")
	}
	jsonContainsKey(t, data, "data")
}

func TestOrganization_PostBody(t *testing.T) {
	body := PostBodyOrganization{Data: fullOrganization()}
	data := mustMarshal(t, body)

	var got PostBodyOrganization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "org-123" {
		t.Errorf("ID: got %q, want %q", got.Data.ID, "org-123")
	}
	jsonContainsKey(t, data, "data")
}

func TestOrganization_GetBodyEmptySlice(t *testing.T) {
	body := GetBodyOrganization{Data: []*Organization{}}
	data := mustMarshal(t, body)

	var got GetBodyOrganization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 0 {
		t.Errorf("expected 0 organizations, got %d", len(got.Data))
	}
}

func TestOrganization_PostBodyNilData(t *testing.T) {
	body := PostBodyOrganization{Data: nil}
	data := mustMarshal(t, body)

	var got PostBodyOrganization
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data != nil {
		t.Errorf("expected nil Data, got %+v", got.Data)
	}
}

func TestOrganizationRelationships_RoundTrip(t *testing.T) {
	rels := &OrganizationRelationships{
		Job:       &OrganizationRelationshipsJob{ID: "j-1", Type: "job"},
		Module:    &OrganizationRelationshipsModule{ID: "m-1", Type: "module"},
		Workspace: &OrganizationRelationshipsWorkspace{ID: "w-1", Type: "workspace"},
	}
	data := mustMarshal(t, rels)

	var got OrganizationRelationships
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Job.ID != "j-1" || got.Job.Type != "job" {
		t.Errorf("Job: got %+v", got.Job)
	}
	if got.Module.ID != "m-1" || got.Module.Type != "module" {
		t.Errorf("Module: got %+v", got.Module)
	}
	if got.Workspace.ID != "w-1" || got.Workspace.Type != "workspace" {
		t.Errorf("Workspace: got %+v", got.Workspace)
	}
}

// ---------------------------------------------------------------------------
// Workspace
// ---------------------------------------------------------------------------

func fullWorkspace() *Workspace {
	return &Workspace{
		ID:   "ws-789",
		Type: "workspace",
		Attributes: &WorkspaceAttributes{
			Name:             "my-workspace",
			Description:      "A test workspace",
			Source:           "https://github.com/example/repo",
			Folder:           "/infra",
			ExecutionMode:    "remote",
			Branch:           "main",
			IacType:          "terraform",
			TerraformVersion: "1.5.0",
		},
		Relationships: &WorkspaceRelationships{
			Job:          &WorkspaceRelationshipsJob{ID: "job-10", Type: "job"},
			Organization: &WorkspaceRelationshipsOrganization{ID: "org-5", Type: "organization"},
			Variable:     &WorkspaceRelationshipsVariable{ID: "var-3", Type: "variable"},
		},
	}
}

func TestWorkspace_RoundTrip(t *testing.T) {
	orig := fullWorkspace()
	data := mustMarshal(t, orig)

	var got Workspace
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	a := got.Attributes
	oa := orig.Attributes
	if a.Name != oa.Name {
		t.Errorf("Name: got %q, want %q", a.Name, oa.Name)
	}
	if a.Description != oa.Description {
		t.Errorf("Description: got %q, want %q", a.Description, oa.Description)
	}
	if a.Source != oa.Source {
		t.Errorf("Source: got %q, want %q", a.Source, oa.Source)
	}
	if a.Folder != oa.Folder {
		t.Errorf("Folder: got %q, want %q", a.Folder, oa.Folder)
	}
	if a.ExecutionMode != oa.ExecutionMode {
		t.Errorf("ExecutionMode: got %q, want %q", a.ExecutionMode, oa.ExecutionMode)
	}
	if a.Branch != oa.Branch {
		t.Errorf("Branch: got %q, want %q", a.Branch, oa.Branch)
	}
	if a.IacType != oa.IacType {
		t.Errorf("IacType: got %q, want %q", a.IacType, oa.IacType)
	}
	if a.TerraformVersion != oa.TerraformVersion {
		t.Errorf("TerraformVersion: got %q, want %q", a.TerraformVersion, oa.TerraformVersion)
	}
	r := got.Relationships
	if r.Job.ID != "job-10" || r.Job.Type != "job" {
		t.Errorf("Job rel: got %+v", r.Job)
	}
	if r.Organization.ID != "org-5" || r.Organization.Type != "organization" {
		t.Errorf("Org rel: got %+v", r.Organization)
	}
	if r.Variable.ID != "var-3" || r.Variable.Type != "variable" {
		t.Errorf("Var rel: got %+v", r.Variable)
	}
}

func TestWorkspace_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullWorkspace())
	for _, key := range []string{
		"id", "type", "attributes", "relationships",
		"name", "description", "source", "folder",
		"executionMode", "branch", "iacType", "terraformVersion",
		"job", "organization", "variable",
	} {
		jsonContainsKey(t, data, key)
	}
	for _, bad := range []string{
		"execution_mode", "iac_type", "terraform_version",
		"ExecutionMode", "IacType", "TerraformVersion",
	} {
		jsonLacksKey(t, data, bad)
	}
}

func TestWorkspace_OmitEmpty(t *testing.T) {
	ws := &Workspace{ID: "ws-1"}
	data := mustMarshal(t, ws)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "relationships")
	jsonLacksKey(t, data, "type")
}

func TestWorkspaceAttributes_OmitEmpty(t *testing.T) {
	attrs := &WorkspaceAttributes{Name: "only-name"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "name")
	for _, key := range []string{
		"description", "source", "folder",
		"executionMode", "branch", "iacType", "terraformVersion",
	} {
		jsonLacksKey(t, data, key)
	}
}

func TestWorkspace_GetBody(t *testing.T) {
	body := GetBodyWorkspace{
		Data: []*Workspace{fullWorkspace(), {ID: "ws-2", Type: "workspace"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyWorkspace
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(got.Data))
	}
	if got.Data[0].ID != "ws-789" {
		t.Errorf("first ws ID: got %q", got.Data[0].ID)
	}
	if got.Data[1].ID != "ws-2" {
		t.Errorf("second ws ID: got %q", got.Data[1].ID)
	}
}

func TestWorkspace_PostBody(t *testing.T) {
	body := PostBodyWorkspace{Data: fullWorkspace()}
	data := mustMarshal(t, body)

	var got PostBodyWorkspace
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "ws-789" {
		t.Errorf("ID: got %q", got.Data.ID)
	}
}

func TestWorkspaceRelationships_RoundTrip(t *testing.T) {
	rels := &WorkspaceRelationships{
		Job:          &WorkspaceRelationshipsJob{ID: "j-2", Type: "job"},
		Organization: &WorkspaceRelationshipsOrganization{ID: "o-2", Type: "organization"},
		Variable:     &WorkspaceRelationshipsVariable{ID: "v-2", Type: "variable"},
	}
	data := mustMarshal(t, rels)

	var got WorkspaceRelationships
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Job.ID != "j-2" {
		t.Errorf("Job.ID: got %q", got.Job.ID)
	}
	if got.Organization.ID != "o-2" {
		t.Errorf("Organization.ID: got %q", got.Organization.ID)
	}
	if got.Variable.ID != "v-2" {
		t.Errorf("Variable.ID: got %q", got.Variable.ID)
	}
}

// ---------------------------------------------------------------------------
// Module
// ---------------------------------------------------------------------------

func fullModule() *Module {
	return &Module{
		ID:   "mod-42",
		Type: "module",
		Attributes: &ModuleAttributes{
			Name:         "vpc",
			Description:  "VPC module",
			Provider:     "aws",
			Source:       "https://github.com/example/terraform-aws-vpc",
			TagPrefix:    strPtr("v"),
			Folder:       strPtr("/modules/vpc"),
			RegistryPath: "aws/vpc/aws",
			Versions:     []string{"1.0.0", "1.1.0", "2.0.0"},
		},
		Relationships: &ModuleRelationships{
			Definition:   &ModuleRelationshipsDefinition{ID: "def-1", Type: "definition"},
			Organization: &ModuleRelationshipsOrganization{ID: "org-1", Type: "organization"},
		},
	}
}

func TestModule_RoundTrip(t *testing.T) {
	orig := fullModule()
	data := mustMarshal(t, orig)

	var got Module
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	a := got.Attributes
	oa := orig.Attributes
	if a.Name != oa.Name {
		t.Errorf("Name: got %q, want %q", a.Name, oa.Name)
	}
	if a.Description != oa.Description {
		t.Errorf("Description: got %q, want %q", a.Description, oa.Description)
	}
	if a.Provider != oa.Provider {
		t.Errorf("Provider: got %q, want %q", a.Provider, oa.Provider)
	}
	if a.Source != oa.Source {
		t.Errorf("Source: got %q, want %q", a.Source, oa.Source)
	}
	if *a.TagPrefix != *oa.TagPrefix {
		t.Errorf("TagPrefix: got %q, want %q", *a.TagPrefix, *oa.TagPrefix)
	}
	if *a.Folder != *oa.Folder {
		t.Errorf("Folder: got %q, want %q", *a.Folder, *oa.Folder)
	}
	if a.RegistryPath != oa.RegistryPath {
		t.Errorf("RegistryPath: got %q, want %q", a.RegistryPath, oa.RegistryPath)
	}
	if len(a.Versions) != len(oa.Versions) {
		t.Fatalf("Versions length: got %d, want %d", len(a.Versions), len(oa.Versions))
	}
	for i, v := range a.Versions {
		if v != oa.Versions[i] {
			t.Errorf("Versions[%d]: got %q, want %q", i, v, oa.Versions[i])
		}
	}
	if got.Relationships.Definition.ID != "def-1" {
		t.Errorf("Definition.ID: got %q", got.Relationships.Definition.ID)
	}
	if got.Relationships.Organization.ID != "org-1" {
		t.Errorf("Organization.ID: got %q", got.Relationships.Organization.ID)
	}
}

func TestModule_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullModule())
	for _, key := range []string{
		"id", "type", "attributes", "relationships",
		"name", "description", "provider", "source",
		"tagPrefix", "folder", "registryPath", "versions",
		"definition", "organization",
	} {
		jsonContainsKey(t, data, key)
	}
	for _, bad := range []string{
		"tag_prefix", "registry_path", "TagPrefix", "RegistryPath",
	} {
		jsonLacksKey(t, data, bad)
	}
}

func TestModule_OmitEmpty(t *testing.T) {
	mod := &Module{ID: "mod-1"}
	data := mustMarshal(t, mod)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "relationships")
	jsonLacksKey(t, data, "type")
}

func TestModuleAttributes_OmitEmpty(t *testing.T) {
	attrs := &ModuleAttributes{Name: "minimal"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "name")
	for _, key := range []string{
		"description", "provider", "source",
		"tagPrefix", "folder", "registryPath", "versions",
	} {
		jsonLacksKey(t, data, key)
	}
}

func TestModuleAttributes_NilPointerFields(t *testing.T) {
	attrs := &ModuleAttributes{
		Name:     "test",
		Provider: "aws",
	}
	data := mustMarshal(t, attrs)

	jsonLacksKey(t, data, "tagPrefix")
	jsonLacksKey(t, data, "folder")
}

func TestModule_GetBody(t *testing.T) {
	body := GetBodyModule{
		Data: []*Module{fullModule(), {ID: "mod-99", Type: "module"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyModule
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(got.Data))
	}
	if got.Data[0].ID != "mod-42" {
		t.Errorf("first mod ID: got %q", got.Data[0].ID)
	}
	if got.Data[1].ID != "mod-99" {
		t.Errorf("second mod ID: got %q", got.Data[1].ID)
	}
}

func TestModule_PostBody(t *testing.T) {
	body := PostBodyModule{Data: fullModule()}
	data := mustMarshal(t, body)

	var got PostBodyModule
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "mod-42" {
		t.Errorf("ID: got %q", got.Data.ID)
	}
}

func TestModuleRelationships_RoundTrip(t *testing.T) {
	rels := &ModuleRelationships{
		Definition:   &ModuleRelationshipsDefinition{ID: "d-1", Type: "definition"},
		Organization: &ModuleRelationshipsOrganization{ID: "o-1", Type: "organization"},
	}
	data := mustMarshal(t, rels)

	var got ModuleRelationships
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Definition.ID != "d-1" || got.Definition.Type != "definition" {
		t.Errorf("Definition: got %+v", got.Definition)
	}
	if got.Organization.ID != "o-1" || got.Organization.Type != "organization" {
		t.Errorf("Organization: got %+v", got.Organization)
	}
}

// ---------------------------------------------------------------------------
// Variable
// ---------------------------------------------------------------------------

func fullVariable() *Variable {
	return &Variable{
		ID:   "var-55",
		Type: "variable",
		Attributes: &VariableAttributes{
			Key:         "AWS_ACCESS_KEY_ID",
			Value:       "AKIAIOSFODNN7EXAMPLE",
			Description: "AWS access key",
			Category:    "ENV",
			Sensitive:   true,
			Hcl:         false,
		},
		Relationships: &VariableRelationships{
			Workspace: &VariableRelationshipsWorkspace{ID: "ws-10", Type: "workspace"},
		},
	}
}

func TestVariable_RoundTrip(t *testing.T) {
	orig := fullVariable()
	data := mustMarshal(t, orig)

	var got Variable
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	a := got.Attributes
	oa := orig.Attributes
	if a.Key != oa.Key {
		t.Errorf("Key: got %q, want %q", a.Key, oa.Key)
	}
	if a.Value != oa.Value {
		t.Errorf("Value: got %q, want %q", a.Value, oa.Value)
	}
	if a.Description != oa.Description {
		t.Errorf("Description: got %q, want %q", a.Description, oa.Description)
	}
	if a.Category != oa.Category {
		t.Errorf("Category: got %q, want %q", a.Category, oa.Category)
	}
	if a.Sensitive != oa.Sensitive {
		t.Errorf("Sensitive: got %v, want %v", a.Sensitive, oa.Sensitive)
	}
	if a.Hcl != oa.Hcl {
		t.Errorf("Hcl: got %v, want %v", a.Hcl, oa.Hcl)
	}
	if got.Relationships.Workspace.ID != "ws-10" {
		t.Errorf("Workspace.ID: got %q", got.Relationships.Workspace.ID)
	}
}

func TestVariable_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullVariable())
	for _, key := range []string{
		"id", "type", "attributes", "relationships",
		"key", "value", "description", "category", "sensitive",
		"workspace",
	} {
		jsonContainsKey(t, data, key)
	}
}

func TestVariable_BooleanSerialization(t *testing.T) {
	v := fullVariable()
	v.Attributes.Sensitive = true
	v.Attributes.Hcl = true
	data := mustMarshal(t, v)

	raw := string(data)
	if !strings.Contains(raw, `"sensitive":true`) {
		t.Errorf("expected sensitive:true in JSON: %s", raw)
	}
	if !strings.Contains(raw, `"hcl":true`) {
		t.Errorf("expected hcl:true in JSON: %s", raw)
	}
}

func TestVariable_BooleanFalseOmitted(t *testing.T) {
	v := &Variable{
		ID: "var-1",
		Attributes: &VariableAttributes{
			Key:       "MY_VAR",
			Sensitive: false,
			Hcl:       false,
		},
	}
	data := mustMarshal(t, v)

	jsonLacksKey(t, data, "sensitive")
	jsonLacksKey(t, data, "hcl")
}

func TestVariable_OmitEmpty(t *testing.T) {
	v := &Variable{ID: "var-1"}
	data := mustMarshal(t, v)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "relationships")
}

func TestVariableAttributes_OmitEmpty(t *testing.T) {
	attrs := &VariableAttributes{Key: "MY_VAR"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "key")
	jsonLacksKey(t, data, "value")
	jsonLacksKey(t, data, "description")
	jsonLacksKey(t, data, "category")
	jsonLacksKey(t, data, "sensitive")
	jsonLacksKey(t, data, "hcl")
}

func TestVariable_GetBody(t *testing.T) {
	body := GetBodyVariable{
		Data: []*Variable{fullVariable(), {ID: "var-2", Type: "variable"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyVariable
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(got.Data))
	}
	if got.Data[0].ID != "var-55" {
		t.Errorf("first var ID: got %q", got.Data[0].ID)
	}
}

func TestVariable_PostBody(t *testing.T) {
	body := PostBodyVariable{Data: fullVariable()}
	data := mustMarshal(t, body)

	var got PostBodyVariable
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "var-55" {
		t.Errorf("ID: got %q", got.Data.ID)
	}
}

func TestVariableRelationships_RoundTrip(t *testing.T) {
	rels := &VariableRelationships{
		Workspace: &VariableRelationshipsWorkspace{ID: "ws-99", Type: "workspace"},
	}
	data := mustMarshal(t, rels)

	var got VariableRelationships
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Workspace.ID != "ws-99" || got.Workspace.Type != "workspace" {
		t.Errorf("Workspace: got %+v", got.Workspace)
	}
}

// ---------------------------------------------------------------------------
// Job
// ---------------------------------------------------------------------------

func fullJob() *Job {
	return &Job{
		ID:   "job-100",
		Type: "job",
		Attributes: &JobAttributes{
			Command: "apply",
			Output:  "Applying changes...\nDone.",
			Status:  "completed",
		},
		Relationships: &JobRelationships{
			Organization: &JobRelationshipsOrganization{ID: "org-7", Type: "organization"},
			Workspace: &JobRelationshipsWorkspace{
				Data: &JobRelationshipsWorkspaceData{ID: "ws-20", Type: "workspace"},
			},
		},
	}
}

func TestJob_RoundTrip(t *testing.T) {
	orig := fullJob()
	data := mustMarshal(t, orig)

	var got Job
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	a := got.Attributes
	oa := orig.Attributes
	if a.Command != oa.Command {
		t.Errorf("Command: got %q, want %q", a.Command, oa.Command)
	}
	if a.Output != oa.Output {
		t.Errorf("Output: got %q, want %q", a.Output, oa.Output)
	}
	if a.Status != oa.Status {
		t.Errorf("Status: got %q, want %q", a.Status, oa.Status)
	}
	if got.Relationships.Organization.ID != "org-7" {
		t.Errorf("Organization.ID: got %q", got.Relationships.Organization.ID)
	}
	if got.Relationships.Workspace.Data.ID != "ws-20" {
		t.Errorf("Workspace.Data.ID: got %q", got.Relationships.Workspace.Data.ID)
	}
	if got.Relationships.Workspace.Data.Type != "workspace" {
		t.Errorf("Workspace.Data.Type: got %q", got.Relationships.Workspace.Data.Type)
	}
}

func TestJob_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullJob())
	for _, key := range []string{
		"id", "type", "attributes", "relationships",
		"command", "output", "status",
		"organization", "workspace", "data",
	} {
		jsonContainsKey(t, data, key)
	}
}

func TestJob_OmitEmpty(t *testing.T) {
	j := &Job{ID: "job-1"}
	data := mustMarshal(t, j)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "relationships")
}

func TestJobAttributes_OmitEmpty(t *testing.T) {
	attrs := &JobAttributes{Command: "plan"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "command")
	jsonLacksKey(t, data, "output")
	jsonLacksKey(t, data, "status")
}

func TestJob_GetBody(t *testing.T) {
	body := GetBodyJob{
		Data: []*Job{fullJob(), {ID: "job-200", Type: "job"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyJob
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got.Data))
	}
	if got.Data[0].ID != "job-100" {
		t.Errorf("first job ID: got %q", got.Data[0].ID)
	}
	if got.Data[1].ID != "job-200" {
		t.Errorf("second job ID: got %q", got.Data[1].ID)
	}
}

func TestJob_PostBody(t *testing.T) {
	body := PostBodyJob{Data: fullJob()}
	data := mustMarshal(t, body)

	var got PostBodyJob
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "job-100" {
		t.Errorf("ID: got %q", got.Data.ID)
	}
}

func TestJob_WorkspaceNestedDataStructure(t *testing.T) {
	rel := &JobRelationshipsWorkspace{
		Data: &JobRelationshipsWorkspaceData{ID: "ws-nested", Type: "workspace"},
	}
	data := mustMarshal(t, rel)

	jsonContainsKey(t, data, "data")
	jsonContainsKey(t, data, "id")
	jsonContainsKey(t, data, "type")

	var got JobRelationshipsWorkspace
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data.ID != "ws-nested" {
		t.Errorf("Data.ID: got %q", got.Data.ID)
	}
}

func TestJob_WorkspaceNilData(t *testing.T) {
	rel := &JobRelationshipsWorkspace{Data: nil}
	data := mustMarshal(t, rel)

	var got JobRelationshipsWorkspace
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data != nil {
		t.Errorf("expected nil Data, got %+v", got.Data)
	}
}

func TestJobRelationships_RoundTrip(t *testing.T) {
	rels := &JobRelationships{
		Organization: &JobRelationshipsOrganization{ID: "o-5", Type: "organization"},
		Workspace: &JobRelationshipsWorkspace{
			Data: &JobRelationshipsWorkspaceData{ID: "w-5", Type: "workspace"},
		},
	}
	data := mustMarshal(t, rels)

	var got JobRelationships
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Organization.ID != "o-5" {
		t.Errorf("Organization.ID: got %q", got.Organization.ID)
	}
	if got.Workspace.Data.ID != "w-5" {
		t.Errorf("Workspace.Data.ID: got %q", got.Workspace.Data.ID)
	}
}

// ---------------------------------------------------------------------------
// Team
// ---------------------------------------------------------------------------

func fullTeam() *Team {
	return &Team{
		ID:   "team-7",
		Type: "team",
		Attributes: &TeamAttributes{
			Name:             "platform-team",
			ManageWorkspace:  true,
			ManageModule:     true,
			ManageProvider:   false,
			ManageState:      true,
			ManageCollection: false,
			ManageVcs:        true,
			ManageTemplate:   false,
		},
	}
}

func TestTeam_RoundTrip(t *testing.T) {
	orig := fullTeam()
	data := mustMarshal(t, orig)

	var got Team
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", got.ID, orig.ID)
	}
	if got.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", got.Type, orig.Type)
	}
	a := got.Attributes
	oa := orig.Attributes
	if a.Name != oa.Name {
		t.Errorf("Name: got %q, want %q", a.Name, oa.Name)
	}
	if a.ManageWorkspace != oa.ManageWorkspace {
		t.Errorf("ManageWorkspace: got %v, want %v", a.ManageWorkspace, oa.ManageWorkspace)
	}
	if a.ManageModule != oa.ManageModule {
		t.Errorf("ManageModule: got %v, want %v", a.ManageModule, oa.ManageModule)
	}
	if a.ManageProvider != oa.ManageProvider {
		t.Errorf("ManageProvider: got %v, want %v", a.ManageProvider, oa.ManageProvider)
	}
	if a.ManageState != oa.ManageState {
		t.Errorf("ManageState: got %v, want %v", a.ManageState, oa.ManageState)
	}
	if a.ManageCollection != oa.ManageCollection {
		t.Errorf("ManageCollection: got %v, want %v", a.ManageCollection, oa.ManageCollection)
	}
	if a.ManageVcs != oa.ManageVcs {
		t.Errorf("ManageVcs: got %v, want %v", a.ManageVcs, oa.ManageVcs)
	}
	if a.ManageTemplate != oa.ManageTemplate {
		t.Errorf("ManageTemplate: got %v, want %v", a.ManageTemplate, oa.ManageTemplate)
	}
}

func TestTeam_JSONFieldNames(t *testing.T) {
	data := mustMarshal(t, fullTeam())
	for _, key := range []string{
		"id", "type", "attributes",
		"name", "manageWorkspace", "manageModule",
		"manageState", "manageVcs",
	} {
		jsonContainsKey(t, data, key)
	}
	for _, bad := range []string{
		"manage_workspace", "manage_module", "manage_provider",
		"manage_state", "manage_collection", "manage_vcs", "manage_template",
		"ManageWorkspace", "ManageModule",
	} {
		jsonLacksKey(t, data, bad)
	}
}

func TestTeam_BooleanTrueSerialization(t *testing.T) {
	team := fullTeam()
	data := mustMarshal(t, team)
	raw := string(data)

	for _, field := range []string{"manageWorkspace", "manageModule", "manageState", "manageVcs"} {
		expected := `"` + field + `":true`
		if !strings.Contains(raw, expected) {
			t.Errorf("expected %s in JSON: %s", expected, raw)
		}
	}
}

func TestTeam_BooleanFalseOmitted(t *testing.T) {
	team := fullTeam()
	data := mustMarshal(t, team)

	// ManageProvider, ManageCollection, ManageTemplate are false so omitempty drops them
	jsonLacksKey(t, data, "manageProvider")
	jsonLacksKey(t, data, "manageCollection")
	jsonLacksKey(t, data, "manageTemplate")
}

func TestTeam_AllBooleansTrueRoundTrip(t *testing.T) {
	team := &Team{
		ID:   "team-all",
		Type: "team",
		Attributes: &TeamAttributes{
			Name:             "admin",
			ManageWorkspace:  true,
			ManageModule:     true,
			ManageProvider:   true,
			ManageState:      true,
			ManageCollection: true,
			ManageVcs:        true,
			ManageTemplate:   true,
		},
	}
	data := mustMarshal(t, team)

	var got Team
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	a := got.Attributes
	if !a.ManageWorkspace || !a.ManageModule || !a.ManageProvider ||
		!a.ManageState || !a.ManageCollection || !a.ManageVcs || !a.ManageTemplate {
		t.Errorf("expected all booleans true, got %+v", a)
	}
}

func TestTeam_OmitEmpty(t *testing.T) {
	team := &Team{ID: "team-1"}
	data := mustMarshal(t, team)

	jsonContainsKey(t, data, "id")
	jsonLacksKey(t, data, "attributes")
	jsonLacksKey(t, data, "type")
}

func TestTeamAttributes_OmitEmpty(t *testing.T) {
	attrs := &TeamAttributes{Name: "only-name"}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "name")
	jsonLacksKey(t, data, "manageWorkspace")
	jsonLacksKey(t, data, "manageModule")
	jsonLacksKey(t, data, "manageProvider")
	jsonLacksKey(t, data, "manageState")
	jsonLacksKey(t, data, "manageCollection")
	jsonLacksKey(t, data, "manageVcs")
	jsonLacksKey(t, data, "manageTemplate")
}

func TestTeam_GetBody(t *testing.T) {
	body := GetBodyTeam{
		Data: []*Team{fullTeam(), {ID: "team-2", Type: "team"}},
	}
	data := mustMarshal(t, body)

	var got GetBodyTeam
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(got.Data))
	}
	if got.Data[0].ID != "team-7" {
		t.Errorf("first team ID: got %q", got.Data[0].ID)
	}
	if got.Data[1].ID != "team-2" {
		t.Errorf("second team ID: got %q", got.Data[1].ID)
	}
}

func TestTeam_PostBody(t *testing.T) {
	body := PostBodyTeam{Data: fullTeam()}
	data := mustMarshal(t, body)

	var got PostBodyTeam
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if got.Data.ID != "team-7" {
		t.Errorf("ID: got %q", got.Data.ID)
	}
}

// ---------------------------------------------------------------------------
// Team: omitempty boolean bug tests
// ---------------------------------------------------------------------------

func TestTeam_FalseBooleanOmitted_ProducesEmptyAttributes(t *testing.T) {
	// BUG: When all boolean permission fields are false, omitempty causes them
	// all to be dropped. The resulting JSON attributes object is empty ("{}"),
	// making it impossible to distinguish "all permissions denied" from
	// "no permissions specified". The API has no way to know the caller
	// intended to set everything to false.
	attrs := TeamAttributes{
		ManageWorkspace:  false,
		ManageModule:     false,
		ManageProvider:   false,
		ManageState:      false,
		ManageCollection: false,
		ManageVcs:        false,
		ManageTemplate:   false,
	}
	data := mustMarshal(t, attrs)

	if string(data) != "{}" {
		t.Errorf("expected empty object {}, got %s", data)
	}
}

func TestVariable_SensitiveFalseDropped(t *testing.T) {
	// BUG: Cannot explicitly mark a variable as non-sensitive via omitempty.
	// Setting Sensitive=false and Hcl=false produces JSON without those fields,
	// so the API never sees an explicit false. Only "key" appears in the output.
	attrs := VariableAttributes{
		Key:       "test",
		Sensitive: false,
		Hcl:       false,
	}
	data := mustMarshal(t, attrs)

	jsonContainsKey(t, data, "key")
	jsonLacksKey(t, data, "sensitive")
	jsonLacksKey(t, data, "hcl")
}

// ---------------------------------------------------------------------------
// Cross-cutting: Unmarshal from raw JSON strings
// ---------------------------------------------------------------------------

func TestOrganization_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"id": "org-raw",
		"type": "organization",
		"attributes": {
			"name": "raw-org",
			"description": "from raw JSON",
			"executionMode": "local",
			"icon": "https://example.com/raw.png"
		}
	}`
	var org Organization
	if err := json.Unmarshal([]byte(raw), &org); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if org.ID != "org-raw" {
		t.Errorf("ID: got %q", org.ID)
	}
	if org.Attributes.Name != "raw-org" {
		t.Errorf("Name: got %q", org.Attributes.Name)
	}
	if *org.Attributes.Description != "from raw JSON" {
		t.Errorf("Description: got %q", *org.Attributes.Description)
	}
	if *org.Attributes.ExecutionMode != "local" {
		t.Errorf("ExecutionMode: got %q", *org.Attributes.ExecutionMode)
	}
}

func TestWorkspace_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"id": "ws-raw",
		"type": "workspace",
		"attributes": {
			"name": "raw-ws",
			"executionMode": "local",
			"iacType": "tofu",
			"terraformVersion": "1.7.0"
		}
	}`
	var ws Workspace
	if err := json.Unmarshal([]byte(raw), &ws); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if ws.ID != "ws-raw" {
		t.Errorf("ID: got %q", ws.ID)
	}
	if ws.Attributes.ExecutionMode != "local" {
		t.Errorf("ExecutionMode: got %q", ws.Attributes.ExecutionMode)
	}
	if ws.Attributes.IacType != "tofu" {
		t.Errorf("IacType: got %q", ws.Attributes.IacType)
	}
	if ws.Attributes.TerraformVersion != "1.7.0" {
		t.Errorf("TerraformVersion: got %q", ws.Attributes.TerraformVersion)
	}
}

func TestModule_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"id": "mod-raw",
		"type": "module",
		"attributes": {
			"name": "raw-mod",
			"tagPrefix": "v",
			"registryPath": "aws/ec2/aws",
			"versions": ["1.0.0", "2.0.0"]
		}
	}`
	var mod Module
	if err := json.Unmarshal([]byte(raw), &mod); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if mod.ID != "mod-raw" {
		t.Errorf("ID: got %q", mod.ID)
	}
	if *mod.Attributes.TagPrefix != "v" {
		t.Errorf("TagPrefix: got %q", *mod.Attributes.TagPrefix)
	}
	if mod.Attributes.RegistryPath != "aws/ec2/aws" {
		t.Errorf("RegistryPath: got %q", mod.Attributes.RegistryPath)
	}
	if len(mod.Attributes.Versions) != 2 {
		t.Fatalf("Versions length: got %d", len(mod.Attributes.Versions))
	}
	if mod.Attributes.Versions[0] != "1.0.0" || mod.Attributes.Versions[1] != "2.0.0" {
		t.Errorf("Versions: got %v", mod.Attributes.Versions)
	}
}

func TestJob_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"id": "job-raw",
		"type": "job",
		"attributes": {
			"command": "destroy",
			"status": "running",
			"output": "Destroying..."
		},
		"relationships": {
			"workspace": {
				"data": {
					"id": "ws-from-json",
					"type": "workspace"
				}
			}
		}
	}`
	var job Job
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if job.ID != "job-raw" {
		t.Errorf("ID: got %q", job.ID)
	}
	if job.Attributes.Command != "destroy" {
		t.Errorf("Command: got %q", job.Attributes.Command)
	}
	if job.Attributes.Status != "running" {
		t.Errorf("Status: got %q", job.Attributes.Status)
	}
	if job.Relationships.Workspace.Data.ID != "ws-from-json" {
		t.Errorf("Workspace.Data.ID: got %q", job.Relationships.Workspace.Data.ID)
	}
}

func TestGetBodyJob_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"data": [
			{"id": "job-1", "type": "job"},
			{"id": "job-2", "type": "job"}
		]
	}`
	var body GetBodyJob
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(body.Data))
	}
	if body.Data[0].ID != "job-1" {
		t.Errorf("first job: got %q", body.Data[0].ID)
	}
}

func TestPostBodyWorkspace_UnmarshalFromRawJSON(t *testing.T) {
	raw := `{
		"data": {
			"id": "ws-post",
			"type": "workspace",
			"attributes": {
				"name": "posted-ws",
				"iacType": "terraform"
			}
		}
	}`
	var body PostBodyWorkspace
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if body.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if body.Data.ID != "ws-post" {
		t.Errorf("ID: got %q", body.Data.ID)
	}
	if body.Data.Attributes.IacType != "terraform" {
		t.Errorf("IacType: got %q", body.Data.Attributes.IacType)
	}
}
