package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"
)

const loginLong = `
Authenticate to a remote Terrakube server using a Personal Access Token (PAT)
or, inside CI/CD, a workload identity (OIDC) token.

The API URL should be provided in the following format:
  http://your-server:port
  https://your-server:port

Note: Do not include /api/v1 in the URL as it will be added automatically.

Workload identity (--oidc) sends the CI provider's OIDC token to Terrakube as-is.
The server accepts it when a federated credential matches its issuer, audience
and claims (see "terrakube federated" and "terrakube federated-claim").
The token is taken from the first available source:
  1. --oidc-token-file (e.g. a Kubernetes projected service account token)
  2. TERRAKUBE_OIDC_TOKEN environment variable
  3. The CI platform the CLI detects it is running on:
     GitHub Actions       requires "permissions: id-token: write" and --oidc-audience
     GitLab CI            id_tokens: { TERRAKUBE_OIDC_TOKEN: { aud: <audience> } }
     Bitbucket Pipelines  "oidc: true" on the step (BITBUCKET_STEP_OIDC_TOKEN)
     Azure Pipelines      workload identity service connection; map SYSTEM_ACCESSTOKEN
                          and AZURESUBSCRIPTION_SERVICE_CONNECTION_ID into the step env

OIDC tokens are short-lived: run login right before the commands that need it.
The API URL must be https. The token is stored on disk, so clean up HOME on
persistent self-hosted runners. On Azure Pipelines the token's audience is fixed
to Azure AD: use a service connection with no Azure role assignments.

--terraform-credentials also stores the token in the Terraform/OpenTofu CLI
credentials file (credentials.tfrc.json) so "terraform init" can use Terrakube
as a remote backend or private registry without extra setup.
`

var loginExamples = `
Login to a local Terrakube server using PAT
  %[1]v login -a http://localhost:8080 -t your-pat-token

Login to a remote Terrakube server
  %[1]v login --api-url https://terrakube.example.com --pat your-pat-token

Login from GitHub Actions and generate Terraform credentials
  %[1]v login -a https://terrakube.example.com --oidc --oidc-audience terrakube --terraform-credentials

Login from GitLab CI, Bitbucket Pipelines or Azure Pipelines (token source is detected)
  %[1]v login -a https://terrakube.example.com --oidc --terraform-credentials

Also write Terraform credentials for a separate registry host
  %[1]v login -a https://terrakube-api.example.com --oidc --terraform-credentials \
    --terraform-host terrakube-api.example.com --terraform-host terrakube-reg.example.com
`

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to Terrakube server",
	Long:  loginLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true // flags are valid; usage would only bury the runtime error in CI logs
		return login(cmd)
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
	f := loginCmd.Flags()
	f.StringVarP(&apiURL, "api-url", "a", "", "API URL (required)")
	_ = loginCmd.MarkFlagRequired("api-url")
	_ = viper.BindEnv("api-url", "TERRAKUBE_API_URL")
	f.StringVarP(&patToken, "pat", "t", "", "Personal Access Token")
	_ = viper.BindEnv("pat", "TERRAKUBE_PAT")
	f.BoolVar(&oidcLogin, "oidc", false, "Login with a CI/CD workload identity (OIDC) token instead of a PAT")
	f.StringVar(&oidcAudience, "oidc-audience", "", "Audience to request, required on GitHub Actions (other platforms set it in pipeline YAML)")
	f.StringVar(&oidcTokenFile, "oidc-token-file", "", "File containing the OIDC token")
	f.BoolVar(&terraformCredentials, "terraform-credentials", false, "Write the token to the Terraform CLI credentials file")
	f.StringSliceVar(&terraformHosts, "terraform-host", nil, "Host(s) to write Terraform credentials for (default: API URL host)")
	loginCmd.MarkFlagsOneRequired("pat", "oidc")
	loginCmd.MarkFlagsMutuallyExclusive("pat", "oidc")
}

func login(cmd *cobra.Command) error {
	apiURL = strings.TrimSuffix(strings.TrimSuffix(apiURL, "/"), "/api/v1")

	token := patToken
	if oidcLogin {
		// CI tokens may also be accepted by other relying parties (e.g. Azure AD), so never send them in clear text.
		if u, err := url.Parse(apiURL); err != nil || (u.Scheme != "https" && !isLoopback(u.Hostname())) {
			return fmt.Errorf("--oidc requires an https API URL, got %q", apiURL)
		}
		var err error
		if token, err = resolveOIDCToken(oidcTokenFile, oidcAudience); err != nil {
			return err
		}
	}

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(apiURL),
		terrakube.WithToken(token),
	)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Test the connection (and the token) by listing organizations
	if _, err = client.Organizations.List(getContext(), nil); err != nil {
		return fmt.Errorf("connecting to Terrakube server: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// A fresh viper keeps the file's existing keys but, unlike the global one, does not
	// dump every bound flag into it (a saved "oidc: false" would later be preset as a flag).
	configFile := filepath.Join(home, ".terrakube-cli.yaml")
	cfg := viper.New()
	cfg.SetConfigFile(configFile)
	cfg.SetConfigPermissions(0o600) // new file; the Chmod below covers an existing one
	if err = cfg.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading configuration: %w", err)
	}
	cfg.Set("api_url", apiURL)
	cfg.Set("token", token)
	if err = cfg.WriteConfig(); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	viper.Set("api_url", apiURL)
	viper.Set("token", token)
	if err = os.Chmod(configFile, 0o600); err != nil {
		return fmt.Errorf("restricting configuration permissions: %w", err)
	}

	out := cmd.OutOrStdout()
	if terraformCredentials {
		hosts := []string{apiURL}
		if len(terraformHosts) > 0 {
			hosts = terraformHosts
		}
		for i, h := range hosts {
			if hosts[i], err = terraformHost(h); err != nil {
				return err
			}
		}
		path, err := writeTerraformCredentials(hosts, token)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "Terraform credentials for %s written to %s\n", strings.Join(hosts, ", "), path)
	}
	_, _ = fmt.Fprintln(out, "Successfully logged in to Terrakube server")
	return nil
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}
