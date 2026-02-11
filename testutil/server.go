package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
)

// RecordedRequest holds captured details of an HTTP request made to the test server.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

type route struct {
	method       string
	pathPrefix   string
	statusCode   int
	responseBody interface{}
}

// TestServer wraps httptest.Server to record requests and return canned responses.
type TestServer struct {
	Server   *httptest.Server
	Requests []RecordedRequest

	mu     sync.Mutex
	routes []route
	t      *testing.T
}

// NewTestServer creates a TestServer that records requests and dispatches to registered routes.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()
	ts := &TestServer{t: t}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ts.mu.Lock()
		ts.Requests = append(ts.Requests, RecordedRequest{
			Method:  r.Method,
			Path:    r.URL.RequestURI(),
			Headers: r.Header.Clone(),
			Body:    body,
		})

		var matched *route
		for i := range ts.routes {
			rt := &ts.routes[i]
			if rt.method == r.Method && strings.HasPrefix(r.URL.Path, rt.pathPrefix) {
				matched = rt
				break
			}
		}
		ts.mu.Unlock()

		if matched == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(matched.statusCode)

		if matched.responseBody != nil {
			if err := jsonapi.MarshalPayload(w, matched.responseBody); err != nil {
				t.Errorf("failed to encode response body: %v", err)
			}
		}
	}))

	t.Cleanup(func() { ts.Server.Close() })
	return ts
}

// On registers a canned response for the given method and path prefix.
// responseBody will be JSON:API-marshaled in the response.
func (ts *TestServer) On(method, path string, statusCode int, responseBody interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.routes = append(ts.routes, route{
		method:       method,
		pathPrefix:   path,
		statusCode:   statusCode,
		responseBody: responseBody,
	})
}

// URL returns the parsed base URL of the test server.
func (ts *TestServer) URL() *url.URL {
	ts.t.Helper()
	u, err := url.Parse(ts.Server.URL)
	if err != nil {
		ts.t.Fatalf("failed to parse test server URL: %v", err)
	}
	return u
}

// Close shuts down the test server.
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// Request returns the recorded request at the given index.
func (ts *TestServer) Request(index int) RecordedRequest {
	ts.t.Helper()
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if index < 0 || index >= len(ts.Requests) {
		ts.t.Fatalf("request index %d out of range (have %d requests)", index, len(ts.Requests))
	}
	return ts.Requests[index]
}

// RequestCount returns the number of recorded requests.
func (ts *TestServer) RequestCount() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return len(ts.Requests)
}

// Reset clears all recorded requests.
func (ts *TestServer) Reset() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Requests = nil
}
