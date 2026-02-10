package client_test

import (
	"net/http"
	"testing"

	client "terrakube/client/client"
	"terrakube/client/models"
	"terrakube/testutil"
)

func newTestClient(ts *testutil.TestServer) *client.Client {
	return client.NewClient(nil, "test-token", ts.URL())
}

// --- List ---

func TestOrganizationList_SendsGetRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyOrganization())
	c := newTestClient(ts)

	_, err := c.Organization.List("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "GET")
	testutil.AssertPath(t, req, "/api/v1/organization")
}

func TestOrganizationList_WithEmptyFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyOrganization())
	c := newTestClient(ts)

	_, err := c.Organization.List("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization")
}

func TestOrganizationList_WithFilter(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyOrganization())
	c := newTestClient(ts)

	_, err := c.Organization.List("filter[name]=acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPathPrefix(t, req, "/api/v1/organization")
	if req.Path != "/api/v1/organization?filter[name]=acme" {
		t.Errorf("expected path with filter, got %q", req.Path)
	}
}

func TestOrganizationList_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyOrganization())
	c := newTestClient(ts)

	orgs, err := c.Organization.List("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgs) != 3 {
		t.Fatalf("expected 3 organizations, got %d", len(orgs))
	}
	if orgs[0].Attributes.Name != "acme-corp" {
		t.Errorf("expected first org name %q, got %q", "acme-corp", orgs[0].Attributes.Name)
	}
}

func TestOrganizationList_ServerError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := newTestClient(ts)

	orgs, err := c.Organization.List("")
	if err == nil {
		t.Error("expected error on 500 response with empty body, got nil")
	}
	if orgs != nil {
		t.Errorf("expected nil orgs on server error, got %v", orgs)
	}
}

// --- Create ---

func TestOrganizationCreate_SendsPostRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyOrganization())
	c := newTestClient(ts)

	org := *testutil.FixtureOrganization()
	_, err := c.Organization.Create(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "POST")
	testutil.AssertPath(t, req, "/api/v1/organization")
}

func TestOrganizationCreate_BodyContainsOrgData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyOrganization())
	c := newTestClient(ts)

	org := *testutil.FixtureOrganization()
	_, err := c.Organization.Create(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "acme-corp")
	testutil.AssertBodyContains(t, req, "organization")
}

func TestOrganizationCreate_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyOrganization())
	c := newTestClient(ts)

	org := *testutil.FixtureOrganization()
	result, err := c.Organization.Create(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Attributes.Name != "acme-corp" {
		t.Errorf("expected name %q, got %q", "acme-corp", result.Attributes.Name)
	}
}

// --- Update ---

func TestOrganizationUpdate_SendsPatchRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := newTestClient(ts)

	org := models.Organization{
		ID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Type: "organization",
		Attributes: &models.OrganizationAttributes{
			Name: "acme-updated",
		},
	}
	err := c.Organization.Update(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "PATCH")
}

func TestOrganizationUpdate_UsesIDInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := newTestClient(ts)

	org := models.Organization{
		ID:   "my-org-id",
		Type: "organization",
		Attributes: &models.OrganizationAttributes{
			Name: "acme-updated",
		},
	}
	err := c.Organization.Update(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/my-org-id")
}

func TestOrganizationUpdate_BodyContainsData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := newTestClient(ts)

	org := models.Organization{
		ID:   "my-org-id",
		Type: "organization",
		Attributes: &models.OrganizationAttributes{
			Name: "acme-updated",
		},
	}
	err := c.Organization.Update(org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, "acme-updated")
}

// --- Delete ---

func TestOrganizationDelete_SendsDeleteRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := newTestClient(ts)

	err := c.Organization.Delete("my-org-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "DELETE")
}

func TestOrganizationDelete_UsesIDInPath(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := newTestClient(ts)

	err := c.Organization.Delete("delete-this-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertPath(t, req, "/api/v1/organization/delete-this-id")
}

func TestOrganizationDelete_NoRequestBody(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := newTestClient(ts)

	err := c.Organization.Delete("my-org-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	if len(req.Body) != 0 {
		t.Errorf("expected empty body for DELETE, got %q", string(req.Body))
	}
}

func TestOrganizationDelete_ReturnsNilOnSuccess(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := newTestClient(ts)

	err := c.Organization.Delete("my-org-id")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- Server Error Tests ---

func TestOrganizationCreate_ServerErrorReturnsError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := newTestClient(ts)

	org := *testutil.FixtureOrganization()
	result, err := c.Organization.Create(org)
	if err == nil {
		t.Error("expected error on 500 response, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on server error, got %v", result)
	}
}

func TestOrganizationUpdate_ServerErrorReturnsError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("PATCH", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := newTestClient(ts)

	org := models.Organization{
		ID:   "my-org-id",
		Type: "organization",
		Attributes: &models.OrganizationAttributes{
			Name: "acme-updated",
		},
	}
	// Update passes nil to do() for v, so no JSON decode happens.
	// With a 500 + nil body, do() returns (resp, nil) — no error.
	err := c.Organization.Update(org)
	if err != nil {
		t.Errorf("unexpected error (Update passes nil target, no decode): %v", err)
	}
}

func TestOrganizationDelete_ServerErrorReturnsError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := newTestClient(ts)

	// Delete passes nil to do() for v, so no JSON decode happens.
	// With a 500 + nil body, do() returns (resp, nil) — no error.
	err := c.Organization.Delete("my-org-id")
	if err != nil {
		t.Errorf("unexpected error (Delete passes nil target, no decode): %v", err)
	}
}

// --- Empty ID Edge Cases ---

// Empty ID sends DELETE to list endpoint — no validation.
func TestOrganizationDelete_EmptyID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := newTestClient(ts)

	err := c.Organization.Delete("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// fmt.Sprintf("organization/%v", "") produces "organization/" so the path
	// becomes /api/v1/organization/ — a trailing slash hitting the list endpoint.
	testutil.AssertPath(t, req, "/api/v1/organization/")
}
