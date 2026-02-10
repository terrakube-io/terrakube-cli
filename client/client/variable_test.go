package client

import (
	"net/http"
	"testing"

	"terrakube/client/models"
	"terrakube/testutil"
)

// --- List ---

func TestVariableList_SendsGetRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Variable.List("org-1", "ws-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "GET")
}

func TestVariableList_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Variable.List("org-1", "ws-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable")
}

func TestVariableList_PathContainsAllThreeSegments(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Variable.List("my-org", "my-ws", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/my-org/workspace/my-ws/variable"
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

// BUG: VariableClient.List accepts a filter parameter but uses newRequest
// instead of newRequestWithFilter, so the filter is silently ignored.
func TestVariableList_FilterIsIgnored(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Variable.List("org-1", "ws-1", "filter[key]=AWS_REGION")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// Filter is silently dropped because List uses newRequest, not newRequestWithFilter
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable")
}

func TestVariableList_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	vars, err := c.Variable.List("org-1", "ws-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vars) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(vars))
	}
	if vars[0].Attributes.Key != "AWS_REGION" {
		t.Errorf("expected first variable key %q, got %q", "AWS_REGION", vars[0].Attributes.Key)
	}
	if vars[1].Attributes.Key != "DB_PASSWORD" {
		t.Errorf("expected second variable key %q, got %q", "DB_PASSWORD", vars[1].Attributes.Key)
	}
}

func TestVariableList_ServerError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	vars, err := c.Variable.List("org-1", "ws-1", "")
	if err == nil {
		t.Error("expected error on 500 response with empty body, got nil")
	}
	if vars != nil {
		t.Errorf("expected nil vars on server error, got %v", vars)
	}
}

// --- Create ---

func TestVariableCreate_SendsPostRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "POST")
}

func TestVariableCreate_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable")
}

func TestVariableCreate_PathContainsAllThreeSegments(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("my-org", "my-ws", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/my-org/workspace/my-ws/variable"
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

func TestVariableCreate_BodyContainsVariableData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "AWS_REGION")
	testutil.AssertBodyContains(t, req, "variable")
}

func TestVariableCreate_BodyWrappedInPostBodyVariable(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// Body should have a top-level "data" key wrapping the variable
	testutil.AssertBodyContains(t, req, `"data"`)
}

func TestVariableCreate_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	result, err := c.Variable.Create("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Attributes.Key != "AWS_REGION" {
		t.Errorf("expected key %q, got %q", "AWS_REGION", result.Attributes.Key)
	}
	if result.Attributes.Value != "us-east-1" {
		t.Errorf("expected value %q, got %q", "us-east-1", result.Attributes.Value)
	}
}

// --- Delete ---

func TestVariableDelete_SendsDeleteRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Variable.Delete("org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "DELETE")
}

func TestVariableDelete_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Variable.Delete("org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable/var-1")
}

func TestVariableDelete_PathContainsAllThreeIDs(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Variable.Delete("my-org", "my-ws", "my-var")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/my-org/workspace/my-ws/variable/my-var"
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

func TestVariableDelete_NoRequestBody(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Variable.Delete("org-1", "ws-1", "var-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	if len(req.Body) != 0 {
		t.Errorf("expected empty body for DELETE, got %q", string(req.Body))
	}
}

func TestVariableDelete_ReturnsNilOnSuccess(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Variable.Delete("org-1", "ws-1", "var-1")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- Update ---

func TestVariableUpdate_SendsPatchRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := models.Variable{
		ID:   "var-1",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:   "AWS_REGION",
			Value: "us-west-2",
		},
	}
	err := c.Variable.Update("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "PATCH")
}

func TestVariableUpdate_CorrectPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := models.Variable{
		ID:   "var-1",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:   "AWS_REGION",
			Value: "us-west-2",
		},
	}
	err := c.Variable.Update("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable/var-1")
}

func TestVariableUpdate_PathContainsAllThreeIDs(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := models.Variable{
		ID:   "my-var",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:   "KEY",
			Value: "val",
		},
	}
	err := c.Variable.Update("my-org", "my-ws", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	expected := "/api/v1/organization/my-org/workspace/my-ws/variable/my-var"
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

func TestVariableUpdate_BodyContainsData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := models.Variable{
		ID:   "var-1",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:   "AWS_REGION",
			Value: "us-west-2",
		},
	}
	err := c.Variable.Update("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "us-west-2")
	testutil.AssertBodyContains(t, req, "AWS_REGION")
}

func TestVariableUpdate_UsesVariableIDInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := models.Variable{
		ID:   "specific-var-id",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:   "KEY",
			Value: "val",
		},
	}
	err := c.Variable.Update("org-1", "ws-1", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace/ws-1/variable/specific-var-id")
}

// --- Server Error Tests ---

func TestVariableCreate_ServerErrorReturnsError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	result, err := c.Variable.Create("org-1", "ws-1", v)
	if err == nil {
		t.Error("expected error on 500 response, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on server error, got %v", result)
	}
}

func TestVariableDelete_ServerErrorReturnsError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	// Delete passes nil to do() for v, so no JSON decode happens.
	// With a 500 + nil body, do() returns (resp, nil) — no error.
	err := c.Variable.Delete("org-1", "ws-1", "var-1")
	if err != nil {
		t.Errorf("unexpected error (Delete passes nil target, no decode): %v", err)
	}
}

// --- Empty ID Edge Cases ---

// Empty workspace ID produces a malformed path with double slash — no validation.
func TestVariableCreate_EmptyWorkspaceID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyVariable())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	v := *testutil.FixtureVariable()
	_, err := c.Variable.Create("org-1", "", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// fmt.Sprintf("organization/%v/workspace/%v/variable", "org-1", "") produces
	// "organization/org-1/workspace//variable" — double slash where workspace ID should be.
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/workspace//variable")
}
