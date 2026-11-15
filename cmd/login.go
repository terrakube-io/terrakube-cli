package cmd

import (
	"errors"
	"fmt"
	"io/fs"
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
or with --web through the browser.

--web reads login.v1 from the server's /.well-known/terraform.json, opens the
browser (the URL is also printed) and waits for the login on a local port from
the advertised range. It saves the token it gets: a Terrakube token from
the Terrakube login broker, or a Dex token.

The API URL should be provided in the following format:
  http://your-server:port
  https://your-server:port

Note: Do not include /api/v1 in the URL as it will be added automatically.
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
		return login()
	},
	Example: fmt.Sprintf(loginExamples, rootCmd.Use),
}

var (
	apiURL   string
	patToken string
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&apiURL, "api-url", "a", "", "API URL (or TERRAKUBE_API_URL)")
	loginCmd.Flags().StringVarP(&patToken, "pat", "t", "", "Personal Access Token (or TERRAKUBE_PAT)")
	loginCmd.Flags().Bool("web", false, "Log in through the browser instead of with a PAT")
	loginCmd.Flags().Duration("timeout", 5*time.Minute, "With --web: how long to wait for the browser login")
	loginCmd.MarkFlagsMutuallyExclusive("web", "pat")
}

func login() error {
	url := apiURL
	if url == "" {
		url = os.Getenv("TERRAKUBE_API_URL")
	}
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

	url = trimAPIURL(url)

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(url),
		terrakube.WithToken(pat),
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
		settings["token"] = pat
		delete(settings, "pat") // written by older versions of login
	}); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	fmt.Println("Successfully logged in to Terrakube server")
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
