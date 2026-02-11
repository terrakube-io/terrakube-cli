package client

import (
	"net/http"
	"testing"

	"terrakube/client/models"
	"terrakube/testutil"
)

const wsTestOrgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// --- List ---

func TestWorkspaceList_SendsGetRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Workspace.List(wsTestOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertMethod(t, req, http.MethodGet)
}

func TestWorkspaceList_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Workspace.List(wsTestOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace"
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceList_InterpolatesOrgID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	customOrgID := "custom-org-id-999"
	_, err := c.Workspace.List(customOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + customOrgID + "/workspace"
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceList_EmptyFilterCleanURL(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Workspace.List(wsTestOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace"
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceList_NonEmptyFilterAppendsQuery(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	filter := "filter[workspace]=name%3Dproduction-vpc"
	_, err := c.Workspace.List(wsTestOrgID, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace?" + filter
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceList_ReturnsWorkspaces(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodGet, "/api/v1/organization/", http.StatusOK, testutil.FixtureGetBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	workspaces, err := c.Workspace.List(wsTestOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
	if workspaces[0].ID != "d4e5f6a7-b8c9-0123-defa-234567890123" {
		t.Errorf("expected first workspace ID %q, got %q", "d4e5f6a7-b8c9-0123-defa-234567890123", workspaces[0].ID)
	}
	if workspaces[0].Attributes.Name != "production-vpc" {
		t.Errorf("expected first workspace name %q, got %q", "production-vpc", workspaces[0].Attributes.Name)
	}
	if workspaces[1].ID != "e5f6a7b8-c9d0-1234-efab-345678901234" {
		t.Errorf("expected second workspace ID %q, got %q", "e5f6a7b8-c9d0-1234-efab-345678901234", workspaces[1].ID)
	}
}

// --- Create ---

func TestWorkspaceCreate_SendsPostRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPost, "/api/v1/organization/", http.StatusCreated, testutil.FixturePostBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	_, err := c.Workspace.Create(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertMethod(t, req, http.MethodPost)
}

func TestWorkspaceCreate_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPost, "/api/v1/organization/", http.StatusCreated, testutil.FixturePostBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	_, err := c.Workspace.Create(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace"
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceCreate_WrapsInPostBody(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPost, "/api/v1/organization/", http.StatusCreated, testutil.FixturePostBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	_, err := c.Workspace.Create(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"data"`)
}

func TestWorkspaceCreate_BodyContainsAllAttributes(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPost, "/api/v1/organization/", http.StatusCreated, testutil.FixturePostBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	_, err := c.Workspace.Create(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"name":"production-vpc"`)
	testutil.AssertBodyContains(t, req, `"description":"Production VPC infrastructure"`)
	testutil.AssertBodyContains(t, req, `"source":"https://github.com/acme-corp/infra.git"`)
	testutil.AssertBodyContains(t, req, `"folder":"/"`)
	testutil.AssertBodyContains(t, req, `"executionMode":"remote"`)
	testutil.AssertBodyContains(t, req, `"branch":"main"`)
	testutil.AssertBodyContains(t, req, `"iacType":"terraform"`)
	testutil.AssertBodyContains(t, req, `"terraformVersion":"1.5.7"`)
}

func TestWorkspaceCreate_ReturnsWorkspace(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPost, "/api/v1/organization/", http.StatusCreated, testutil.FixturePostBodyWorkspace())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	result, err := c.Workspace.Create(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil workspace")
	}
	if result.ID != "d4e5f6a7-b8c9-0123-defa-234567890123" {
		t.Errorf("expected ID %q, got %q", "d4e5f6a7-b8c9-0123-defa-234567890123", result.ID)
	}
	if result.Attributes.Name != "production-vpc" {
		t.Errorf("expected name %q, got %q", "production-vpc", result.Attributes.Name)
	}
	if result.Attributes.TerraformVersion != "1.5.7" {
		t.Errorf("expected terraform version %q, got %q", "1.5.7", result.Attributes.TerraformVersion)
	}
}

// --- Delete ---

func TestWorkspaceDelete_SendsDeleteRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization/", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	wsID := "d4e5f6a7-b8c9-0123-defa-234567890123"
	err := c.Workspace.Delete(wsTestOrgID, wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertMethod(t, req, http.MethodDelete)
}

func TestWorkspaceDelete_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization/", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	wsID := "d4e5f6a7-b8c9-0123-defa-234567890123"
	err := c.Workspace.Delete(wsTestOrgID, wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace/" + wsID
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceDelete_BothIDsInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization/", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	customOrg := "org-abc"
	customWs := "ws-xyz"
	err := c.Workspace.Delete(customOrg, customWs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + customOrg + "/workspace/" + customWs
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceDelete_ReturnsNoError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization/", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	wsID := "d4e5f6a7-b8c9-0123-defa-234567890123"
	err := c.Workspace.Delete(wsTestOrgID, wsID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// --- Update ---

func TestWorkspaceUpdate_SendsPatchRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPatch, "/api/v1/organization/", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	err := c.Workspace.Update(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertMethod(t, req, http.MethodPatch)
}

func TestWorkspaceUpdate_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPatch, "/api/v1/organization/", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	err := c.Workspace.Update(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace/" + ws.ID
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceUpdate_UsesWorkspaceIDForURL(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPatch, "/api/v1/organization/", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := models.Workspace{
		ID:   "custom-ws-id",
		Type: "workspace",
		Attributes: &models.WorkspaceAttributes{
			Name: "updated-workspace",
		},
	}
	err := c.Workspace.Update(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + wsTestOrgID + "/workspace/custom-ws-id"
	testutil.AssertPath(t, req, expected)
}

func TestWorkspaceUpdate_BodyContainsUpdatedAttributes(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPatch, "/api/v1/organization/", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := models.Workspace{
		ID:   "ws-123",
		Type: "workspace",
		Attributes: &models.WorkspaceAttributes{
			Name:             "renamed-workspace",
			Description:      "Updated description",
			Branch:           "develop",
			TerraformVersion: "1.6.0",
		},
	}
	err := c.Workspace.Update(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"data"`)
	testutil.AssertBodyContains(t, req, `"name":"renamed-workspace"`)
	testutil.AssertBodyContains(t, req, `"description":"Updated description"`)
	testutil.AssertBodyContains(t, req, `"branch":"develop"`)
	testutil.AssertBodyContains(t, req, `"terraformVersion":"1.6.0"`)
}

func TestWorkspaceUpdate_ReturnsNoError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodPatch, "/api/v1/organization/", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	ws := *testutil.FixtureWorkspace()
	err := c.Workspace.Update(wsTestOrgID, ws)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// --- Empty ID Edge Cases ---

// Empty org ID produces /api/v1/organization//workspace/ws-123 — double slash, no validation.
func TestWorkspaceDelete_EmptyOrgID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Workspace.Delete("", "ws-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// fmt.Sprintf("organization/%v/workspace/%v", "", "ws-123") produces
	// "organization//workspace/ws-123"
	testutil.AssertPath(t, req, "/api/v1/organization//workspace/ws-123")
}

// Empty workspace ID produces /api/v1/organization/org-123/workspace/ — trailing slash, no validation.
func TestWorkspaceDelete_EmptyWsID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On(http.MethodDelete, "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Workspace.Delete("org-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/workspace/")
}
