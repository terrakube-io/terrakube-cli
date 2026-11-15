package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
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
	stateCmd.AddCommand(newStatePullCmd(), newStateOutputsCmd(), newStateRollbackCmd(), newStateDiffCmd())
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

func newStateRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback <workspace> <history-id>",
		Short: "make an earlier state version the workspace's current state",
		Long: `Make the state of one history entry (see: history list) the workspace's current
state, as the UI's rollback does. On a terminal it asks you to type the
workspace name; without a terminal it needs --yes.`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes && !resource.IsTerminal() {
				return errors.New("state rollback needs confirmation: run it in a terminal or pass --yes")
			}
			s, err := newStateScope(cmd, args[0])
			if err != nil {
				return err
			}
			h, err := s.version(args[1])
			if err != nil {
				return err
			}
			// The server does not check that the version belongs to the workspace.
			if h.Workspace != nil && h.Workspace.ID != "" && h.Workspace.ID != s.wsID {
				return fmt.Errorf("state version %s belongs to workspace %s, not %s", h.ID, h.Workspace.ID, s.ws)
			}
			if !yes {
				if err := s.confirm(cmd, h); err != nil {
					return err
				}
			}
			if err := s.client.State.Rollback(s.ctx, s.orgID, s.wsID, h); err != nil {
				if terrakube.IsNotFound(err) {
					return fmt.Errorf("state version %s has no raw state on the server, so it cannot be restored: %w", h.ID, err)
				}
				return err
			}
			_, err = fmt.Fprintf(os.Stdout, "workspace %s rolled back to state version %s\n", s.ws, h.ID)
			return err
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	cmd.Flags().Bool("yes", false, "skip the confirmation prompt")
	return cmd
}

// confirm asks for the workspace name on the terminal and fails unless it is
// typed back exactly.
func (s *stateScope) confirm(cmd *cobra.Command, h *terrakube.History) error {
	name := s.ws
	if resource.IsUUID(name) {
		ws, err := s.client.Workspaces.Get(s.ctx, s.orgID, s.wsID)
		if err != nil {
			return err
		}
		name = ws.Name
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Roll back workspace %s to state version %s (serial %d)? Type the workspace name to confirm: ", name, h.ID, h.Serial)
	typed, _ := bufio.NewReader(resource.Stdin).ReadString('\n')
	if strings.TrimSpace(typed) != name {
		return fmt.Errorf("confirmation did not match %q: nothing changed", name)
	}
	return nil
}

// stateDiff is what changed from one state version to another: resource
// instances by address, and outputs by name. It holds no values, so nothing
// sensitive is printed.
type stateDiff struct {
	Resources diffSet `json:"resources" yaml:"resources"`
	Outputs   diffSet `json:"outputs" yaml:"outputs"`
}

type diffSet struct {
	Added   []string `json:"added" yaml:"added"`
	Removed []string `json:"removed" yaml:"removed"`
	Changed []string `json:"changed" yaml:"changed"`
}

func newStateDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <workspace> <history-id-a> <history-id-b>",
		Short: "list the resources and outputs added, removed or changed between two state versions",
		Long: `Compare two state versions (see: history list) of a workspace: resource instances
by address (a change is any change of their attributes) and outputs by name.
Values are never printed.

The default is text, one "+ added", "- removed" or "~ changed" line per address.
An explicit --output json or yaml prints {resources, outputs} with added, removed
and changed lists; table and tsv print one change, kind, address row each.`,
		Args:         cobra.ExactArgs(3),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newStateScope(cmd, args[0])
			if err != nil {
				return err
			}
			resA, outA, err := s.readState(args[1])
			if err != nil {
				return err
			}
			resB, outB, err := s.readState(args[2])
			if err != nil {
				return err
			}
			d := stateDiff{Resources: diffMaps(resA, resB), Outputs: diffMaps(outA, outB)}
			switch {
			case !cmd.Flags().Changed("output"):
				return writeDiffText(os.Stdout, d)
			case output == "table" || output == "tsv":
				return outputpkg.RenderRows(os.Stdout, []string{"change", "kind", "address"}, d.rows(), output)
			default:
				return outputpkg.Render(os.Stdout, d, output)
			}
		},
	}
	resource.AddScopeFlags(cmd, orgScope)
	return cmd
}

// readState returns the resource instances (address to attributes) and the
// outputs (name to output) of a state version's raw state.
func (s *stateScope) readState(version string) (resources, outputs map[string]any, err error) {
	body, err := s.open(version, terrakube.StateRaw)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close() //nolint:errcheck // read-only body
	var st struct {
		Outputs   map[string]any `json:"outputs"`
		Resources []struct {
			Module    string `json:"module"`
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Name      string `json:"name"`
			Instances []struct {
				IndexKey   any `json:"index_key"`
				Attributes any `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}
	dec := json.NewDecoder(body)
	dec.UseNumber()
	if err := dec.Decode(&st); err != nil {
		return nil, nil, fmt.Errorf("state version %s: %w", version, err)
	}
	resources = map[string]any{}
	for _, r := range st.Resources {
		base := r.Type + "." + r.Name
		if r.Mode == "data" {
			base = "data." + base
		}
		if r.Module != "" {
			base = r.Module + "." + base
		}
		for _, inst := range r.Instances {
			resources[base+indexSuffix(inst.IndexKey)] = inst.Attributes
		}
	}
	return resources, st.Outputs, nil
}

// indexSuffix is the address suffix of an instance key: [0] for count,
// ["key"] for for_each, nothing for a single instance.
func indexSuffix(key any) string {
	switch k := key.(type) {
	case nil:
		return ""
	case string:
		return "[" + strconv.Quote(k) + "]"
	default:
		return fmt.Sprintf("[%v]", k)
	}
}

// diffMaps returns the keys only in b (added), only in a (removed) and in both
// with different values (changed), each sorted.
func diffMaps(a, b map[string]any) diffSet {
	d := diffSet{Added: []string{}, Removed: []string{}, Changed: []string{}}
	for k, va := range a {
		vb, ok := b[k]
		switch {
		case !ok:
			d.Removed = append(d.Removed, k)
		case !reflect.DeepEqual(va, vb):
			d.Changed = append(d.Changed, k)
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			d.Added = append(d.Added, k)
		}
	}
	slices.Sort(d.Added)
	slices.Sort(d.Removed)
	slices.Sort(d.Changed)
	return d
}

// rows returns one change, kind, address row per difference.
func (d stateDiff) rows() [][]string {
	var rows [][]string
	for _, sec := range []struct {
		kind string
		set  diffSet
	}{{"resource", d.Resources}, {"output", d.Outputs}} {
		for _, ch := range []struct {
			name  string
			addrs []string
		}{{"added", sec.set.Added}, {"removed", sec.set.Removed}, {"changed", sec.set.Changed}} {
			for _, a := range ch.addrs {
				rows = append(rows, []string{ch.name, sec.kind, a})
			}
		}
	}
	return rows
}

func writeDiffText(w io.Writer, d stateDiff) error {
	var b strings.Builder
	for _, sec := range []struct {
		title string
		set   diffSet
	}{{"Resources", d.Resources}, {"Outputs", d.Outputs}} {
		fmt.Fprintf(&b, "%s: %d added, %d removed, %d changed\n", sec.title, len(sec.set.Added), len(sec.set.Removed), len(sec.set.Changed))
		for _, a := range sec.set.Added {
			fmt.Fprintf(&b, "  + %s\n", a)
		}
		for _, a := range sec.set.Removed {
			fmt.Fprintf(&b, "  - %s\n", a)
		}
		for _, a := range sec.set.Changed {
			fmt.Fprintf(&b, "  ~ %s\n", a)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
