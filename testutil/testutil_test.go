package testutil

import (
	"net/http"
	"testing"
)

func TestTestServerRecordsRequests(t *testing.T) {
	ts := NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, FixtureOrganizationList())

	u := ts.URL()
	if u == nil {
		t.Fatal("expected non-nil URL")
	}

	if ts.RequestCount() != 0 {
		t.Fatalf("expected 0 requests, got %d", ts.RequestCount())
	}
}

func TestTestServerRouteMatching(t *testing.T) {
	ts := NewTestServer(t)
	body := FixtureOrganizationList()
	ts.On("GET", "/api/v1/organization", http.StatusOK, body)

	resp, err := ts.Server.Client().Get(ts.Server.URL + "/api/v1/organization")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}

	req := ts.Request(0)
	AssertMethod(t, req, "GET")
	AssertPathPrefix(t, req, "/api/v1/organization")
}

func TestTestServerReturns404ForUnregisteredRoute(t *testing.T) {
	ts := NewTestServer(t)

	resp, err := ts.Server.Client().Get(ts.Server.URL + "/unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestTestServerReset(t *testing.T) {
	ts := NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, nil)

	_, err := ts.Server.Client().Get(ts.Server.URL + "/api/v1/organization")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}

	ts.Reset()
	if ts.RequestCount() != 0 {
		t.Fatalf("expected 0 requests after reset, got %d", ts.RequestCount())
	}
}

func TestFixturesArePopulated(t *testing.T) {
	org := FixtureOrganization()
	if org.ID == "" {
		t.Error("organization fixture has empty ID")
	}
	if org.Name == "" {
		t.Error("organization fixture has empty name")
	}

	ws := FixtureWorkspace()
	if ws.ID == "" {
		t.Error("workspace fixture has empty ID")
	}

	mod := FixtureModule()
	if mod.ID == "" {
		t.Error("module fixture has empty ID")
	}

	v := FixtureVariable()
	if v.ID == "" {
		t.Error("variable fixture has empty ID")
	}

	job := FixtureJob()
	if job.ID == "" {
		t.Error("job fixture has empty ID")
	}

	team := FixtureTeam()
	if team.ID == "" {
		t.Error("team fixture has empty ID")
	}
}

func TestAssertions(t *testing.T) {
	req := RecordedRequest{
		Method: "POST",
		Path:   "/api/v1/organization",
		Headers: http.Header{
			"Authorization": []string{"Bearer test-token"},
			"Content-Type":  []string{"application/vnd.api+json"},
		},
		Body: []byte(`{"data":{"type":"organization"}}`),
	}

	AssertMethod(t, req, "POST")
	AssertPath(t, req, "/api/v1/organization")
	AssertPathPrefix(t, req, "/api/v1/")
	AssertHasAuthHeader(t, req, "test-token")
	AssertHasContentType(t, req)
	AssertBodyContains(t, req, "organization")
	AssertBodyJSON(t, req, "data", map[string]interface{}{"type": "organization"})
}
