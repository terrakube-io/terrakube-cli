package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// --- NewClient ---

func TestNewClient_ReturnsNonNil(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_SubClientsInitialized(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	if c.Organization == nil {
		t.Error("Organization sub-client is nil")
	}
	if c.Module == nil {
		t.Error("Module sub-client is nil")
	}
	if c.Workspace == nil {
		t.Error("Workspace sub-client is nil")
	}
	if c.Variable == nil {
		t.Error("Variable sub-client is nil")
	}
	if c.Job == nil {
		t.Error("Job sub-client is nil")
	}
	if c.Team == nil {
		t.Error("Team sub-client is nil")
	}
}

func TestNewClient_NilHttpClientUsesDefault(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)
	if c.HttpClient != http.DefaultClient {
		t.Error("expected http.DefaultClient when nil is passed")
	}
}

func TestNewClient_CustomHttpClient(t *testing.T) {
	custom := &http.Client{}
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(custom, "tok", baseURL)
	if c.HttpClient != custom {
		t.Error("expected custom http client to be used")
	}
}

func TestNewClient_BasePathDefault(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)
	if c.BasePath != "/api/v1/" {
		t.Errorf("expected BasePath %q, got %q", "/api/v1/", c.BasePath)
	}
}

func TestNewClient_BasePathWithExistingPath(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost/custom/path")
	c := NewClient(nil, "tok", baseURL)
	// Code appends "/" then "/api/v1/", producing double slash
	expected := "/custom/path//api/v1/"
	if c.BasePath != expected {
		t.Errorf("expected BasePath %q, got %q", expected, c.BasePath)
	}
}

func TestNewClient_BasePathWithTrailingSlash(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost/custom/path/")
	c := NewClient(nil, "tok", baseURL)
	expected := "/custom/path//api/v1/"
	if c.BasePath != expected {
		t.Errorf("expected BasePath %q, got %q", expected, c.BasePath)
	}
}

func TestNewClient_StoresTokenAndBaseURL(t *testing.T) {
	baseURL, _ := url.Parse("http://example.com")
	c := NewClient(nil, "my-secret-token", baseURL)
	if c.Token != "my-secret-token" {
		t.Errorf("expected token %q, got %q", "my-secret-token", c.Token)
	}
	if c.BaseURL.String() != "http://example.com" {
		t.Errorf("expected base URL %q, got %q", "http://example.com", c.BaseURL.String())
	}
}

// --- newRequest ---

func TestNewRequest_ConstructsCorrectURL(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequest("GET", "organization", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "http://localhost/api/v1/organization"
	if req.URL.String() != expected {
		t.Errorf("expected URL %q, got %q", expected, req.URL.String())
	}
}

func TestNewRequest_SetsAuthorizationHeader(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "my-token", baseURL)

	req, err := c.newRequest("GET", "organization", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := req.Header.Get("Authorization")
	expected := "Bearer my-token"
	if got != expected {
		t.Errorf("expected Authorization %q, got %q", expected, got)
	}
}

func TestNewRequest_SetsContentTypeWithBody(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	body := map[string]string{"key": "value"}
	req, err := c.newRequest("POST", "organization", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ct := req.Header.Get("Content-Type")
	if ct != "application/vnd.api+json" {
		t.Errorf("expected Content-Type %q, got %q", "application/vnd.api+json", ct)
	}
}

func TestNewRequest_NoContentTypeWithoutBody(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequest("GET", "organization", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		t.Errorf("expected no Content-Type, got %q", ct)
	}
}

func TestNewRequest_MarshalsBodyToJSON(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	body := map[string]string{"name": "acme"}
	req, err := c.newRequest("POST", "organization", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if decoded["name"] != "acme" {
		t.Errorf("expected body name %q, got %q", "acme", decoded["name"])
	}
}

func TestNewRequest_SetsMethod(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	methods := []string{"GET", "POST", "PATCH", "DELETE"}
	for _, method := range methods {
		req, err := c.newRequest(method, "organization", nil)
		if err != nil {
			t.Fatalf("unexpected error for method %s: %v", method, err)
		}
		if req.Method != method {
			t.Errorf("expected method %q, got %q", method, req.Method)
		}
	}
}

func TestNewRequest_WithCustomBasePath(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost/custom")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequest("GET", "organization", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Code produces double slash due to BasePath construction bug
	expected := "http://localhost/custom//api/v1/organization"
	if req.URL.String() != expected {
		t.Errorf("expected URL %q, got %q", expected, req.URL.String())
	}
}

// --- newRequestWithFilter ---

func TestNewRequestWithFilter_EmptyFilter(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequestWithFilter("GET", "organization", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "http://localhost/api/v1/organization"
	if req.URL.String() != expected {
		t.Errorf("expected URL %q, got %q", expected, req.URL.String())
	}
}

func TestNewRequestWithFilter_WithFilter(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequestWithFilter("GET", "organization", "filter[name]=acme", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "http://localhost/api/v1/organization?filter[name]=acme"
	if req.URL.String() != expected {
		t.Errorf("expected URL %q, got %q", expected, req.URL.String())
	}
}

func TestNewRequestWithFilter_SetsAuthHeader(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "my-token", baseURL)

	req, err := c.newRequestWithFilter("GET", "organization", "filter[name]=acme", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := req.Header.Get("Authorization")
	if got != "Bearer my-token" {
		t.Errorf("expected Authorization %q, got %q", "Bearer my-token", got)
	}
}

func TestNewRequestWithFilter_SetsContentTypeWithBody(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	body := map[string]string{"key": "val"}
	req, err := c.newRequestWithFilter("POST", "organization", "", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ct := req.Header.Get("Content-Type")
	if ct != "application/vnd.api+json" {
		t.Errorf("expected Content-Type %q, got %q", "application/vnd.api+json", ct)
	}
}

func TestNewRequestWithFilter_NoContentTypeWithoutBody(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	req, err := c.newRequestWithFilter("GET", "organization", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		t.Errorf("expected no Content-Type, got %q", ct)
	}
}

// --- do ---

func TestDo_DecodesJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"acme"}`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	resp, err := c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if result["name"] != "acme" {
		t.Errorf("expected name %q, got %q", "acme", result["name"])
	}
}

func TestDo_ReturnsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("POST", srv.URL+"/test", nil)
	resp, err := c.do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestDo_NilTargetSkipsDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("DELETE", srv.URL+"/test", nil)
	resp, err := c.do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestDo_PropagatesClientErrors(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost")
	c := NewClient(nil, "tok", baseURL)

	// Request to a URL that will not connect
	req, _ := http.NewRequest("GET", "http://127.0.0.1:1/nonexistent", nil)
	_, err := c.do(req, nil)
	if err == nil {
		t.Error("expected error for failed connection, got nil")
	}
}

// BUG: do() does not check HTTP status codes. A 500 with valid JSON is decoded successfully.
func TestDo_500WithValidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	resp, err := c.do(req, &result)
	// BUG: do() does not check HTTP status codes — it happily decodes 500 responses.
	if err != nil {
		t.Fatalf("expected no error (bug: status not checked), got %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}
	if result["error"] != "internal server error" {
		t.Errorf("expected decoded error message, got %v", result)
	}
}

func TestDo_401Response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	resp, err := c.do(req, &result)
	// Empty body + non-nil v -> json decode hits EOF
	if err == nil {
		t.Error("expected error decoding empty 401 body, got nil")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

// BUG: do() does not check HTTP status codes. A 404 with valid JSON error body is decoded.
func TestDo_404WithJSONError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	_, err := c.do(req, &result)
	if err != nil {
		t.Fatalf("expected no error (bug: status not checked), got %v", err)
	}
	if result["error"] != "not found" {
		t.Errorf("expected decoded error %q, got %v", "not found", result)
	}
}

func TestDo_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{garbage`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	_, err := c.do(req, &result)
	if err == nil {
		t.Error("expected json decode error for malformed JSON, got nil")
	}
}

func TestDo_HTMLResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html>error</html>`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	_, err := c.do(req, &result)
	if err == nil {
		t.Error("expected json decode error for HTML body, got nil")
	}
}

func TestDo_EmptyBody200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
	var result map[string]string
	_, err := c.do(req, &result)
	// Empty body + non-nil v -> json decode hits EOF
	if err == nil {
		t.Error("expected error decoding empty 200 body, got nil")
	}
}

func TestDo_DecodesComplexJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"abc","type":"organization","attributes":{"name":"acme"}}]}`))
	}))
	defer srv.Close()

	baseURL, _ := url.Parse(srv.URL)
	c := NewClient(srv.Client(), "tok", baseURL)

	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)

	type orgAttrs struct {
		Name string `json:"name"`
	}
	type orgData struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Attributes *orgAttrs `json:"attributes"`
	}
	type orgResp struct {
		Data []orgData `json:"data"`
	}

	var result orgResp
	_, err := c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Data))
	}
	if result.Data[0].ID != "abc" {
		t.Errorf("expected ID %q, got %q", "abc", result.Data[0].ID)
	}
	if result.Data[0].Attributes.Name != "acme" {
		t.Errorf("expected name %q, got %q", "acme", result.Data[0].Attributes.Name)
	}
}
