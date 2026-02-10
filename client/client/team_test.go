package client

import (
	"encoding/json"
	"net/http"
	"testing"

	"terrakube/client/models"
	"terrakube/testutil"
)

// --- List ---

func TestTeamList_SendsGetRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Team.List("org-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "GET")
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/team")
}

func TestTeamList_WithEmptyFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Team.List("org-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/team")
}

func TestTeamList_WithFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Team.List("org-123", "filter[name]=platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPathPrefix(t, req, "/api/v1/organization/org-123/team")
	if req.Path != "/api/v1/organization/org-123/team?filter[name]=platform" {
		t.Errorf("expected path with filter, got %q", req.Path)
	}
}

func TestTeamList_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	teams, err := c.Team.List("org-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	if teams[0].Attributes.Name != "platform-engineering" {
		t.Errorf("expected first team name %q, got %q", "platform-engineering", teams[0].Attributes.Name)
	}
}

func TestTeamList_UnmarshalsPermissions(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	teams, err := c.Team.List("org-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	team := teams[0]
	if !team.Attributes.ManageWorkspace {
		t.Error("expected ManageWorkspace=true")
	}
	if !team.Attributes.ManageModule {
		t.Error("expected ManageModule=true")
	}
	if team.Attributes.ManageProvider {
		t.Error("expected ManageProvider=false")
	}
	if !team.Attributes.ManageState {
		t.Error("expected ManageState=true")
	}
	if team.Attributes.ManageCollection {
		t.Error("expected ManageCollection=false")
	}
	if !team.Attributes.ManageVcs {
		t.Error("expected ManageVcs=true")
	}
	if team.Attributes.ManageTemplate {
		t.Error("expected ManageTemplate=false")
	}
}

func TestTeamList_ServerError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	teams, err := c.Team.List("org-123", "")
	if err == nil {
		t.Error("expected error on 500 response with empty body, got nil")
	}
	if teams != nil {
		t.Errorf("expected nil teams on server error, got %v", teams)
	}
}

// --- Create ---

func TestTeamCreate_SendsPostRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := *testutil.FixtureTeam()
	_, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "POST")
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/team")
}

func TestTeamCreate_BodyContainsTeamData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := *testutil.FixtureTeam()
	_, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "platform-engineering")
	testutil.AssertBodyContains(t, req, "team")
}

func TestTeamCreate_BodyContainsBooleanPermissions(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := *testutil.FixtureTeam()
	_, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"manageWorkspace":true`)
	testutil.AssertBodyContains(t, req, `"manageModule":true`)
	testutil.AssertBodyContains(t, req, `"manageState":true`)
	testutil.AssertBodyContains(t, req, `"manageVcs":true`)
}

func TestTeamCreate_OmitsEmptyFalseBooleans(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:            "read-only",
			ManageWorkspace: false,
			ManageModule:    false,
		},
	}
	_, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	var body map[string]interface{}
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})

	// With omitempty, false booleans are omitted from JSON
	if _, exists := attrs["manageWorkspace"]; exists {
		t.Error("expected manageWorkspace to be omitted when false (omitempty)")
	}
	if _, exists := attrs["manageModule"]; exists {
		t.Error("expected manageModule to be omitted when false (omitempty)")
	}
}

func TestTeamCreate_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := *testutil.FixtureTeam()
	result, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Attributes.Name != "platform-engineering" {
		t.Errorf("expected name %q, got %q", "platform-engineering", result.Attributes.Name)
	}
	if !result.Attributes.ManageWorkspace {
		t.Error("expected ManageWorkspace=true in response")
	}
}

// --- Delete ---

func TestTeamDelete_SendsDeleteRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Team.Delete("org-123", "team-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "DELETE")
}

func TestTeamDelete_UsesOrgAndTeamIDInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Team.Delete("org-123", "team-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/team/team-456")
}

func TestTeamDelete_NoRequestBody(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Team.Delete("org-123", "team-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	if len(req.Body) != 0 {
		t.Errorf("expected empty body for DELETE, got %q", string(req.Body))
	}
}

func TestTeamDelete_ReturnsNilOnSuccess(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Team.Delete("org-123", "team-456")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- Update ---

func TestTeamUpdate_SendsPatchRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		ID:   "team-456",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:            "updated-team",
			ManageWorkspace: true,
		},
	}
	err := c.Team.Update("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "PATCH")
}

func TestTeamUpdate_UsesOrgAndTeamIDInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		ID:   "team-456",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name: "updated-team",
		},
	}
	err := c.Team.Update("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-123/team/team-456")
}

func TestTeamUpdate_BodyContainsUpdatedData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		ID:   "team-456",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:            "updated-team",
			ManageWorkspace: true,
			ManageModule:    true,
		},
	}
	err := c.Team.Update("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "updated-team")
	testutil.AssertBodyContains(t, req, `"manageWorkspace":true`)
	testutil.AssertBodyContains(t, req, `"manageModule":true`)
}

func TestTeamUpdate_FalseBooleanDroppedFromBody(t *testing.T) {
	// BUG: omitempty on bool fields means false values are silently dropped.
	// You cannot disable a permission via Update because false is omitted from JSON.
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		ID:   "team-456",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:            "disable-perms",
			ManageWorkspace: false,
		},
	}
	err := c.Team.Update("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	var body map[string]interface{}
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})

	if _, exists := attrs["manageWorkspace"]; exists {
		t.Error("expected manageWorkspace to be omitted when false (omitempty), but it was present")
	}
}

func TestTeamCreate_AllPermissionsFalse(t *testing.T) {
	// BUG: Creating a team with all permissions false is identical to creating
	// one with unset permissions. The API receives no permission fields at all.
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyTeam())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:             "no-perms",
			ManageWorkspace:  false,
			ManageModule:     false,
			ManageProvider:   false,
			ManageState:      false,
			ManageCollection: false,
			ManageVcs:        false,
			ManageTemplate:   false,
		},
	}
	_, err := c.Team.Create("org-123", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	var body map[string]interface{}
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})

	permFields := []string{
		"manageWorkspace", "manageModule", "manageProvider",
		"manageState", "manageCollection", "manageVcs", "manageTemplate",
	}
	for _, field := range permFields {
		if _, exists := attrs[field]; exists {
			t.Errorf("expected %s to be omitted when false (omitempty), but it was present", field)
		}
	}
}

func TestTeamUpdate_ReturnsNilOnSuccess(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	team := models.Team{
		ID:   "team-456",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name: "updated-team",
		},
	}
	err := c.Team.Update("org-123", team)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
