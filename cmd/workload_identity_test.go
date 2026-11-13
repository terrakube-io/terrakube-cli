package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"
)

// isolateOIDCEnv points HOME at a temp dir and clears every OIDC token source.
func isolateOIDCEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir on Windows
	t.Setenv("APPDATA", home)
	t.Setenv("TERRAKUBE_OIDC_TOKEN", "")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "")
	for _, p := range ciProviders { // tests may themselves run in CI
		t.Setenv(p.detect, "")
	}
	for _, v := range []string{"BITBUCKET_STEP_OIDC_TOKEN", "SYSTEM_OIDCREQUESTURI", "SYSTEM_ACCESSTOKEN", "AZURESUBSCRIPTION_SERVICE_CONNECTION_ID"} {
		t.Setenv(v, "")
	}
	return home
}

// terrakubeServer accepts only requests bearing wantToken.
func terrakubeServer(t *testing.T, wantToken string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+wantToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_ = jsonapi.MarshalPayload(w, []*terrakube.Organization{})
	}))
	t.Cleanup(ts.Close)
	return ts
}

func TestCmdLoginOIDCGitHubActionsWritesTerraformCredentials(t *testing.T) {
	resetGlobalFlags()
	home := isolateOIDCEnv(t)

	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer gh-request-token" ||
			r.URL.Query().Get("audience") != "terrakube" || r.URL.Query().Get("api-version") != "2.0" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(`{"value":"gh-oidc-jwt"}`))
	}))
	defer gh.Close()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", gh.URL+"/token?api-version=2.0")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "gh-request-token")
	t.Setenv("GITHUB_ACTIONS", "true")

	ts := terrakubeServer(t, "gh-oidc-jwt")

	credsPath := filepath.Join(home, ".terraform.d", "credentials.tfrc.json")
	if runtime.GOOS == "windows" {
		credsPath = filepath.Join(home, "terraform.d", "credentials.tfrc.json")
	}
	if err := os.MkdirAll(filepath.Dir(credsPath), 0o700); err != nil {
		t.Fatal(err)
	}
	existing := `{"credentials":{"app.terraform.io":{"token":"keep-me"}}}`
	if err := os.WriteFile(credsPath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand("login", "--api-url", ts.URL+"/api/v1/", "--oidc",
		"--oidc-audience", "terrakube", "--terraform-credentials")
	if err != nil {
		t.Fatalf("unexpected error: %v (out=%s)", err, out)
	}
	if viper.GetString("token") != "gh-oidc-jwt" || viper.GetString("api_url") != ts.URL {
		t.Errorf("unexpected config: api_url=%q token=%q", viper.GetString("api_url"), viper.GetString("token"))
	}

	b, err := os.ReadFile(credsPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Credentials map[string]struct{ Token string } `json:"credentials"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	host := strings.TrimPrefix(ts.URL, "http://")
	if doc.Credentials[host].Token != "gh-oidc-jwt" {
		t.Errorf("expected token for %s, got %s", host, b)
	}
	if doc.Credentials["app.terraform.io"].Token != "keep-me" {
		t.Errorf("existing credentials were not preserved: %s", b)
	}
	if runtime.GOOS != "windows" {
		for _, p := range []string{credsPath, filepath.Join(home, ".terrakube-cli.yaml")} {
			fi, err := os.Stat(p)
			if err != nil {
				t.Errorf("stat %s: %v", p, err)
			} else if fi.Mode().Perm() != 0o600 {
				t.Errorf("%s: expected mode 0600, got %v", p, fi.Mode().Perm())
			}
		}
	}
}

func TestCmdLoginOIDCEnvTokenCustomHosts(t *testing.T) {
	resetGlobalFlags()
	home := isolateOIDCEnv(t)
	t.Setenv("TERRAKUBE_OIDC_TOKEN", " gitlab-jwt\n")
	ts := terrakubeServer(t, "gitlab-jwt")

	_, err := executeCommand("login", "-a", ts.URL, "--oidc", "--terraform-credentials",
		"--terraform-host", "API.example.com", "--terraform-host", "https://reg.example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path, _ := terraformCredentialsPath()
	if !strings.HasPrefix(path, home) {
		t.Fatalf("credentials path %s not under temp home %s", path, home)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), `"api.example.com"`) || !strings.Contains(string(b), `"reg.example.com"`) ||
		strings.Contains(string(b), "127.0.0.1") {
		t.Errorf("unexpected credentials file: %s", b)
	}
}

func TestCmdLoginRejectedTokenFails(t *testing.T) {
	resetGlobalFlags()
	isolateOIDCEnv(t)
	tokenFile := filepath.Join(t.TempDir(), "token")
	_ = os.WriteFile(tokenFile, []byte("wrong-jwt"), 0o600)
	ts := terrakubeServer(t, "right-jwt")

	if _, err := executeCommand("login", "-a", ts.URL, "--oidc", "--oidc-token-file", tokenFile); err == nil {
		t.Fatal("expected login to fail when the server rejects the token")
	}
}

func TestResolveOIDCToken(t *testing.T) {
	isolateOIDCEnv(t)

	if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "no OIDC token found") {
		t.Errorf("expected 'no OIDC token found' error, got %v", err)
	}

	// The token file wins over the environment variable.
	t.Setenv("TERRAKUBE_OIDC_TOKEN", "env-jwt")
	tokenFile := filepath.Join(t.TempDir(), "token")
	_ = os.WriteFile(tokenFile, []byte("file-jwt\n"), 0o600)
	if tok, err := resolveOIDCToken(tokenFile, ""); err != nil || tok != "file-jwt" {
		t.Errorf("expected file-jwt, got %q (err=%v)", tok, err)
	}
	if tok, err := resolveOIDCToken("", ""); err != nil || tok != "env-jwt" {
		t.Errorf("expected env-jwt, got %q (err=%v)", tok, err)
	}

	_ = os.WriteFile(tokenFile, []byte("  \n"), 0o600)
	if _, err := resolveOIDCToken(tokenFile, ""); err == nil {
		t.Error("expected error for empty token file")
	}
}

func TestCmdLoginOIDCRequiresHTTPS(t *testing.T) {
	resetGlobalFlags()
	isolateOIDCEnv(t)
	t.Setenv("TERRAKUBE_OIDC_TOKEN", "jwt")
	_, err := executeCommand("login", "-a", "http://terrakube.example.com", "--oidc")
	if err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("expected https error, got %v", err)
	}
}

func TestCmdLoginPatAndOIDCMutuallyExclusive(t *testing.T) {
	resetGlobalFlags()
	isolateOIDCEnv(t)
	_, err := executeCommand("login", "-a", "http://localhost:1", "--pat", "x", "--oidc")
	if err == nil || !strings.Contains(err.Error(), "none of the others can be") {
		t.Fatalf("expected mutually exclusive flags error, got %v", err)
	}
}

func TestResolveOIDCTokenCIProviders(t *testing.T) {
	isolateOIDCEnv(t)

	t.Run("bitbucket", func(t *testing.T) {
		t.Setenv("BITBUCKET_BUILD_NUMBER", "7")
		if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "oidc: true") {
			t.Errorf("expected hint about oidc: true, got %v", err)
		}
		t.Setenv("BITBUCKET_STEP_OIDC_TOKEN", "bb-jwt")
		if tok, err := resolveOIDCToken("", ""); err != nil || tok != "bb-jwt" {
			t.Errorf("expected bb-jwt, got %q (err=%v)", tok, err)
		}
	})

	t.Run("gitlab", func(t *testing.T) {
		t.Setenv("GITLAB_CI", "true")
		if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "id_tokens") {
			t.Errorf("expected id_tokens hint, got %v", err)
		}
		t.Setenv("TERRAKUBE_OIDC_TOKEN", "gl-jwt")
		if tok, err := resolveOIDCToken("", ""); err != nil || tok != "gl-jwt" {
			t.Errorf("expected gl-jwt, got %q (err=%v)", tok, err)
		}
	})

	t.Run("azure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer ado-access" ||
				q.Get("api-version") != "7.1" || q.Get("serviceConnectionId") != "sc-1" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_, _ = w.Write([]byte(`{"oidcToken":"ado-jwt"}`))
		}))
		defer srv.Close()
		t.Setenv("TF_BUILD", "True")
		t.Setenv("SYSTEM_OIDCREQUESTURI", srv.URL+"/oidctoken")
		t.Setenv("SYSTEM_ACCESSTOKEN", "ado-access")
		if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "AZURESUBSCRIPTION_SERVICE_CONNECTION_ID") {
			t.Errorf("expected missing service connection error, got %v", err)
		}
		t.Setenv("AZURESUBSCRIPTION_SERVICE_CONNECTION_ID", "sc-1")
		if tok, err := resolveOIDCToken("", ""); err != nil || tok != "ado-jwt" {
			t.Errorf("expected ado-jwt, got %q (err=%v)", tok, err)
		}
	})

	t.Run("github requires audience", func(t *testing.T) {
		t.Setenv("GITHUB_ACTIONS", "true")
		t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "http://127.0.0.1:1/token")
		t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "x")
		if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "--oidc-audience") {
			t.Errorf("expected --oidc-audience error, got %v", err)
		}
	})

	t.Run("github missing permission", func(t *testing.T) {
		t.Setenv("GITHUB_ACTIONS", "true")
		if _, err := resolveOIDCToken("", ""); err == nil || !strings.Contains(err.Error(), "id-token: write") {
			t.Errorf("expected id-token permission hint, got %v", err)
		}
	})
}

func TestTerraformHost(t *testing.T) {
	for in, want := range map[string]string{
		"Reg.Example.com":              "reg.example.com",
		"host:8443":                    "host:8443",
		"https://reg.example.com/":     "reg.example.com",
		"http://127.0.0.1:8080/api/v1": "127.0.0.1:8080",
	} {
		if got, err := terraformHost(in); err != nil || got != want {
			t.Errorf("terraformHost(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "reg.example.com/path", "https://"} {
		if _, err := terraformHost(in); err == nil {
			t.Errorf("terraformHost(%q): expected error", in)
		}
	}
}
