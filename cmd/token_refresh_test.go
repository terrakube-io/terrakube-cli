package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/jsonapi"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/exit"
)

// refreshFixture is an API server that records the bearer token and a
// Dex-like token endpoint for the refresh_token grant.
type refreshFixture struct {
	api, dex *httptest.Server
	cfg      string
	handler  http.HandlerFunc // token endpoint response; default rotates both tokens

	mu       sync.Mutex
	bearer   string
	apiCalls int
	form     url.Values
	calls    int
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()
	f := &refreshFixture{}
	f.api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.bearer = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		f.apiCalls++
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{})
	}))
	t.Cleanup(f.api.Close)
	f.dex = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.form = r.PostForm
		f.calls++
		f.mu.Unlock()
		if f.handler != nil {
			f.handler(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"r2","id_token":"x","token_type":"bearer","expires_in":3600}`)
	}))
	t.Cleanup(f.dex.Close)
	return f
}

// writeConfig writes a Dex session to the config: token "old", refresh
// token "r1", expiring at expiresAt, renewed at the fixture's token endpoint.
func (f *refreshFixture) writeConfig(t *testing.T, expiresAt time.Time) {
	t.Helper()
	f.writeConfigWithEndpoint(t, expiresAt, f.dex.URL+"/token")
}

func (f *refreshFixture) writeConfigWithEndpoint(t *testing.T, expiresAt time.Time, endpoint string) {
	t.Helper()
	f.cfg = filepath.Join(t.TempDir(), "cfg.yaml")
	body := fmt.Sprintf("api_url: %s\ntoken: old\nrefresh_token: r1\ntoken_expires_at: %d\ntoken_endpoint: %q\nclient_id: dex-client\ncustom: keep\n",
		f.api.URL, expiresAt.Unix(), endpoint)
	if err := os.WriteFile(f.cfg, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset() // read the config file, not the overrides from resetGlobalFlags
}

func (f *refreshFixture) state() (bearer string, apiCalls int, form url.Values, calls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.bearer, f.apiCalls, f.form, f.calls
}

func TestTokenRefreshBeforeRequest(t *testing.T) {
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.writeConfig(t, time.Now().Add(-10*time.Second))

	if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bearer, _, form, calls := f.state()
	if bearer != "new-access" {
		t.Errorf("API saw token %q, want the refreshed one", bearer)
	}
	if calls != 1 || form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "r1" || form.Get("client_id") != "dex-client" {
		t.Errorf("refresh calls=%d form=%v", calls, form)
	}
	got := readConfig(t, f.cfg)
	for k, v := range map[string]any{
		"token": "new-access", "refresh_token": "r2", "custom": "keep",
		"token_endpoint": f.dex.URL + "/token", "client_id": "dex-client", "api_url": f.api.URL,
	} {
		if got[k] != v {
			t.Errorf("config[%q] = %v, want %v", k, got[k], v)
		}
	}
	exp, _ := got["token_expires_at"].(int)
	if d := time.Until(time.Unix(int64(exp), 0)); d < 59*time.Minute || d > time.Hour {
		t.Errorf("token_expires_at %v is not about an hour from now", got["token_expires_at"])
	}
	assertMode0600(t, f.cfg)

	// The refreshed session is used from now on, without a second refresh.
	if _, err := newClient(); err != nil {
		t.Fatal(err)
	}
	if currentToken() != "new-access" {
		t.Errorf("currentToken = %q after refresh", currentToken())
	}
	if _, _, _, calls := f.state(); calls != 1 {
		t.Errorf("refresh calls = %d, want 1", calls)
	}
}

func TestTokenRefreshWithinAMinuteOfExpiry(t *testing.T) {
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.writeConfig(t, time.Now().Add(30*time.Second))

	if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bearer, _, _, calls := f.state(); calls != 1 || bearer != "new-access" {
		t.Errorf("calls=%d bearer=%q, want a refresh", calls, bearer)
	}
}

func TestTokenRefreshNotDue(t *testing.T) {
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.writeConfig(t, time.Now().Add(time.Hour))

	if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bearer, _, _, calls := f.state(); calls != 0 || bearer != "old" {
		t.Errorf("calls=%d bearer=%q, want no refresh", calls, bearer)
	}
}

func TestTokenRefreshKeepsUnrotatedRefreshToken(t *testing.T) {
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.handler = func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"access_token":"new-access"}`)
	}
	f.writeConfig(t, time.Now().Add(-time.Second))

	if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readConfig(t, f.cfg)
	if got["token"] != "new-access" || got["refresh_token"] != "r1" {
		t.Errorf("config = %v, want the new token and the old refresh token", got)
	}
	if _, ok := got["token_expires_at"]; ok {
		t.Errorf("token_expires_at = %v, want it removed when the response has no expires_in", got["token_expires_at"])
	}
}

func TestTokenRefreshFailure(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		noEndpoint bool
	}{
		{
			name: "invalid_grant",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, `{"error":"invalid_grant"}`)
			},
		},
		{
			name:    "no access token",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, `{"refresh_token":"r2"}`) },
		},
		{
			name:    "connection dropped",
			handler: func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) },
		},
		{name: "no token endpoint saved", noEndpoint: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			f := newRefreshFixture(t)
			f.handler = tt.handler
			endpoint := f.dex.URL + "/token"
			if tt.noEndpoint {
				endpoint = ""
			}
			f.writeConfigWithEndpoint(t, time.Now().Add(-time.Minute), endpoint)
			before, _ := os.ReadFile(f.cfg)

			_, err := executeCommand("organization", "list", "--config", f.cfg, "--verbose")
			if err == nil || err.Error() != "session expired, run terrakube login --web" {
				t.Fatalf("error = %v, want the session expired message", err)
			}
			if code := exit.Code(err); code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if _, apiCalls, _, _ := f.state(); apiCalls != 0 {
				t.Errorf("API called %d times, want 0", apiCalls)
			}
			if after, _ := os.ReadFile(f.cfg); string(after) != string(before) {
				t.Errorf("config changed:\n%s", after)
			}
		})
	}
}

func TestTokenRefreshNotForEnvTokens(t *testing.T) {
	for _, env := range []string{"TERRAKUBE_TOKEN", "TERRAKUBE_PAT"} {
		t.Run(env, func(t *testing.T) {
			resetGlobalFlags()
			f := newRefreshFixture(t)
			f.writeConfig(t, time.Now().Add(-time.Hour))
			before, _ := os.ReadFile(f.cfg)
			t.Setenv(env, "env-token")

			if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if bearer, _, _, calls := f.state(); calls != 0 || bearer != "env-token" {
				t.Errorf("calls=%d bearer=%q, want the env token without a refresh", calls, bearer)
			}
			if after, _ := os.ReadFile(f.cfg); string(after) != string(before) {
				t.Errorf("config changed:\n%s", after)
			}
		})
	}
}

func TestTokenRefreshPATConfigUntouched(t *testing.T) {
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.cfg = filepath.Join(t.TempDir(), "cfg.yaml")
	body := "api_url: " + f.api.URL + "\ntoken: my-pat\n"
	if err := os.WriteFile(f.cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	if _, err := executeCommand("organization", "list", "--config", f.cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bearer, _, _, calls := f.state(); calls != 0 || bearer != "my-pat" {
		t.Errorf("calls=%d bearer=%q, want the PAT without a refresh", calls, bearer)
	}
	info, _ := os.Stat(f.cfg)
	if after, _ := os.ReadFile(f.cfg); string(after) != body || info.Mode().Perm() != 0o644 {
		t.Errorf("config rewritten (mode %o):\n%s", info.Mode().Perm(), after)
	}
}

func TestTokenRefreshSaveFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write to a read-only directory")
	}
	resetGlobalFlags()
	f := newRefreshFixture(t)
	f.writeConfig(t, time.Now().Add(-time.Minute))
	dir := filepath.Dir(f.cfg)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	_, err := executeCommand("organization", "list", "--config", f.cfg)
	if err == nil || !strings.Contains(err.Error(), "saving refreshed token") {
		t.Fatalf("expected a save error, got %v", err)
	}
}

// staleSession is refresh data a previous login --web with Dex left behind.
const staleSession = "refresh_token: stale\ntoken_expires_at: 1\ntoken_endpoint: http://dex.invalid/token\nclient_id: dex-client\n"

func assertNoRefreshData(t *testing.T, cfg string) {
	t.Helper()
	got := readConfig(t, cfg)
	for _, k := range []string{"refresh_token", "token_expires_at", "token_endpoint", "client_id"} {
		if _, ok := got[k]; ok {
			t.Errorf("config still has %s: %v", k, got)
		}
	}
}

func TestCmdLoginWebDexStoresRefreshData(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, true)
	port := freePort(t)
	idp.setPorts(port, port)
	stubBrowser(t, nil)
	cfg := writeTempFile(t, "cfg.yaml", "custom: keep\n")

	if out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	got := readConfig(t, cfg)
	for k, v := range map[string]any{
		"token": "dex-access", "refresh_token": "dex-refresh", "custom": "keep",
		"token_endpoint": idp.ts.URL + "/dex/token", "client_id": "dex-client",
	} {
		if got[k] != v {
			t.Errorf("config[%q] = %v, want %v", k, got[k], v)
		}
	}
	exp, _ := got["token_expires_at"].(int)
	if d := time.Until(time.Unix(int64(exp), 0)); d < 4*time.Minute || d > 5*time.Minute {
		t.Errorf("token_expires_at %v is not about 5 minutes from now", got["token_expires_at"])
	}
	assertMode0600(t, cfg)
}

func TestCmdLoginWebBrokerClearsRefreshData(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	stubBrowser(t, nil)
	cfg := writeTempFile(t, "cfg.yaml", "token: old\n"+staleSession)

	if out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if got := readConfig(t, cfg); got["token"] != "broker-pat" {
		t.Errorf("config = %v", got)
	}
	assertNoRefreshData(t, cfg)
}

func TestCmdLoginPATClearsRefreshData(t *testing.T) {
	resetGlobalFlags()
	ts := orgListServer(t)
	cfg := writeTempFile(t, "cfg.yaml", "token: old\n"+staleSession)

	if _, err := executeCommand("login", "--config", cfg, "--api-url", ts.URL, "--pat", "my-pat"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readConfig(t, cfg); got["token"] != "my-pat" {
		t.Errorf("config = %v", got)
	}
	assertNoRefreshData(t, cfg)
}

func TestCmdLogoutClearsRefreshData(t *testing.T) {
	resetGlobalFlags()
	cfg := writeTempFile(t, "cfg.yaml", "api_url: http://x\ntoken: old\n"+staleSession)

	if _, err := executeCommand("logout", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readConfig(t, cfg); got["token"] != nil || got["api_url"] != "http://x" {
		t.Errorf("config = %v", got)
	}
	assertNoRefreshData(t, cfg)
}
