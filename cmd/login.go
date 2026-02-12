package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	terrakube "github.com/terrakube-io/terrakube-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const loginLong = `
Authenticate to a remote Terrakube server using a Personal Access Token (PAT).

The API URL should be provided in the following format:
  http://your-server:port
  https://your-server:port

Note: Do not include /api/v1 in the URL as it will be added automatically.
`

var loginExamples = `
Login to a local Terrakube server using PAT
  %v login -a http://localhost:8080 -t your-pat-token

Login to a remote Terrakube server
  %v login --api-url https://terrakube.example.com --pat your-pat-token
`

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to Terrakube server",
	Long:  loginLong,
	Run: func(cmd *cobra.Command, args []string) {
		login()
	},
	Example: fmt.Sprintf(loginExamples, rootCmd.Use),
}

var apiURL string
var patToken string

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&apiURL, "api-url", "a", "", "API URL (required)")
	_ = loginCmd.MarkFlagRequired("api-url")
	_ = viper.BindEnv("api-url", "TERRAKUBE_API_URL")
	loginCmd.Flags().StringVarP(&patToken, "pat", "t", "", "Personal Access Token (required)")
	_ = loginCmd.MarkFlagRequired("pat")
	_ = viper.BindEnv("pat", "TERRAKUBE_PAT")
}

func login() {
	// Strip /api/v1 from the URL if present
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1/")

	client, err := terrakube.NewClient(
		terrakube.WithEndpoint(apiURL),
		terrakube.WithToken(patToken),
	)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	// Test the connection by listing organizations
	ctx := getContext()
	_, err = client.Organizations.List(ctx, nil)
	if err != nil {
		fmt.Printf("Error connecting to Terrakube server: %v\n", err)
		return
	}

	// Save configuration
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	configFile := filepath.Join(home, ".terrakube-cli.yaml")
	viper.SetConfigFile(configFile)
	viper.Set("api_url", apiURL)
	viper.Set("token", patToken)

	err = viper.WriteConfig()
	if err != nil {
		// If the config file doesn't exist, try to create it
		if os.IsNotExist(err) {
			err = viper.SafeWriteConfig()
		}
		if err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}
	}

	fmt.Println("Successfully logged in to Terrakube server")
}
