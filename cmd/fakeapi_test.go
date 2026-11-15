package cmd

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"
)

// apiCall is one request seen by fakeAPI.
type apiCall struct {
	method, path string
	body         []byte
}

// fakeAPI answers a GET whose path ends in "/"+key of lists with that JSON:API
// payload, and every write with resp (201 for POST). It records all requests;
// any other GET fails the test.
func fakeAPI(t *testing.T, lists map[string]any, resp any) func() []apiCall {
	t.Helper()
	var mu sync.Mutex
	var calls []apiCall
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, apiCall{r.Method, r.URL.Path, b})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Method == http.MethodGet {
			for key, items := range lists {
				if strings.HasSuffix(r.URL.Path, "/"+key) {
					_ = jsonapi.MarshalPayload(w, items)
					return
				}
			}
			t.Errorf("unexpected GET %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_ = jsonapi.MarshalPayload(w, resp)
	}))
	t.Cleanup(ts.Close)
	return func() []apiCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]apiCall(nil), calls...)
	}
}

// onlyWrite returns the data object of the single non-GET request in calls.
func onlyWrite(t *testing.T, calls []apiCall) (path string, attrs, rels map[string]any) {
	t.Helper()
	var writes []apiCall
	for _, c := range calls {
		if c.method != http.MethodGet {
			writes = append(writes, c)
		}
	}
	if len(writes) != 1 {
		t.Fatalf("expected one write, got %d: %v", len(writes), calls)
	}
	data := writeData(t, writes[0].body)
	attrs, _ = data["attributes"].(map[string]any)
	rels, _ = data["relationships"].(map[string]any)
	return writes[0].path, attrs, rels
}

// relID returns the ID a to-one relationship in rels links to, or "" if unset.
func relID(rels map[string]any, name string) string {
	r, _ := rels[name].(map[string]any)
	d, _ := r["data"].(map[string]any)
	id, _ := d["id"].(string)
	return id
}
