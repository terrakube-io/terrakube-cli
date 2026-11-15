package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"
	"gopkg.in/yaml.v3"

	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

func init() {
	stateCmd := &cobra.Command{
		Use:   "state pull|outputs|rollback|diff",
		Short: "read, compare and roll back a workspace's Terraform state",
	}
	stateCmd.AddCommand(newStatePullCmd(), newStateOutputsCmd())
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

// sensitiveMask replaces sensitive output values unless --show-sensitive.
const sensitiveMask = "(sensitive)"

// stateOutput is one output in the json/yaml form of state outputs, keyed by
// name like "terraform output -json".
type stateOutput struct {
	Sensitive bool `json:"sensitive" yaml:"sensitive"`
	Type      any  `json:"type" yaml:"type"`
	Value     any  `json:"value" yaml:"value"`
}

func newStateOutputsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outputs <workspace>",
		Short: "print a workspace's Terraform outputs",
		Long: `Print the outputs of the workspace's newest state. json and yaml print a map of
name to {sensitive, type, value}, like "terraform output -json"; table and tsv
print one line per output. Sensitive values print as (sensitive) unless
--show-sensitive is set.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newStateScope(cmd, args[0])
			if err != nil {
				return err
			}
			outs, err := s.client.Outputs.List(s.ctx, s.orgID, s.wsID)
			if err != nil {
				return err
			}
			show, _ := cmd.Flags().GetBool("show-sensitive")
			byName := make(map[string]stateOutput, len(outs))
			rows := make([][]string, 0, len(outs)) // Outputs.List sorts by name
			for _, o := range outs {
				v := o.Value
				if o.Sensitive && !show {
					v = sensitiveMask
				}
				if output == "yaml" {
					v = yamlNumbers(v)
				}
				byName[o.Name] = stateOutput{Sensitive: o.Sensitive, Type: o.Type, Value: v}
				rows = append(rows, []string{o.Name, outputText(v), strconv.FormatBool(o.Sensitive)})
			}
			if output == "table" || output == "tsv" {
				return outputpkg.RenderRows(os.Stdout, []string{"name", "value", "sensitive"}, rows, output)
			}
			return outputpkg.Render(os.Stdout, byName, output)
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	cmd.Flags().Bool("show-sensitive", false, "print sensitive values instead of (sensitive)")
	return cmd
}

// outputText is v as one line: a string as is, anything else as compact JSON.
func outputText(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v) // v was decoded from JSON, so it encodes again
	return string(b)
}

// yamlNumbers returns v with its json.Number values as YAML number nodes:
// yaml would print a json.Number as a quoted string.
func yamlNumbers(v any) any {
	switch x := v.(type) {
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(string(x), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: string(x)}
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = yamlNumbers(e)
		}
		return m
	case []any:
		l := make([]any, len(x))
		for i, e := range x {
			l[i] = yamlNumbers(e)
		}
		return l
	}
	return v
}
