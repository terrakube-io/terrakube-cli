package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

var (
	cfgFile   string
	output    string
	hideNulls bool
	verbose   bool
	configErr error // config file read/parse error from initConfig
)

const envPrefix = "TERRAKUBE"

var envReplacer = strings.NewReplacer("-", "_")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terrakube",
	Short: "terrakube command line tool",
	Long: `
terrakube is a CLI to handle remote terraform workspace and modules in organizations
and handle all the lifecycle (plan, apply, destroy).`,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := applyEnv(cmd); err != nil {
			return err
		}
		switch output {
		case "json", "yaml", "table", "tsv", "none":
		default:
			return fmt.Errorf("invalid --output %q: must be one of json, yaml, table, tsv, none", output)
		}
		outputpkg.HideNulls = hideNulls
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = FormatVersion()
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrakube-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "json", "Output format: json, yaml, table, tsv, or none")
	rootCmd.PersistentFlags().BoolVar(&hideNulls, "hide-nulls", true, "Hide null values in JSON output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	cobra.AddTemplateFunc("StyleHeading", color.New(color.FgCyan).SprintFunc())
	usageTemplate := rootCmd.UsageTemplate()
	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Aliases:`, `{{StyleHeading "Aliases:"}}`,
		`Available Commands:`, `{{StyleHeading "Commands:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
		`Examples:`, `{{StyleHeading "Examples:"}}`,
	).Replace(usageTemplate)
	re := regexp.MustCompile(`(?m)^Flags:\s*$`)
	usageTemplate = re.ReplaceAllLiteralString(usageTemplate, `{{StyleHeading "Flags:"}}`)
	rootCmd.SetUsageTemplate(usageTemplate)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	path, err := configPath()
	cobra.CheckErr(err)
	viper.SetConfigFile(path)
	viper.SetConfigType(configType(path))

	// Only api_url and token are read through viper (see newClient).
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()

	// A missing config file is fine; a broken one is reported by newClient.
	configErr = nil
	if err := viper.ReadInConfig(); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			configErr = fmt.Errorf("reading config file %s: %w", path, err)
		}
	} else if verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// configPath returns the --config value or the default $HOME/.terrakube-cli.yaml.
func configPath() (string, error) {
	if cfgFile != "" {
		return cfgFile, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".terrakube-cli.yaml"), nil
}

// configType returns the viper config type for path. Extensionless paths and
// extensions viper does not support are treated as YAML.
func configType(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if !slices.Contains(viper.SupportedExts, ext) {
		return "yaml"
	}
	return ext
}

func newClient() (*terrakube.Client, error) {
	if configErr != nil {
		return nil, configErr
	}
	endpoint := viper.GetString("api_url")
	// Env beats the config file: TERRAKUBE_TOKEN, then TERRAKUBE_PAT, then config.
	token := os.Getenv("TERRAKUBE_TOKEN")
	if token == "" {
		token = os.Getenv("TERRAKUBE_PAT")
	}
	if token == "" {
		token = viper.GetString("token")
	}
	if endpoint == "" || token == "" {
		return nil, errors.New(`not logged in: run "terrakube login" or set TERRAKUBE_API_URL and TERRAKUBE_TOKEN (or TERRAKUBE_PAT)`)
	}
	return terrakube.NewClient(
		terrakube.WithEndpoint(endpoint),
		terrakube.WithToken(token),
	)
}

func getContext() context.Context {
	return context.Background()
}

func renderOutput(result interface{}, format string) {
	if err := outputpkg.Render(os.Stdout, result, format); err != nil {
		log.Fatal("Failed to render output: ", err)
	}
}

// applyEnv fills flags of the executing command from TERRAKUBE_* environment
// variables when they were not set explicitly. Only global flags and parent
// scope flags (annotated by the resource framework) are eligible.
func applyEnv(cmd *cobra.Command) error {
	var err error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed || err != nil {
			return
		}
		names := f.Annotations[resource.EnvAnnotation]
		if f.Name == "output" || f.Name == "hide-nulls" {
			names = []string{f.Name}
		}
		for _, name := range names {
			key := envPrefix + "_" + strings.ToUpper(envReplacer.Replace(name))
			if val := os.Getenv(key); val != "" {
				if setErr := cmd.Flags().Set(f.Name, val); setErr != nil {
					err = fmt.Errorf("invalid %s: %w", key, setErr)
				}
				return
			}
		}
	})
	return err
}
