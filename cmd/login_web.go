package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// wellKnown is the part of /.well-known/terraform.json that login --web reads.
type wellKnown struct {
	Login *loginV1 `json:"login.v1"`
}

// loginV1 is the login.v1 service: either the Terrakube login broker
// (<api>/oauth/authorize, which issues a Terrakube token) or Dex itself
// (<dex>/auth, which issues a short-lived token and a refresh_token).
type loginV1 struct {
	Client     string   `json:"client"`
	GrantTypes []string `json:"grant_types"`
	Authz      string   `json:"authz"`
	Token      string   `json:"token"`
	Ports      []int    `json:"ports"`
}

// tokenResponse is an OAuth 2.0 token endpoint response.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// loginHTTPClient talks to the discovery and token endpoints.
var loginHTTPClient = &http.Client{Timeout: 30 * time.Second}

// openBrowser opens u in the default browser. Tests replace it.
var openBrowser = func(u string) error {
	name, args := browserCommand(runtime.GOOS, u)
	c := exec.Command(name, args...)
	if err := c.Start(); err != nil {
		return err
	}
	go func() { _ = c.Wait() }()
	return nil
}

// browserCommand returns the command that opens u in the browser on goos.
func browserCommand(goos, u string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{u}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", u}
	default:
		return "xdg-open", []string{u}
	}
}

// loginWeb logs in through the browser: it reads login.v1 from the server,
// runs the authorization code flow with PKCE (S256) on a loopback port from
// the advertised range and saves the token. Progress goes to w.
func loginWeb(w io.Writer, timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	api := apiURL
	if api == "" {
		api = viper.GetString("api_url") // TERRAKUBE_API_URL, then the config
	}
	if api == "" {
		return errors.New("--api-url, TERRAKUBE_API_URL or api_url in the config is required")
	}
	api = trimAPIURL(api)

	ctx := getContext()
	svc, authzURL, tokenURL, err := discoverLogin(ctx, api)
	if err != nil {
		return err
	}

	ln, port, err := listenLoopback(svc.Ports[0], svc.Ports[1])
	if err != nil {
		return err
	}
	verifier, state := randomURLString(), randomURLString()
	redirectURI := fmt.Sprintf("http://localhost:%d/login", port)
	q := authzURL.Query()
	q.Set("client_id", svc.Client)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", pkceChallenge(verifier))
	q.Set("code_challenge_method", "S256")
	authzURL.RawQuery = q.Encode()

	_, _ = fmt.Fprintf(w, "Open this URL in your browser to log in:\n\n  %s\n\n", authzURL)
	if err := openBrowser(authzURL.String()); err != nil {
		_, _ = fmt.Fprintf(w, "Could not open a browser (%v); open the URL above yourself.\n", err)
	}

	code, err := waitForCallback(ctx, ln, state, timeout)
	if err != nil {
		return err
	}
	tok, err := requestToken(ctx, tokenURL, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirectURI},
		"client_id":     {svc.Client},
	})
	if err != nil {
		return err
	}

	if err := updateConfig(func(settings map[string]any) {
		settings["api_url"] = api
		settings["token"] = tok.AccessToken
		delete(settings, "pat") // written by older versions of login
	}); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	fmt.Println("Successfully logged in to Terrakube server")
	return nil
}

// discoverLogin reads and checks login.v1 from api's /.well-known/terraform.json.
// Relative authz and token URLs are resolved against the discovery URL.
func discoverLogin(ctx context.Context, api string) (*loginV1, *url.URL, string, error) {
	discovery := api + "/.well-known/terraform.json"
	fail := func(err error) (*loginV1, *url.URL, string, error) {
		return nil, nil, "", fmt.Errorf("reading login settings from %s: %w", discovery, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discovery, nil)
	if err != nil {
		return fail(err)
	}
	resp, err := loginHTTPClient.Do(req)
	if err != nil {
		return fail(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fail(errors.New(resp.Status))
	}
	var doc wellKnown
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&doc); err != nil {
		return fail(fmt.Errorf("parsing login settings: %w", err))
	}
	svc := doc.Login
	switch {
	case svc == nil:
		return fail(errors.New("the server does not advertise login.v1"))
	case !slices.Contains(svc.GrantTypes, "authz_code"):
		return fail(errors.New("login.v1 does not support the authz_code grant"))
	case svc.Client == "" || svc.Authz == "" || svc.Token == "":
		return fail(errors.New("login.v1 is incomplete: client, authz and token are required"))
	case len(svc.Ports) != 2 || svc.Ports[0] < 1 || svc.Ports[0] > svc.Ports[1] || svc.Ports[1] > 65535:
		return fail(fmt.Errorf("login.v1 has invalid ports %v", svc.Ports))
	}
	base := req.URL
	authz, err := base.Parse(svc.Authz)
	if err != nil {
		return fail(fmt.Errorf("login.v1 has an invalid authz URL: %w", err))
	}
	token, err := base.Parse(svc.Token)
	if err != nil {
		return fail(fmt.Errorf("login.v1 has an invalid token URL: %w", err))
	}
	return svc, authz, token.String(), nil
}

// listenLoopback listens on the first free port in [low, high] on 127.0.0.1.
func listenLoopback(low, high int) (net.Listener, int, error) {
	for port := low; port <= high; port++ {
		if ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
			return ln, port, nil
		}
	}
	return nil, 0, fmt.Errorf("no free port in %d-%d for the login callback; close what is using them and retry", low, high)
}

// waitForCallback serves the redirect_uri on ln until the first request to
// /login, and returns its authorization code. A state other than the one
// sent, an error from the authorization server, or no request within
// timeout fails the login.
func waitForCallback(ctx context.Context, ln net.Listener, state string, timeout time.Duration) (string, error) {
	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)
	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/login" {
				http.NotFound(w, r)
				return
			}
			q := r.URL.Query()
			var res result
			switch {
			case q.Get("state") != state:
				res.err = errors.New("login callback state does not match; run terrakube login --web again")
			case q.Get("error") != "":
				res.err = fmt.Errorf("login failed: %q", q.Get("error"))
				if d := q.Get("error_description"); d != "" {
					res.err = fmt.Errorf("%w (%q)", res.err, d)
				}
			case q.Get("code") == "":
				res.err = errors.New("login callback has no authorization code")
			default:
				res.code = q.Get("code")
			}
			if res.err != nil {
				http.Error(w, "Login failed. Return to the terminal for details.", http.StatusBadRequest)
			} else {
				_, _ = fmt.Fprintln(w, "Login complete. You can close this window and return to the terminal.")
			}
			select {
			case results <- res:
			default: // only the first callback counts
			}
		}),
	}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		// Let the handler finish its response before the server goes away.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case res := <-results:
		return res.code, res.err
	case <-timer.C:
		return "", fmt.Errorf("timed out after %s waiting for the browser login", timeout)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// requestToken posts form to the token endpoint and returns its response.
func requestToken(ctx context.Context, tokenURL string, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := loginHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var tok tokenResponse
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tok)
	if resp.StatusCode != http.StatusOK || tok.AccessToken == "" {
		if tok.Error != "" {
			return nil, fmt.Errorf("token request failed: %s: %q", resp.Status, tok.Error)
		}
		return nil, fmt.Errorf("token request failed: %s", resp.Status)
	}
	return &tok, nil
}

// randomURLString returns 32 random bytes, base64url-encoded (43 characters):
// a PKCE code verifier or an OAuth state.
func randomURLString() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // never fails (crypto/rand)
	return base64.RawURLEncoding.EncodeToString(b)
}

// pkceChallenge returns the S256 code challenge for verifier (RFC 7636).
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
