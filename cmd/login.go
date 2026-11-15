package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
	"gopkg.in/yaml.v3"
)

const loginLong = `
Authenticate to a remote Terrakube server using a Personal Access Token (PAT),
with --web through the browser, or with --oidc inside CI/CD using the
platform's workload identity (OIDC) token.

--web reads login.v1 from the server's /.well-known/terraform.json, opens the
browser (the URL is also printed) and waits for the login on a local port from
the advertised range. With the Terrakube login broker it saves a Terrakube
token. With Dex it also saves the refresh token and renews the Dex token
before a command uses it once it expires.

The API URL should be provided in the following format:
  http://your-server:port
  https://your-server:port

Note: Do not include /api/v1 in the URL as it will be added automatically.

--oidc sends the CI platform's OIDC token to Terrakube as-is. The server accepts
it when a federated credential matches its issuer, audience and claims (see
"terrakube federated" and "terrakube federated-claim"). The token is taken from
the first available source:
  1. --oidc-token-file (e.g. a Kubernetes projected service account token)
  2. TERRAKUBE_OIDC_TOKEN environment variable
  3. The CI platform the CLI detects it is running on:
     GitHub Actions       requires "permissions: id-token: write" and --oidc-audience
     GitLab CI            id_tokens: { TERRAKUBE_OIDC_TOKEN: { aud: <audience> } }
     Bitbucket Pipelines  "oidc: true" on the step (BITBUCKET_STEP_OIDC_TOKEN)
     Azure Pipelines      workload identity service connection; map SYSTEM_ACCESSTOKEN
                          and AZURESUBSCRIPTION_SERVICE_CONNECTION_ID into the step env

OIDC tokens are short-lived: run login right before the commands that need it.
--oidc requires an https API URL. The token is stored on disk, so clean up HOME
on persistent self-hosted runners. On Azure Pipelines the token's audience is
fixed to Azure AD: use a service connection with no Azure role assignments.

--terraform-credentials also stores the token in the Terraform/OpenTofu CLI
credentials file (credentials.tfrc.json), so "terraform init" can use Terrakube
as a remote backend or private registry without extra setup.
`

var loginExamples = `
Login to a local Terrakube server using PAT
  %[1]v login -a http://localhost:8080 -t your-pat-token

Login to a remote Terrakube server
  %[1]v login --api-url https://terrakube.example.com --pat your-pat-token

Login through the browser (--api-url can be omitted when the config has api_url)
  %[1]v login --web --api-url https://terrakube.example.com

Login using environment variables
  TERRAKUBE_API_URL=https://terrakube.example.com TERRAKUBE_PAT=your-pat-token %[1]v login

Login from GitHub Actions and write Terraform credentials
  %[1]v login -a https://terrakube.example.com --oidc --oidc-audience terrakube --terraform-credentials

Login from GitLab CI, Bitbucket Pipelines or Azure Pipelines (the token source is detected)
  %[1]v login -a https://terrakube.example.com --oidc --terraform-credentials

Also write Terraform credentials for a separate registry host
  %[1]v login -a https://terrakube-api.example.com --oidc --terraform-credentials \
    --terraform-host terrakube-api.example.com --terraform-host terrakube-reg.example.com
`

var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "login to Terrakube server",
	Long:         loginLong,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if web, _ := cmd.Flags().GetBool("web"); web {
			timeout, _ := cmd.Flags().GetDuration("timeout")
			return loginWeb(cmd.ErrOrStderr(), timeout)
		}
		if oidcLogin {
			return loginOIDC(cmd.OutOrStdout())
		}
		return login(cmd.OutOrStdout())
	},
	Example: fmt.Sprintf(loginExamples, rootCmd.Use),
}

var (
	apiURL               string
	patToken             string
	oidcLogin            bool
	oidcAudience         string
	oidcTokenFile        string
	terraformCredentials bool
	terraformHosts       []string
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&apiURL, "api-url", "a", "", "API URL (or TERRAKUBE_API_URL)")
	loginCmd.Flags().StringVarP(&patToken, "pat", "t", "", "Personal Access Token (or TERRAKUBE_PAT)")
	loginCmd.Flags().Bool("web", false, "Log in through the browser instead of with a PAT")
	loginCmd.Flags().Duration("timeout", 5*time.Minute, "With --web: how long to wait for the browser login")
	loginCmd.Flags().BoolVar(&oidcLogin, "oidc", false, "Log in with the CI/CD workload identity (OIDC) token instead of a PAT")
	loginCmd.Flags().StringVar(&oidcAudience, "oidc-audience", "", "With --oidc: audience to request, required on GitHub Actions (other platforms set it in the pipeline)")
	loginCmd.Flags().StringVar(&oidcTokenFile, "oidc-token-file", "", "With --oidc: file containing the OIDC token")
	loginCmd.Flags().BoolVar(&terraformCredentials, "terraform-credentials", false, "Also write the token to the Terraform CLI credentials file")
	loginCmd.Flags().StringSliceVar(&terraformHosts, "terraform-host", nil, "With --terraform-credentials: host(s) to write credentials for (default: the API URL host)")
	loginCmd.MarkFlagsMutuallyExclusive("web", "pat", "oidc")
	// A --web token may be renewed later; the copy in credentials.tfrc.json would not be.
	loginCmd.MarkFlagsMutuallyExclusive("web", "terraform-credentials")
}

func login(out io.Writer) error {
	url := loginAPIURL()
	pat := patToken
	if pat == "" {
		pat = os.Getenv("TERRAKUBE_PAT")
	}
	if url == "" {
		return errors.New("--api-url or TERRAKUBE_API_URL is required")
	}
	if pat == "" {
		return errors.New("--pat or TERRAKUBE_PAT is required")
	}
	return saveLogin(out, url, pat)
}

// loginOIDC logs in with the CI/CD workload identity token.
func loginOIDC(out io.Writer) error {
	url := loginAPIURL()
	if url == "" {
		return errors.New("--api-url or TERRAKUBE_API_URL is required")
	}
	// CI tokens may also be accepted by other relying parties (e.g. Azure AD), so never send them in clear text.
	if u, err := neturl.Parse(url); err != nil || (u.Scheme != "https" && !isLoopback(u.Hostname())) {
		return fmt.Errorf("--oidc requires an https API URL, got %q", url)
	}
	token, err := resolveOIDCToken(oidcTokenFile, oidcAudience)
	if err != nil {
		return err
	}
	return saveLogin(out, url, token)
}

func loginAPIURL() string {
	if apiURL != "" {
		return trimAPIURL(apiURL)
	}
	return trimAPIURL(os.Getenv("TERRAKUBE_API_URL"))
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

// saveLogin checks token against the server and saves it, and with
// --terraform-credentials also writes it to credentials.tfrc.json.
func saveLogin(out io.Writer, url, token string) error {
	// Check the hosts first, so a typo fails before anything is saved.
	var tfHosts []string
	if terraformCredentials {
		hosts := terraformHosts
		if len(hosts) == 0 {
			hosts = []string{url}
		}
		for _, h := range hosts {
			key, err := terraformHost(h)
			if err != nil {
				return err
			}
			tfHosts = append(tfHosts, key)
		}
	}

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(url),
		terrakube.WithToken(token),
	)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Test the connection by listing organizations
	if _, err := client.Organizations.List(getContext(), nil); err != nil {
		return fmt.Errorf("connecting to Terrakube server: %w", err)
	}

	if err := updateConfig(func(settings map[string]any) {
		settings["api_url"] = url
		settings["token"] = token
		delete(settings, "pat") // written by older versions of login
		for _, k := range refreshKeys {
			delete(settings, k) // a PAT or CI token is never refreshed
		}
	}); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	if terraformCredentials {
		path, err := writeTerraformCredentials(tfHosts, token)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "Terraform credentials for %s written to %s\n", strings.Join(tfHosts, ", "), path)
	}

	_, _ = fmt.Fprintln(out, "Successfully logged in to Terrakube server")
	return nil
}

// trimAPIURL strips a trailing /api/v1 (with or without a trailing slash).
func trimAPIURL(url string) string {
	return strings.TrimRight(strings.TrimSuffix(strings.TrimRight(url, "/"), "/api/v1"), "/")
}

// updateConfig applies fn to the config file's settings and atomically
// writes it back with 0600 permissions, keeping keys fn does not touch.
func updateConfig(fn func(settings map[string]any)) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if t := configType(path); t != "yaml" && t != "yml" {
		return fmt.Errorf("config file %s must be YAML, not %s", path, t)
	}
	// Write through symlinks (e.g. dotfile managers) instead of replacing them.
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}

	settings := map[string]any{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		if settings == nil {
			settings = map[string]any{}
		}
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}

	fn(settings)

	data, err = yaml.Marshal(settings)
	if err != nil {
		return err
	}
	// CreateTemp uses mode 0600; rename replaces the old file in one step.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".terrakube-cli-*.yaml")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
