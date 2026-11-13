package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ciProvider obtains a workload identity token from the CI platform it runs on.
type ciProvider struct {
	name   string
	detect string // environment variable the platform always sets
	token  func(audience string) (string, error)
}

// ciProviders is checked in order; the first detected platform supplies the token.
var ciProviders = []ciProvider{
	{"GitHub Actions", "GITHUB_ACTIONS", githubOIDCToken},
	{"GitLab CI", "GITLAB_CI", gitlabOIDCToken},
	{"Bitbucket Pipelines", "BITBUCKET_BUILD_NUMBER", bitbucketOIDCToken},
	{"Azure Pipelines", "TF_BUILD", azureOIDCToken},
}

// resolveOIDCToken returns a CI/CD workload identity token from the first available
// source: token file, TERRAKUBE_OIDC_TOKEN, or the detected CI platform.
func resolveOIDCToken(tokenFile, audience string) (string, error) {
	if tokenFile != "" {
		b, err := os.ReadFile(tokenFile)
		if err != nil {
			return "", fmt.Errorf("reading OIDC token file: %w", err)
		}
		return nonEmptyToken(string(b), tokenFile)
	}
	if t := os.Getenv("TERRAKUBE_OIDC_TOKEN"); t != "" {
		return nonEmptyToken(t, "TERRAKUBE_OIDC_TOKEN")
	}
	for _, p := range ciProviders {
		if os.Getenv(p.detect) == "" {
			continue
		}
		t, err := p.token(audience)
		if err != nil {
			return "", fmt.Errorf("%s: %w", p.name, err)
		}
		return nonEmptyToken(t, p.name)
	}
	names := make([]string, len(ciProviders))
	for i, p := range ciProviders {
		names[i] = p.name
	}
	return "", fmt.Errorf("no OIDC token found: not running in a supported CI (%s); "+
		"use --oidc-token-file or set TERRAKUBE_OIDC_TOKEN", strings.Join(names, ", "))
}

func nonEmptyToken(t, source string) (string, error) {
	if t = strings.TrimSpace(t); t == "" {
		return "", fmt.Errorf("OIDC token from %s is empty", source)
	}
	return t, nil
}

// requireEnv returns the values of the given variables, or an error naming the
// first missing one together with the configuration hint.
func requireEnv(hint string, names ...string) ([]string, error) {
	vals := make([]string, len(names))
	for i, n := range names {
		if vals[i] = os.Getenv(n); vals[i] == "" {
			return nil, fmt.Errorf("%s is not set: %s", n, hint)
		}
	}
	return vals, nil
}

// fetchJSONToken sends req and returns the string under field in the JSON response.
func fetchJSONToken(req *http.Request, field string) (string, error) {
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting OIDC token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("requesting OIDC token: unexpected status %s", resp.Status)
	}
	var body map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", fmt.Errorf("decoding OIDC token response: %w", err)
	}
	t, _ := body[field].(string)
	return t, nil
}

// withQuery returns raw with the given query parameters set (empty values skipped).
func withQuery(raw string, kv ...string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i+1] != "" {
			q.Set(kv[i], kv[i+1])
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// githubOIDCToken requests an ID token from the GitHub Actions OIDC provider.
// https://docs.github.com/en/actions/reference/security/oidc
func githubOIDCToken(audience string) (string, error) {
	env, err := requireEnv(`add "permissions: id-token: write" to the workflow or job`,
		"ACTIONS_ID_TOKEN_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if err != nil {
		return "", err
	}
	if audience == "" {
		// The default audience (https://github.com/<owner>) is trusted by other relying parties too.
		return "", errors.New("--oidc-audience is required: set it to the audience of the Terrakube federated credential")
	}
	u, err := withQuery(env[0], "audience", audience)
	if err != nil {
		return "", fmt.Errorf("parsing ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	req, err := http.NewRequestWithContext(getContext(), http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+env[1])
	return fetchJSONToken(req, "value")
}

// gitlabOIDCToken: GitLab only exposes ID tokens declared in the job, and
// resolveOIDCToken already reads TERRAKUBE_OIDC_TOKEN, so reaching here means it is missing.
// https://docs.gitlab.com/ci/secrets/id_token_authentication/
func gitlabOIDCToken(string) (string, error) {
	return "", errors.New("TERRAKUBE_OIDC_TOKEN is not set: declare it in the job as " +
		"\"id_tokens: { TERRAKUBE_OIDC_TOKEN: { aud: <federated audience> } }\"")
}

// bitbucketOIDCToken reads the step ID token; the audience is set in bitbucket-pipelines.yml.
// https://support.atlassian.com/bitbucket-cloud/docs/integrate-pipelines-with-resource-servers-using-oidc/
func bitbucketOIDCToken(string) (string, error) {
	env, err := requireEnv(`add "oidc: true" to the pipeline step`, "BITBUCKET_STEP_OIDC_TOKEN")
	if err != nil {
		return "", err
	}
	return env[0], nil
}

// azureOIDCToken requests an ID token for a workload identity service connection.
// The audience is fixed by Azure DevOps.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/distributedtask/oidctoken/create
func azureOIDCToken(string) (string, error) {
	env, err := requireEnv("map it in the step, e.g. \"env: { SYSTEM_ACCESSTOKEN: $(System.AccessToken), "+
		"AZURESUBSCRIPTION_SERVICE_CONNECTION_ID: <service connection id> }\"",
		"SYSTEM_OIDCREQUESTURI", "SYSTEM_ACCESSTOKEN", "AZURESUBSCRIPTION_SERVICE_CONNECTION_ID")
	if err != nil {
		return "", err
	}
	u, err := withQuery(env[0], "api-version", "7.1", "serviceConnectionId", env[2])
	if err != nil {
		return "", fmt.Errorf("parsing SYSTEM_OIDCREQUESTURI: %w", err)
	}
	req, err := http.NewRequestWithContext(getContext(), http.MethodPost, u, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+env[1])
	req.Header.Set("Content-Type", "application/json")
	return fetchJSONToken(req, "oidcToken")
}

// terraformHost turns a URL or bare hostname[:port] into a credentials.tfrc.json key.
func terraformHost(v string) (string, error) {
	if strings.Contains(v, "://") {
		u, err := url.Parse(v)
		if err != nil || u.Host == "" {
			return "", fmt.Errorf("cannot determine host from %q", v)
		}
		v = u.Host
	}
	if v == "" || strings.ContainsAny(v, "/?#@ ") {
		return "", fmt.Errorf("invalid Terraform host %q: expected hostname[:port]", v)
	}
	return strings.ToLower(v), nil
}

// terraformCredentialsPath is where Terraform and OpenTofu look for CLI credentials.
// https://developer.hashicorp.com/terraform/cli/config/config-file#credentials-1
func terraformCredentialsPath() (string, error) {
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir() // %APPDATA%
		return filepath.Join(dir, "terraform.d", "credentials.tfrc.json"), err
	}
	home, err := os.UserHomeDir()
	return filepath.Join(home, ".terraform.d", "credentials.tfrc.json"), err
}

// writeTerraformCredentials sets the token for each host in credentials.tfrc.json,
// keeping credentials for other hosts intact.
func writeTerraformCredentials(hosts []string, token string) (string, error) {
	path, err := terraformCredentialsPath()
	if err != nil {
		return "", fmt.Errorf("locating Terraform credentials file: %w", err)
	}

	doc := map[string]any{}
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &doc); err != nil {
			return "", fmt.Errorf("parsing %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	creds, _ := doc["credentials"].(map[string]any)
	if creds == nil {
		creds = map[string]any{}
	}
	for _, h := range hosts {
		creds[h] = map[string]string{"token": token}
	}
	doc["credentials"] = creds

	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("creating Terraform config directory: %w", err)
	}
	// Write a 0600 temp file and rename it over the target so a failed write never
	// corrupts credentials for other hosts.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".credentials-*.tfrc.json")
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // no-op after a successful rename
	if _, err = tmp.Write(append(b, '\n')); err == nil {
		err = tmp.Close()
	} else {
		_ = tmp.Close()
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	return path, nil
}
