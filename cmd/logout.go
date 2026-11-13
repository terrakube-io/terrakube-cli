package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

const logoutLong = `
Remove credentials stored for a remote Terrakube server.
`

var logoutCmd = &cobra.Command{
	Use:          "logout",
	Short:        "logout from Terrakube server",
	Long:         logoutLong,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := configPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			fmt.Println("No stored credentials")
			return nil
		}
		if err := updateConfig(func(settings map[string]any) {
			delete(settings, "token")
			delete(settings, "pat") // written by older versions of login
		}); err != nil {
			return fmt.Errorf("saving configuration: %w", err)
		}
		fmt.Println("Successfully logged out")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
