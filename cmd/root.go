package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spf13/viper"

	terrakube "github.com/denniswebb/terrakube-go"
	"terrakube/client/client"
)

var cfgFile string
var output string
var envPrefix string = "TERRAKUBE"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terrakube",
	Short: "terrakube command line tool",
	Long: `
terrakube is a CLI to handle remote terraform workspace and modules in organizations 
and handle all the lifecycle (plan, apply, destroy).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrakube-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "json", "Use json, table, tsv or none to format CLI output")
	_ = viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	_ = rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

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
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configFile := filepath.Join(home, ".terrakube-cli.yaml")
		viper.SetConfigFile(configFile)
	}

	viper.SetEnvPrefix(envPrefix)
	_ = viper.BindEnv("workspace-id", "TERRAKUBE_WORKSPACE_ID")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	postInitCommands(rootCmd.Commands())
}

func newClient() *client.Client {
	baseURL, err := url.Parse(viper.GetString("api_url"))
	if err != nil {
		fmt.Printf("Error parsing API URL: %v\n", err)
		os.Exit(1)
	}

	return client.NewClient(nil, viper.GetString("token"), baseURL)
}

//nolint:unused // used by resource commands during terrakube-go migration
func newTerrakubeClient() *terrakube.Client {
	c, err := terrakube.NewClient(
		terrakube.WithEndpoint(viper.GetString("api_url")),
		terrakube.WithToken(viper.GetString("token")),
	)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	return c
}

//nolint:unused // used by resource commands during terrakube-go migration
func getContext() context.Context {
	return context.Background()
}

//nolint:unused // used by resource commands during terrakube-go migration
func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func renderOutput(result interface{}, format string) {
	switch format {
	case "json":
		printJSON, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Printf("%s\n", string(printJSON))
	case "tsv":
		data, _ := splitInterface(result)
		for _, v := range data {
			fmt.Println(strings.Join(v[:], "\t"))
		}
	case "table":
		data, header := splitInterface(result)
		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.AppendBulk(data)
			table.SetHeader(header)
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCaption(true, " ")
			table.SetCenterSeparator("|")
			table.Render()
		}
	case "none":

	}
}

func splitInterface(input interface{}) ([][]string, []string) {
	reflectData := reflect.ValueOf(input)
	headers := []string{"ID"}
	result := make([][]string, 0)

	if reflectData.Kind() == reflect.Slice {
		for i := 0; i < reflectData.Len(); i++ {
			data := reflectData.Index(i).Interface()
			d := reflect.Indirect(reflect.ValueOf(data))
			row := []string{d.FieldByName("ID").String()}

			if isNestedModel(d) {
				row, headers = appendNestedFields(d, row, headers, i == 0)
			} else {
				row, headers = appendFlatFields(d, row, headers, i == 0)
			}
			result = append(result, row)
		}
	} else {
		d := reflect.Indirect(reflectData)
		row := []string{d.FieldByName("ID").String()}

		if isNestedModel(d) {
			row, headers = appendNestedFields(d, row, headers, true)
		} else {
			row, headers = appendFlatFields(d, row, headers, true)
		}
		result = append(result, row)
	}
	return result, headers
}

// isNestedModel returns true if the struct has a non-nil Attributes field
// (old client/models pattern).
func isNestedModel(d reflect.Value) bool {
	f := d.FieldByName("Attributes")
	return f.IsValid() && f.Kind() == reflect.Ptr && !f.IsNil()
}

// appendNestedFields extracts columns from the old nested Attributes sub-struct.
func appendNestedFields(d reflect.Value, row []string, headers []string, buildHeaders bool) ([]string, []string) {
	attr := reflect.Indirect(reflect.ValueOf(d.FieldByName("Attributes").Interface()))
	for j := 0; j < attr.NumField(); j++ {
		if buildHeaders {
			headers = append(headers, attr.Type().Field(j).Name)
		}
		row = append(row, formatFieldValue(attr.Field(j)))
	}
	return row, headers
}

// appendFlatFields extracts columns from a flat struct (terrakube-go pattern),
// skipping the ID field (already handled) and any jsonapi relation fields.
func appendFlatFields(d reflect.Value, row []string, headers []string, buildHeaders bool) ([]string, []string) {
	for j := 0; j < d.NumField(); j++ {
		field := d.Type().Field(j)
		if field.Name == "ID" || !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("jsonapi")
		if strings.HasPrefix(tag, "relation,") {
			continue
		}
		if buildHeaders {
			headers = append(headers, field.Name)
		}
		row = append(row, formatFieldValue(d.Field(j)))
	}
	return row, headers
}

func formatFieldValue(fieldValue reflect.Value) string {
	switch fieldValue.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%t", fieldValue.Bool())
	case reflect.Ptr:
		if fieldValue.IsNil() {
			return ""
		}
		derefValue := fieldValue.Elem()
		switch derefValue.Kind() {
		case reflect.String:
			return derefValue.String()
		case reflect.Bool:
			return fmt.Sprintf("%t", derefValue.Bool())
		default:
			return fmt.Sprintf("%v", derefValue.Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", fieldValue.Int())
	default:
		return fieldValue.String()
	}
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	_ = viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			_ = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
