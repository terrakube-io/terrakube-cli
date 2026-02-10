package client

import (
	"net/http"
	"strings"
	"testing"

	"terrakube/client/models"
	"terrakube/testutil"
)

const testModuleOrgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// --- List ---

func TestModuleList_SendsGETToCorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusOK, testutil.FixtureGetBodyModule())

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	_, err := c.Module.List(testModuleOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "GET")
	testutil.AssertPath(t, req, "/api/v1/organization/"+testModuleOrgID+"/module")
}

func TestModuleList_WithEmptyFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusOK, testutil.FixtureGetBodyModule())

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	_, err := c.Module.List(testModuleOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/"+testModuleOrgID+"/module")
}

func TestModuleList_WithFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusOK, testutil.FixtureGetBodyModule())

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	_, err := c.Module.List(testModuleOrgID, "filter[name]=vpc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/"+testModuleOrgID+"/module?filter[name]=vpc")
}

func TestModuleList_ReturnsModules(t *testing.T) {
	fixture := testutil.FixtureGetBodyModule()
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusOK, fixture)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	modules, err := c.Module.List(testModuleOrgID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(modules) != len(fixture.Data) {
		t.Fatalf("expected %d modules, got %d", len(fixture.Data), len(modules))
	}
	if modules[0].ID != fixture.Data[0].ID {
		t.Errorf("expected first module ID %q, got %q", fixture.Data[0].ID, modules[0].ID)
	}
	if modules[0].Attributes.Name != fixture.Data[0].Attributes.Name {
		t.Errorf("expected first module name %q, got %q", fixture.Data[0].Attributes.Name, modules[0].Attributes.Name)
	}
	if modules[1].ID != fixture.Data[1].ID {
		t.Errorf("expected second module ID %q, got %q", fixture.Data[1].ID, modules[1].ID)
	}
}

// --- Create ---

func TestModuleCreate_SendsPOSTToCorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusCreated, testutil.FixturePostBodyModule())

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	module := *testutil.FixtureModule()
	_, err := c.Module.Create(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "POST")
	testutil.AssertPathPrefix(t, req, "/api/v1/organization/"+testModuleOrgID+"/module")
}

func TestModuleCreate_WrapsBodyInPostBodyModule(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusCreated, testutil.FixturePostBodyModule())

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	module := *testutil.FixtureModule()
	_, err := c.Module.Create(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"data"`)
	testutil.AssertBodyContains(t, req, `"type":"module"`)
	testutil.AssertBodyContains(t, req, `"attributes"`)
}

func TestModuleCreate_ReturnsModule(t *testing.T) {
	fixture := testutil.FixturePostBodyModule()
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization/"+testModuleOrgID+"/module", http.StatusCreated, fixture)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	module := *testutil.FixtureModule()
	result, err := c.Module.Create(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil module")
	}
	if result.ID != fixture.Data.ID {
		t.Errorf("expected ID %q, got %q", fixture.Data.ID, result.ID)
	}
	if result.Attributes.Name != fixture.Data.Attributes.Name {
		t.Errorf("expected name %q, got %q", fixture.Data.Attributes.Name, result.Attributes.Name)
	}
	if result.Attributes.Provider != fixture.Data.Attributes.Provider {
		t.Errorf("expected provider %q, got %q", fixture.Data.Attributes.Provider, result.Attributes.Provider)
	}
}

// --- Delete ---

func TestModuleDelete_SendsDELETEToCorrectPath(t *testing.T) {
	moduleID := "f6a7b8c9-d0e1-2345-fabc-456789012345"
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization/"+testModuleOrgID+"/module/"+moduleID, http.StatusNoContent, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Delete(testModuleOrgID, moduleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "DELETE")
	testutil.AssertPath(t, req, "/api/v1/organization/"+testModuleOrgID+"/module/"+moduleID)
}

func TestModuleDelete_ReturnsNilOnSuccess(t *testing.T) {
	moduleID := "f6a7b8c9-d0e1-2345-fabc-456789012345"
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization/"+testModuleOrgID+"/module/"+moduleID, http.StatusNoContent, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Delete(testModuleOrgID, moduleID)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- Update ---

func TestModuleUpdate_SendsPATCHToCorrectPath(t *testing.T) {
	module := *testutil.FixtureModule()
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization/"+testModuleOrgID+"/module/"+module.ID, http.StatusOK, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Update(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "PATCH")
	testutil.AssertPath(t, req, "/api/v1/organization/"+testModuleOrgID+"/module/"+module.ID)
}

func TestModuleUpdate_UsesModuleIDInURL(t *testing.T) {
	module := models.Module{
		ID:   "custom-module-id-123",
		Type: "module",
		Attributes: &models.ModuleAttributes{
			Name: "updated-module",
		},
	}
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization/"+testModuleOrgID+"/module/"+module.ID, http.StatusOK, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Update(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/" + testModuleOrgID + "/module/custom-module-id-123"
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

func TestModuleUpdate_WrapsBodyInPostBodyModule(t *testing.T) {
	module := *testutil.FixtureModule()
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization/"+testModuleOrgID+"/module/"+module.ID, http.StatusOK, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Update(testModuleOrgID, module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	body := string(req.Body)
	if !strings.Contains(body, `"data"`) {
		t.Error("expected body to contain \"data\" key")
	}
	if !strings.Contains(body, `"type":"module"`) {
		t.Error("expected body to contain module type")
	}
}

func TestModuleUpdate_ReturnsNilOnSuccess(t *testing.T) {
	module := *testutil.FixtureModule()
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization/"+testModuleOrgID+"/module/"+module.ID, http.StatusOK, nil)

	c := NewClient(ts.Server.Client(), "test-token", ts.URL())
	err := c.Module.Update(testModuleOrgID, module)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
