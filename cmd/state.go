package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	stateCmd := &cobra.Command{
		Use:   "state pull|outputs|rollback|diff",
		Short: "read, compare and roll back a workspace's Terraform state",
	}
	stateCmd.AddCommand(newStatePullCmd())
	rootCmd.AddCommand(stateCmd)
}

// stateScope is the client, organization and workspace a state command works on.
type stateScope struct {
	ctx    context.Context
	client *terrakube.Client
	orgID  string
	wsID   string
	ws     string // the workspace as given: a name or an ID
}

// newStateScope resolves -o and the workspace argument, a name or an ID.
func newStateScope(cmd *cobra.Command, ws string) (*stateScope, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}
	ctx := getContext()
	ids, err := resource.ResolveScope(ctx, client, cmd, orgScope)
	if err != nil {
		return nil, err
	}
	wsID := ws
	if !resource.IsUUID(ws) {
		if wsID, err = workspaceResolver(ctx, client, ids, ws); err != nil {
			return nil, err
		}
	}
	return &stateScope{ctx: ctx, client: client, orgID: ids[0], wsID: wsID, ws: ws}, nil
}

// version returns the workspace's history entry id.
func (s *stateScope) version(id string) (*terrakube.History, error) {
	return s.client.History.Get(s.ctx, s.orgID, s.wsID, id)
}

// open returns the current raw state when version is "", else the given file
// of that history version. The caller closes it.
func (s *stateScope) open(version string, format terrakube.StateFormat) (io.ReadCloser, error) {
	if version == "" {
		body, err := s.client.State.Current(s.ctx, s.orgID, s.wsID)
		if errors.Is(err, terrakube.ErrStateNotFound) {
			return nil, fmt.Errorf("workspace %s has no state", s.ws)
		}
		return body, err
	}
	h, err := s.version(version)
	if err != nil {
		return nil, err
	}
	body, err := s.client.State.Download(s.ctx, h, format)
	switch {
	case errors.Is(err, terrakube.ErrForeignHost):
		return nil, fmt.Errorf("state version %s is stored at %s, outside the API endpoint, so the token is not sent there: %w", version, h.Output, err)
	case errors.Is(err, terrakube.ErrStateNotFound):
		return nil, fmt.Errorf("state version %s of workspace %s has no state file", version, s.ws)
	}
	return body, err
}

func newStatePullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull <workspace>",
		Short: "write a workspace's Terraform state, byte for byte, to stdout or a file",
		Long: `Write the workspace's current Terraform state, or with --version the state of one
history entry (see: history list), exactly as the server stores it.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newStateScope(cmd, args[0])
			if err != nil {
				return err
			}
			version, _ := cmd.Flags().GetString("version")
			body, err := s.open(version, terrakube.StateRaw)
			if err != nil {
				return err
			}
			defer body.Close() //nolint:errcheck // read-only body
			if file, _ := cmd.Flags().GetString("file"); file != "" {
				return writeStateFile(file, body)
			}
			_, err = io.Copy(os.Stdout, body)
			return err
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	cmd.Flags().String("version", "", "history entry ID of the state version to write (default: the current state)")
	cmd.Flags().StringP("file", "f", "", "write to this file (mode 0600) instead of stdout")
	return cmd
}

// writeStateFile copies body to path, created with mode 0600 as state holds
// secrets. A failed copy removes the partial file.
func writeStateFile(path string, body io.Reader) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:gosec // the user names the file
	if err != nil {
		return err
	}
	_, err = io.Copy(f, body)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
	}
	return err
}
