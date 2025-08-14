package cmd

import (
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
	headers := make([]string, 0)
	headers = append(headers, "ID")
	result := make([][]string, 0)
	if reflectData.Kind() == reflect.Slice {
		for i := 0; i < reflectData.Len(); i++ {
			data := reflectData.Index(i).Interface()
			d := reflect.Indirect(reflect.ValueOf(data))

			row := make([]string, 0)
			id := d.FieldByName("ID").String()
			row = append(row, id)

			attr := reflect.Indirect(reflect.ValueOf(d.FieldByName("Attributes").Interface()))
			for j := 0; j < attr.NumField(); j++ {
				if i == 0 {
					headers = append(headers, attr.Type().Field(j).Name)
				}
				fieldValue := attr.Field(j)
				var valueStr string
				switch fieldValue.Kind() {
				case reflect.Bool:
					valueStr = fmt.Sprintf("%t", fieldValue.Bool())
				case reflect.Ptr:
					if fieldValue.IsNil() {
						valueStr = ""
					} else {
						// Dereference the pointer and get its value
						derefValue := fieldValue.Elem()
						switch derefValue.Kind() {
						case reflect.String:
							valueStr = derefValue.String()
						case reflect.Bool:
							valueStr = fmt.Sprintf("%t", derefValue.Bool())
						default:
							valueStr = fmt.Sprintf("%v", derefValue.Interface())
						}
					}
				default:
					valueStr = fieldValue.String()
				}

				row = append(row, valueStr)
			}
			result = append(result, row)
		}
	} else {
		d := reflect.Indirect(reflectData)
		row := make([]string, 0)
		id := d.FieldByName("ID").String()
		row = append(row, id)

		attr := reflect.Indirect(reflect.ValueOf(d.FieldByName("Attributes").Interface()))
		for j := 0; j < attr.NumField(); j++ {
			headers = append(headers, attr.Type().Field(j).Name)
			fieldValue := attr.Field(j)
			var valueStr string
			switch fieldValue.Kind() {
			case reflect.Bool:
				valueStr = fmt.Sprintf("%t", fieldValue.Bool())
			case reflect.Ptr:
				if fieldValue.IsNil() {
					valueStr = ""
				} else {
					// Dereference the pointer and get its value
					derefValue := fieldValue.Elem()
					switch derefValue.Kind() {
					case reflect.String:
						valueStr = derefValue.String()

					case reflect.Bool:
						valueStr = fmt.Sprintf("%t", derefValue.Bool())
					default:
						valueStr = fmt.Sprintf("%v", derefValue.Interface())
					}
				}
			default:
				valueStr = fieldValue.String()
			}
			row = append(row, valueStr)
		}
		result = append(result, row)

	}
	return result, headers
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
