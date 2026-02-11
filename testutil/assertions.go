package testutil

import (
	"encoding/json"
	"strings"
	"testing"
)

// AssertMethod checks that the recorded request used the expected HTTP method.
func AssertMethod(t *testing.T, req RecordedRequest, expected string) {
	t.Helper()
	if req.Method != expected {
		t.Errorf("expected method %q, got %q", expected, req.Method)
	}
}

// AssertPath checks that the recorded request path matches exactly.
func AssertPath(t *testing.T, req RecordedRequest, expected string) {
	t.Helper()
	if req.Path != expected {
		t.Errorf("expected path %q, got %q", expected, req.Path)
	}
}

// AssertPathPrefix checks that the recorded request path starts with the expected prefix.
func AssertPathPrefix(t *testing.T, req RecordedRequest, expected string) {
	t.Helper()
	if !strings.HasPrefix(req.Path, expected) {
		t.Errorf("expected path to start with %q, got %q", expected, req.Path)
	}
}

// AssertHasAuthHeader checks that the request has an Authorization header with the given token.
func AssertHasAuthHeader(t *testing.T, req RecordedRequest, token string) {
	t.Helper()
	expected := "Bearer " + token
	got := req.Headers.Get("Authorization")
	if got != expected {
		t.Errorf("expected Authorization header %q, got %q", expected, got)
	}
}

// AssertHasContentType checks that the request has the JSON:API content type.
func AssertHasContentType(t *testing.T, req RecordedRequest) {
	t.Helper()
	ct := req.Headers.Get("Content-Type")
	if ct != "application/vnd.api+json" {
		t.Errorf("expected Content-Type %q, got %q", "application/vnd.api+json", ct)
	}
}

// AssertNoContentType checks that the request does not have a Content-Type header.
func AssertNoContentType(t *testing.T, req RecordedRequest) {
	t.Helper()
	ct := req.Headers.Get("Content-Type")
	if ct != "" {
		t.Errorf("expected no Content-Type header, got %q", ct)
	}
}

// AssertBodyContains checks that the request body contains the given substring.
func AssertBodyContains(t *testing.T, req RecordedRequest, substring string) {
	t.Helper()
	body := string(req.Body)
	if !strings.Contains(body, substring) {
		t.Errorf("expected body to contain %q, got %q", substring, body)
	}
}

// AssertBodyJSON checks that a top-level key in the JSON request body has the expected value.
func AssertBodyJSON(t *testing.T, req RecordedRequest, key string, expected interface{}) {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(req.Body, &m); err != nil {
		t.Fatalf("failed to unmarshal request body as JSON: %v", err)
	}
	got, ok := m[key]
	if !ok {
		t.Errorf("expected key %q in JSON body, not found", key)
		return
	}
	expectedJSON, _ := json.Marshal(expected)
	gotJSON, _ := json.Marshal(got)
	if string(expectedJSON) != string(gotJSON) {
		t.Errorf("expected body[%q] = %s, got %s", key, expectedJSON, gotJSON)
	}
}
