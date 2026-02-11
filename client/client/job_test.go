package client

import (
	"net/http"
	"testing"

	"terrakube/client/models"
	"terrakube/testutil"
)

// --- List ---

func TestJobClient_List_SendsGetRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Job.List("org-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "GET")
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/job")
}

func TestJobClient_List_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	jobs, err := c.Job.List("org-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Attributes.Command != "plan" {
		t.Errorf("expected first job command %q, got %q", "plan", jobs[0].Attributes.Command)
	}
	if jobs[1].Attributes.Command != "apply" {
		t.Errorf("expected second job command %q, got %q", "apply", jobs[1].Attributes.Command)
	}
}

// BUG: JobClient.List uses newRequest instead of newRequestWithFilter.
// The filter parameter is accepted but silently ignored — any filter string
// passed to List will have no effect on the actual HTTP request.
// Fix: Change to use newRequestWithFilter like OrganizationClient.List does.
func TestJobClient_List_FilterIsIgnored(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusOK, testutil.FixtureGetBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	_, err := c.Job.List("org-1", "filter[status]=running")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	// The filter is silently ignored because List uses newRequest, not newRequestWithFilter.
	// The path should NOT contain the filter query string.
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/job")
}

func TestJobClient_List_ServerError(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("GET", "/api/v1/organization", http.StatusInternalServerError, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	jobs, err := c.Job.List("org-1", "")
	if err == nil {
		t.Error("expected error on 500 response with empty body, got nil")
	}
	if jobs != nil {
		t.Errorf("expected nil jobs on server error, got %v", jobs)
	}
}

// --- Create ---

func TestJobClient_Create_SendsPostRequest(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	job := *testutil.FixtureJob()
	_, err := c.Job.Create("org-1", job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "POST")
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/job")
}

func TestJobClient_Create_BodyContainsJobData(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	job := *testutil.FixtureJob()
	_, err := c.Job.Create("org-1", job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	testutil.AssertBodyContains(t, req, `"command":"plan"`)
	testutil.AssertBodyContains(t, req, `"type":"job"`)
}

func TestJobClient_Create_UnmarshalsResponse(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("POST", "/api/v1/organization", http.StatusCreated, testutil.FixturePostBodyJob())
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	job := *testutil.FixtureJob()
	result, err := c.Job.Create("org-1", job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Attributes.Command != "plan" {
		t.Errorf("expected command %q, got %q", "plan", result.Attributes.Command)
	}
	if result.Attributes.Status != "completed" {
		t.Errorf("expected status %q, got %q", "completed", result.Attributes.Status)
	}
}

// --- Delete ---

// BUG: JobClient.Delete sends DELETE to /module/ path instead of /job/
// This is a copy-paste error from ModuleClient. The parameter is also
// named "moduleId" instead of "jobId".
// Fix: Change path to "organization/%v/job/%v" and rename parameter.
func TestJobClient_Delete_BugWrongEndpoint(t *testing.T) {
	ts := testutil.NewTestServer(t)
	// Register the WRONG path that the buggy code actually hits
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Job.Delete("org-1", "job-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "DELETE")
	// The bug: path uses /module/ instead of /job/
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/module/job-42")
}

func TestJobClient_Delete_NoRequestBody(t *testing.T) {
	ts := testutil.NewTestServer(t)
	ts.On("DELETE", "/api/v1/organization", http.StatusNoContent, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	err := c.Job.Delete("org-1", "job-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := ts.Request(0)
	if len(req.Body) != 0 {
		t.Errorf("expected empty body for DELETE, got %q", string(req.Body))
	}
}

// --- Update ---

// BUG: JobClient.Update accepts models.Module instead of models.Job
// and wraps it in PostBodyModule instead of PostBodyJob.
// It sends PATCH to /module/ path instead of /job/.
// This is a copy-paste error from ModuleClient.
// Fix: Change signature to accept models.Job, use PostBodyJob, fix path.
func TestJobClient_Update_BugWrongModel(t *testing.T) {
	ts := testutil.NewTestServer(t)
	// Register the WRONG path that the buggy code actually hits
	ts.On("PATCH", "/api/v1/organization", http.StatusOK, nil)
	c := NewClient(ts.Server.Client(), "test-token", ts.URL())

	// Must pass a Module because the buggy signature requires it
	mod := models.Module{
		ID:   "job-99",
		Type: "module",
		Attributes: &models.ModuleAttributes{
			Name: "not-a-job",
		},
	}
	err := c.Job.Update("org-1", mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", ts.RequestCount())
	}
	req := ts.Request(0)
	testutil.AssertMethod(t, req, "PATCH")
	// The bug: path uses /module/ instead of /job/
	testutil.AssertPath(t, req, "/api/v1/organization/org-1/module/job-99")
	// The body is wrapped in PostBodyModule instead of PostBodyJob
	testutil.AssertBodyContains(t, req, `"type":"module"`)
	testutil.AssertBodyContains(t, req, `"name":"not-a-job"`)
}

