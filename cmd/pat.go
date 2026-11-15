package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
)

func init() {
	patCmd := &cobra.Command{
		Use:   "pat create|list|delete",
		Short: "manage your personal access tokens",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "create a personal access token and print it once",
		Long: `Create a personal access token. The token is printed once and cannot be read again.
--days 0 asks for a token that does not expire; the server may refuse it.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			days, _ := cmd.Flags().GetInt("days")
			if days < 0 {
				return errors.New("--days must not be negative")
			}
			description, _ := cmd.Flags().GetString("description")
			client, err := newClient()
			if err != nil {
				return err
			}
			created, err := client.PATs.Create(getContext(), days, description)
			if terrakube.IsForbidden(err) {
				return errors.New("the server refused to create a token (403): only a user logged in through the browser (terrakube login --web) may create one")
			}
			if err != nil {
				return err
			}
			// The SDK keeps ID out of JSON; print it, it is what pat delete takes.
			return outputpkg.Render(os.Stdout, struct {
				ID    string `json:"id"`
				Token string `json:"token"`
			}{created.ID, created.Token}, output)
		},
	}
	create.Flags().Int("days", 30, "days until the token expires (0 = no expiry, if the server allows it)")
	create.Flags().String("description", "", "what the token is for")
	_ = create.MarkFlagRequired("description")

	list := &cobra.Command{
		Use:          "list",
		Short:        "list your personal access tokens",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(*cobra.Command, []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			pats, err := client.PATs.List(getContext())
			if err != nil {
				return err
			}
			return outputpkg.Render(os.Stdout, pats, output)
		},
	}

	del := &cobra.Command{
		Use:          "delete",
		Short:        "delete (revoke) a personal access token",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			id, _ := cmd.Flags().GetString("id")
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.PATs.Delete(getContext(), id); err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, "pat deleted")
			return err
		},
	}
	del.Flags().String("id", "", "token ID (see: pat list)")
	_ = del.MarkFlagRequired("id")

	patCmd.AddCommand(create, list, del)
	rootCmd.AddCommand(patCmd)
}
