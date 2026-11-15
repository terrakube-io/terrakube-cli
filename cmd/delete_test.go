package cmd

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonapi"

	"terrakube/internal/resource"
	"terrakube/testutil"
)

type recordedCall struct{ method, path, body string }

// deleteServer records every request and answers GETs with get.
func deleteServer(t *testing.T, get interface{}) func() []recordedCall {
	t.Helper()
	var mu sync.Mutex
	var calls []recordedCall
	ts := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		calls = append(calls, recordedCall{r.Method, r.URL.Path, string(b)})
		mu.Unlock()
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_ = jsonapi.MarshalPayload(w, get)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(ts.Close)
	return func() []recordedCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]recordedCall(nil), calls...)
	}
}

// fakeTerminal makes the confirmation prompt see a TTY (or not) with input typed on it.
func fakeTerminal(t *testing.T, tty bool, input string) {
	t.Helper()
	oldTTY, oldIn := resource.IsTerminal, resource.Stdin
	resource.IsTerminal = func() bool { return tty }
	resource.Stdin = strings.NewReader(input)
	t.Cleanup(func() { resource.IsTerminal, resource.Stdin = oldTTY, oldIn })
}

const delOrgID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

// writes drops the GETs: the SDK re-reads a resource after a 204 PATCH.
func writes(calls []recordedCall) []recordedCall {
	var out []recordedCall
	for _, c := range calls {
		if c.method != http.MethodGet {
			out = append(out, c)
		}
	}
	return out
}

// AC-9: delete marks the workspace deleted instead of sending DELETE.
func TestCmdWorkspaceDeleteIsSoft(t *testing.T) {
	resetGlobalFlags()
	calls := deleteServer(t, testutil.FixtureWorkspace())

	out, err := executeCommand("workspace", "delete", "--organization-id", delOrgID, "--id", "d4e5f6a7-b8c9-0123-defa-234567890123")
	if err != nil {
		t.Fatal(err)
	}
	got := writes(calls())
	if len(got) != 1 || got[0].method != http.MethodPatch || !strings.HasSuffix(got[0].path, "/workspace/d4e5f6a7-b8c9-0123-defa-234567890123") {
		t.Fatalf("expected one PATCH of the workspace, got %+v", got)
	}
	if !strings.Contains(got[0].body, `"attributes":{"deleted":true}`) {
		t.Errorf("expected deleted=true only, got %s", got[0].body)
	}
	if !strings.Contains(out, "workspace marked as deleted") {
		t.Errorf("unexpected output %q", out)
	}
}

// AC-10: --hard sends DELETE.
func TestCmdWorkspaceDeleteHard(t *testing.T) {
	resetGlobalFlags()
	calls := deleteServer(t, nil)

	out, err := executeCommand("workspace", "delete", "--organization-id", delOrgID, "--id", "d4e5f6a7-b8c9-0123-defa-234567890123", "--hard")
	if err != nil {
		t.Fatal(err)
	}
	if got := calls(); len(got) != 1 || got[0].method != http.MethodDelete {
		t.Fatalf("expected one DELETE, got %+v", got)
	}
	if !strings.Contains(out, "workspace deleted") {
		t.Errorf("unexpected output %q", out)
	}
}

// AC-11: --yes disables the organization without a prompt.
func TestCmdOrganizationDeleteWithYes(t *testing.T) {
	resetGlobalFlags()
	fakeTerminal(t, false, "")
	calls := deleteServer(t, testutil.FixtureOrganization())

	out, err := executeCommand("organization", "delete", "--id", delOrgID, "--yes")
	if err != nil {
		t.Fatal(err)
	}
	got := writes(calls())
	if len(got) != 1 || got[0].method != http.MethodPatch || !strings.Contains(got[0].body, `"attributes":{"disabled":true}`) {
		t.Fatalf("expected one PATCH with disabled=true, got %+v", got)
	}
	if !strings.Contains(out, "organization disabled") {
		t.Errorf("unexpected output %q", out)
	}
}

// AC-11: typing the organization name at the prompt confirms.
func TestCmdOrganizationDeleteConfirmedByName(t *testing.T) {
	resetGlobalFlags()
	fakeTerminal(t, true, "acme-corp\n")
	calls := deleteServer(t, testutil.FixtureOrganization())

	if _, err := executeCommand("organization", "delete", "--id", delOrgID); err != nil {
		t.Fatal(err)
	}
	got := calls()
	if len(got) < 2 || got[0].method != http.MethodGet || len(writes(got)) != 1 || writes(got)[0].method != http.MethodPatch {
		t.Fatalf("expected a GET for the name, then one PATCH, got %+v", got)
	}
}

// AC-12: no TTY and no --yes: fail before any request.
func TestCmdOrganizationDeleteWithoutTTYOrYes(t *testing.T) {
	resetGlobalFlags()
	fakeTerminal(t, false, "")
	calls := deleteServer(t, testutil.FixtureOrganization())

	_, err := executeCommand("organization", "delete", "--id", delOrgID)
	if err == nil || err.Error() != "organization delete needs confirmation: run it in a terminal or pass --yes" {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := calls(); len(got) != 0 {
		t.Errorf("expected no request, got %+v", got)
	}
}

// AC-12: a wrong name at the prompt sends no write.
func TestCmdOrganizationDeleteWrongName(t *testing.T) {
	resetGlobalFlags()
	fakeTerminal(t, true, "acme\n")
	calls := deleteServer(t, testutil.FixtureOrganization())

	_, err := executeCommand("organization", "delete", "--id", delOrgID)
	if err == nil || err.Error() != `confirmation did not match "acme-corp": nothing changed` {
		t.Fatalf("unexpected error: %v", err)
	}
	if w := writes(calls()); len(w) != 0 {
		t.Errorf("expected no write, got %+v", w)
	}
}
