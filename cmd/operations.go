package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
)

var operationsCmd = &cobra.Command{
	Use:     "operations",
	Short:   "submit JSON:API atomic operations batch requests",
	Aliases: []string{"ops"},
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()

		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			return fmt.Errorf("flag --file is required")
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading operations file: %w", err)
		}

		var req terrakube.AtomicRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("parsing operations JSON: %w", err)
		}

		resp, err := client.Operations.Submit(ctx, &req)
		if err != nil {
			return err
		}

		renderOutput(resp, output)
		return nil
	},
}

func init() {
	operationsCmd.Flags().StringP("file", "f", "", "Path to JSON file containing atomic operations")
	_ = operationsCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(operationsCmd)
}
