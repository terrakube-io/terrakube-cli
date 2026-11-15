package cmd

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

var newPATToken = fakeJWT(`{"jti":"pat-1"}`)

type patCall struct{ method, path, body string }

func patServer(t *testing.T) func() []patCall {
	t.Helper()
	var mu sync.Mutex
	var calls []patCall
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, patCall{r.Method, r.URL.Path, string(b)})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pat/v1"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token":"` + newPATToken + `"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/pat/v1"):
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`[{"id":"pat-1","description":"ci","days":30,"deleted":false,"source":"API","createdDate":"2026-11-13T19:36:35.628Z"}]`))
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/pat/v1/pat-1"):
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(ts.Close)
	return func() []patCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]patCall(nil), calls...)
	}
}

// AC-42: create prints the token once.
func TestCmdPatCreate(t *testing.T) {
	resetGlobalFlags()
	calls := patServer(t)
	out, err := executeCommand("pat", "create", "--days", "30", "--description", "ci")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, newPATToken) != 1 || !strings.Contains(out, `"id": "pat-1"`) {
		t.Errorf("expected the token once and its id, got %q", out)
	}
	got := calls()
	if len(got) != 1 || !strings.Contains(got[0].body, `"days":30`) || !strings.Contains(got[0].body, `"description":"ci"`) {
		t.Errorf("unexpected request %+v", got)
	}
}

// AC-42: days 0 goes to the server as is; negative days fail before any request.
func TestCmdPatCreateDays(t *testing.T) {
	resetGlobalFlags()
	calls := patServer(t)
	if _, err := executeCommand("pat", "create", "--days", "0", "--description", "forever"); err != nil {
		t.Fatal(err)
	}
	if got := calls(); len(got) != 1 || !strings.Contains(got[0].body, `"days":0`) {
		t.Errorf("days 0 not sent as is: %+v", got)
	}

	resetGlobalFlags()
	calls = patServer(t)
	_, err := executeCommand("pat", "create", "--days", "-1", "--description", "x")
	if err == nil || err.Error() != "--days must not be negative" {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := calls(); len(got) != 0 {
		t.Errorf("expected no request, got %+v", got)
	}
}

// AC-43: list and delete.
func TestCmdPatListDelete(t *testing.T) {
	resetGlobalFlags()
	calls := patServer(t)
	out, err := executeCommand("pat", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"description": "ci"`) {
		t.Errorf("list output %q", out)
	}
	out, err = executeCommand("pat", "delete", "--id", "pat-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "pat deleted") {
		t.Errorf("delete output %q", out)
	}
	if got := calls(); len(got) != 2 || got[1].method != http.MethodDelete {
		t.Errorf("calls %+v", got)
	}
}

// A 403 on create explains who may create tokens.
func TestCmdPatCreateForbiddenHint(t *testing.T) {
	resetGlobalFlags()
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()
	_, err := executeCommand("pat", "create", "--description", "x")
	want := "the server refused to create a token (403): only a user logged in through the browser (terrakube login --web) may create one"
	if err == nil || err.Error() != want {
		t.Fatalf("expected %q, got %v", want, err)
	}
}
