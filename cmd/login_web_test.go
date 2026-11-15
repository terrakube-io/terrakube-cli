package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"

	"terrakube/internal/exit"
)

// fakeIdP serves /.well-known/terraform.json and a fake authorization and
// token endpoint, either as the Terrakube login broker or as Dex.
type fakeIdP struct {
	ts    *httptest.Server
	dex   bool
	ports [2]int

	// redirect, when set, replaces the query the authorization endpoint
	// sends back to redirect_uri (default: code and the request's state).
	redirect func(q url.Values) url.Values
	// token, when set, replaces the token endpoint's response.
	token http.HandlerFunc

	mu         sync.Mutex
	authzQuery url.Values
	tokenForm  url.Values
	tokenCalls int
}

// newFakeIdP starts the fake. Callers pick ports afterwards with setPorts,
// so the fake's own listener cannot take a port the test picked.
func newFakeIdP(t *testing.T, dex bool) *fakeIdP {
	t.Helper()
	f := &fakeIdP{dex: dex}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"login.v1": %s, "tfe.v2": "/remote/tfe/v2/"}`, f.loginV1())
	})
	authz := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f.mu.Lock()
		f.authzQuery = q
		f.mu.Unlock()
		back := url.Values{"code": {"the-code"}, "state": {q.Get("state")}}
		if f.redirect != nil {
			back = f.redirect(q)
		}
		http.Redirect(w, r, q.Get("redirect_uri")+"?"+back.Encode(), http.StatusFound)
	}
	token := func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.tokenForm = r.PostForm
		f.tokenCalls++
		f.mu.Unlock()
		if f.token != nil {
			f.token(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if f.dex {
			_, _ = fmt.Fprint(w, `{"access_token":"dex-access","id_token":"dex-id","refresh_token":"dex-refresh","token_type":"bearer","expires_in":300}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"access_token":"broker-pat","token_type":"Bearer","expires_in":2592000}`)
	}
	mux.HandleFunc("/oauth/authorize", authz)
	mux.HandleFunc("/oauth/token", token)
	mux.HandleFunc("/dex/auth", authz)
	mux.HandleFunc("/dex/token", token)
	f.ts = httptest.NewServer(mux)
	t.Cleanup(f.ts.Close)
	return f
}

func (f *fakeIdP) setPorts(low, high int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ports = [2]int{low, high}
}

func (f *fakeIdP) loginV1() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dex {
		return fmt.Sprintf(`{"client": "dex-client", "grant_types": ["authz_code", "openid", "profile", "email", "offline_access", "groups"],
			"authz": "%[1]s/dex/auth?scope=openid+profile+email+offline_access+groups", "token": "%[1]s/dex/token", "ports": [%[2]d, %[3]d]}`,
			f.ts.URL, f.ports[0], f.ports[1])
	}
	return fmt.Sprintf(`{"client": "terraform-cli", "grant_types": ["authz_code"],
		"authz": "%[1]s/oauth/authorize", "token": "%[1]s/oauth/token", "ports": [%[2]d, %[3]d]}`,
		f.ts.URL, f.ports[0], f.ports[1])
}

func (f *fakeIdP) snapshot() (authz, form url.Values, calls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.authzQuery, f.tokenForm, f.tokenCalls
}

// stubBrowser replaces openBrowser with a "browser" that follows the
// authorization URL (and its redirect to the loopback callback) in the
// background. It returns a func that reports the URLs it was given.
func stubBrowser(t *testing.T, openErr error) func() []string {
	t.Helper()
	var mu sync.Mutex
	var opened []string
	old := openBrowser
	openBrowser = func(u string) error {
		mu.Lock()
		opened = append(opened, u)
		mu.Unlock()
		go func() {
			if resp, err := http.Get(u); err == nil {
				_ = resp.Body.Close()
			}
		}()
		return openErr
	}
	t.Cleanup(func() { openBrowser = old })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), opened...)
	}
}

// freePort returns a loopback port that was free a moment ago.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

// busyThenFreePort returns a port p that the test keeps busy while p+1 is free.
func busyThenFreePort(t *testing.T) int {
	t.Helper()
	for range 50 {
		ln, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		next, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port+1))
		if err == nil {
			_ = next.Close()
			t.Cleanup(func() { _ = ln.Close() })
			return port
		}
		_ = ln.Close()
	}
	t.Skip("no adjacent free port pair found")
	return 0
}

func s256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func assertNoConfig(t *testing.T, cfg string) {
	t.Helper()
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Errorf("config must not be written, stat err: %v", err)
	}
}

func TestCmdLoginWebBroker(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	opened := stubBrowser(t, nil)
	cfg := writeTempFile(t, "cfg.yaml", "custom: keep\ntoken: old\n")
	if err := os.Chmod(cfg, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL+"/api/v1/")
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}

	authz, form, _ := idp.snapshot()
	redirectURI := fmt.Sprintf("http://localhost:%d/login", port)
	for k, want := range map[string]string{
		"client_id":             "terraform-cli",
		"response_type":         "code",
		"redirect_uri":          redirectURI,
		"code_challenge_method": "S256",
	} {
		if got := authz.Get(k); got != want {
			t.Errorf("authorize %s = %q, want %q", k, got, want)
		}
	}
	if len(authz.Get("state")) < 32 {
		t.Errorf("state %q is too short", authz.Get("state"))
	}
	for k, want := range map[string]string{
		"grant_type":   "authorization_code",
		"code":         "the-code",
		"redirect_uri": redirectURI,
		"client_id":    "terraform-cli",
	} {
		if got := form.Get(k); got != want {
			t.Errorf("token %s = %q, want %q", k, got, want)
		}
	}
	verifier := form.Get("code_verifier")
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("code_verifier length %d outside 43..128", len(verifier))
	}
	if s256(verifier) != authz.Get("code_challenge") {
		t.Errorf("code_challenge %q is not S256(%q)", authz.Get("code_challenge"), verifier)
	}

	if urls := opened(); len(urls) != 1 || !strings.Contains(out, urls[0]) {
		t.Errorf("browser opened %v; output must print the URL:\n%s", urls, out)
	}

	got := readConfig(t, cfg)
	want := map[string]any{"api_url": idp.ts.URL, "token": "broker-pat", "custom": "keep"}
	if len(got) != len(want) {
		t.Errorf("config = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("config[%q] = %v, want %v", k, got[k], v)
		}
	}
	assertMode0600(t, cfg)
}

func TestCmdLoginWebDex(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, true)
	port := freePort(t)
	idp.setPorts(port, port+1)
	stubBrowser(t, nil)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	if out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}

	authz, form, _ := idp.snapshot()
	if got := authz.Get("scope"); got != "openid profile email offline_access groups" {
		t.Errorf("scope = %q, the advertised scope must be kept", got)
	}
	if authz.Get("client_id") != "dex-client" || form.Get("client_id") != "dex-client" {
		t.Errorf("client_id authorize=%q token=%q, want dex-client", authz.Get("client_id"), form.Get("client_id"))
	}
	if s256(form.Get("code_verifier")) != authz.Get("code_challenge") {
		t.Error("code_challenge does not match code_verifier")
	}
	if got := readConfig(t, cfg); got["token"] != "dex-access" || got["api_url"] != idp.ts.URL {
		t.Errorf("config = %v, want the Dex access token", got)
	}
	assertMode0600(t, cfg)
}

func TestCmdLoginWebUsesConfiguredAPIURL(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	stubBrowser(t, nil)
	cfg := writeTempFile(t, "cfg.yaml", "api_url: "+idp.ts.URL+"\n")
	viper.Reset()

	if out, err := executeCommand("login", "--web", "--config", cfg); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if got := readConfig(t, cfg); got["token"] != "broker-pat" {
		t.Errorf("config = %v", got)
	}
}

func TestCmdLoginWebSkipsBusyPort(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := busyThenFreePort(t)
	idp.setPorts(port, port+1)
	stubBrowser(t, nil)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	if out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	authz, _, _ := idp.snapshot()
	if want := fmt.Sprintf("http://localhost:%d/login", port+1); authz.Get("redirect_uri") != want {
		t.Errorf("redirect_uri = %q, want %q", authz.Get("redirect_uri"), want)
	}
}

func TestCmdLoginWebAllPortsBusy(t *testing.T) {
	resetGlobalFlags()
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port
	idp := newFakeIdP(t, false)
	idp.setPorts(port, port)
	opened := stubBrowser(t, nil)
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	_, err = executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("no free port in %d-%d", port, port)) {
		t.Fatalf("expected busy port error, got %v", err)
	}
	if code := exit.Code(err); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if len(opened()) != 0 {
		t.Error("browser must not be opened")
	}
	assertNoConfig(t, cfg)
}

func TestCmdLoginWebCallbackFailures(t *testing.T) {
	tests := []struct {
		name     string
		redirect func(q url.Values) url.Values
		wantErr  string
	}{
		{
			name:     "state mismatch",
			redirect: func(url.Values) url.Values { return url.Values{"code": {"c"}, "state": {"forged"}} },
			wantErr:  "state does not match",
		},
		{
			name:     "missing state",
			redirect: func(url.Values) url.Values { return url.Values{"code": {"c"}} },
			wantErr:  "state does not match",
		},
		{
			name: "error from the authorization server",
			redirect: func(q url.Values) url.Values {
				return url.Values{"error": {"access_denied"}, "error_description": {"user said no"}, "state": {q.Get("state")}}
			},
			wantErr: `login failed: "access_denied" ("user said no")`,
		},
		{
			name:     "no code",
			redirect: func(q url.Values) url.Values { return url.Values{"state": {q.Get("state")}} },
			wantErr:  "no authorization code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			idp := newFakeIdP(t, false)
			port := freePort(t)
			idp.setPorts(port, port)
			idp.redirect = tt.redirect
			stubBrowser(t, nil)
			cfg := filepath.Join(t.TempDir(), "cfg.yaml")

			_, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if code := exit.Code(err); code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if _, _, calls := idp.snapshot(); calls != 0 {
				t.Errorf("token endpoint called %d times, want 0", calls)
			}
			assertNoConfig(t, cfg)
		})
	}
}

func TestCmdLoginWebTimeout(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	old := openBrowser
	openBrowser = func(string) error { return nil } // nobody completes the login
	t.Cleanup(func() { openBrowser = old })
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	_, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL, "--timeout", "150ms")
	if err == nil || !strings.Contains(err.Error(), "timed out after 150ms") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if code := exit.Code(err); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	assertNoConfig(t, cfg)
}

func TestCmdLoginWebBrowserFailsStillPrintsURL(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	opened := stubBrowser(t, errors.New("no display"))
	cfg := filepath.Join(t.TempDir(), "cfg.yaml")

	out, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if urls := opened(); len(urls) != 1 || !strings.Contains(out, urls[0]) || !strings.Contains(out, "no display") {
		t.Errorf("output must print the URL and the browser error:\n%s", out)
	}
}

func TestCmdLoginWebTokenEndpointFails(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "invalid_grant",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, `{"error":"invalid_grant"}`)
			},
			wantErr: `token request failed: 400 Bad Request: "invalid_grant"`,
		},
		{
			name:    "no access token",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, `{}`) },
			wantErr: "token request failed: 200 OK",
		},
		{
			name:    "connection dropped",
			handler: func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) },
			wantErr: "token request failed: Post",
		},
		{
			name:    "not json",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, `<html>`) },
			wantErr: "token request failed: 200 OK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			idp := newFakeIdP(t, false)
			port := freePort(t)
			idp.setPorts(port, port)
			idp.token = tt.handler
			stubBrowser(t, nil)
			cfg := filepath.Join(t.TempDir(), "cfg.yaml")

			_, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			assertNoConfig(t, cfg)
		})
	}
}

func TestCmdLoginWebDiscoveryErrors(t *testing.T) {
	login := func(body string) string {
		return `{"login.v1": ` + body + `}`
	}
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{"not found", http.StatusNotFound, "", "404 Not Found"},
		{"not json", http.StatusOK, "<html>", "parsing login settings"},
		{"no login.v1", http.StatusOK, `{"tfe.v2": "/remote/tfe/v2/"}`, "does not advertise login.v1"},
		{"no authz_code grant", http.StatusOK, login(`{"client":"c","grant_types":["device"],"authz":"/a","token":"/t","ports":[10000,10001]}`), "authz_code"},
		{"no client", http.StatusOK, login(`{"grant_types":["authz_code"],"authz":"/a","token":"/t","ports":[10000,10001]}`), "incomplete"},
		{"no token url", http.StatusOK, login(`{"client":"c","grant_types":["authz_code"],"authz":"/a","ports":[10000,10001]}`), "incomplete"},
		{"one port", http.StatusOK, login(`{"client":"c","grant_types":["authz_code"],"authz":"/a","token":"/t","ports":[10000]}`), "invalid ports"},
		{"reversed ports", http.StatusOK, login(`{"client":"c","grant_types":["authz_code"],"authz":"/a","token":"/t","ports":[10001,10000]}`), "invalid ports"},
		{"bad authz url", http.StatusOK, login(`{"client":"c","grant_types":["authz_code"],"authz":"http://[::1","token":"/t","ports":[10000,10001]}`), "invalid authz URL"},
		{"bad token url", http.StatusOK, login(`{"client":"c","grant_types":["authz_code"],"authz":"/a","token":"http://[::1","ports":[10000,10001]}`), "invalid token URL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer ts.Close()
			cfg := filepath.Join(t.TempDir(), "cfg.yaml")

			_, err := executeCommand("login", "--web", "--config", cfg, "--api-url", ts.URL)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			assertNoConfig(t, cfg)
		})
	}
}

func TestCmdLoginWebArgumentErrors(t *testing.T) {
	unreachable := httptest.NewServer(http.NotFoundHandler())
	unreachable.Close()
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"no api url", nil, "--api-url, TERRAKUBE_API_URL or api_url in the config is required"},
		{"with pat", []string{"--api-url", "http://x", "--pat", "p"}, "none of the others can be"},
		{"zero timeout", []string{"--api-url", "http://x", "--timeout", "0s"}, "--timeout must be positive"},
		{"bad api url", []string{"--api-url", "http://[::1"}, "reading login settings"},
		{"unreachable", []string{"--api-url", unreachable.URL}, "reading login settings"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalFlags()
			cfg := filepath.Join(t.TempDir(), "cfg.yaml")

			_, err := executeCommand(append([]string{"login", "--web", "--config", cfg}, tt.args...)...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			assertNoConfig(t, cfg)
		})
	}
}

func TestCmdLoginWebSaveFails(t *testing.T) {
	resetGlobalFlags()
	idp := newFakeIdP(t, false)
	port := freePort(t)
	idp.setPorts(port, port)
	stubBrowser(t, nil)
	cfg := filepath.Join(t.TempDir(), "cfg.json")

	_, err := executeCommand("login", "--web", "--config", cfg, "--api-url", idp.ts.URL)
	if err == nil || !strings.Contains(err.Error(), "saving configuration") {
		t.Fatalf("expected saving configuration error, got %v", err)
	}
}

func TestLoginCallbackIgnoresOtherPaths(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	base := "http://" + ln.Addr().String()
	go func() {
		for _, p := range []string{"/favicon.ico", "/login?state=s&code=c"} {
			if resp, err := http.Get(base + p); err == nil {
				if p == "/favicon.ico" && resp.StatusCode != http.StatusNotFound {
					t.Errorf("GET %s = %d, want 404", p, resp.StatusCode)
				}
				_ = resp.Body.Close()
			}
		}
	}()
	code, err := waitForCallback(t.Context(), ln, "s", 5*time.Second)
	if err != nil || code != "c" {
		t.Fatalf("code = %q, err = %v", code, err)
	}
}

func TestLoginCallbackContextCancelled(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := waitForCallback(ctx, ln, "s", time.Minute); err == nil {
		t.Fatal("expected an error for a cancelled context")
	}
}

func TestPKCEChallengeS256(t *testing.T) {
	// RFC 7636 appendix B.
	if got := pkceChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Errorf("challenge = %q", got)
	}
	a, b := randomURLString(), randomURLString()
	if a == b || len(a) != 43 {
		t.Errorf("random strings %q, %q: want distinct 43-char values", a, b)
	}
}

func TestBrowserCommand(t *testing.T) {
	const u = "https://idp.example/auth?a=1&b=2"
	tests := []struct {
		goos, name string
		args       []string
	}{
		{"darwin", "open", []string{u}},
		{"linux", "xdg-open", []string{u}},
		{"freebsd", "xdg-open", []string{u}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", u}},
	}
	for _, tt := range tests {
		name, args := browserCommand(tt.goos, u)
		if name != tt.name || strings.Join(args, " ") != strings.Join(tt.args, " ") {
			t.Errorf("%s: got %s %v, want %s %v", tt.goos, name, args, tt.name, tt.args)
		}
	}
}

func TestOpenBrowserRunsCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as the fake browser")
	}
	dir := t.TempDir()
	got := filepath.Join(dir, "opened")
	name, _ := browserCommand(runtime.GOOS, "")
	script := "#!/bin/sh\nprintf '%s' \"$1\" > " + got + "\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	if err := openBrowser("https://idp.example/auth"); err != nil {
		t.Fatalf("openBrowser: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if b, err := os.ReadFile(got); err == nil && string(b) == "https://idp.example/auth" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("fake browser was not run with the URL")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestLoginV1JSONTag(t *testing.T) {
	var doc wellKnown
	if err := json.Unmarshal([]byte(`{"login.v1":{"client":"c"}}`), &doc); err != nil || doc.Login == nil || doc.Login.Client != "c" {
		t.Fatalf("login.v1 not decoded: %+v, %v", doc, err)
	}
}
