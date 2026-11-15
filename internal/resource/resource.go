package resource

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/output"
)

// FieldType represents the type of a CLI flag.
type FieldType int

const (
	String FieldType = iota
	Bool
	Int
	StringSlice
	// Relationship is a to-one JSON:API relationship set by ID or name; "" clears it.
	Relationship
	// StringMap is a map attribute set by repeated KEY=VALUE flags (job --variables).
	StringMap
	// RelationshipMany is a to-many relationship set by IDs or names (comma-separated
	// or repeated); the list replaces the whole set, and "" empties it.
	RelationshipMany
)

// ParentScope defines a parent resource for name-to-ID resolution.
type ParentScope struct {
	Name      string   // e.g. "organization"
	Flag      string   // unified flag name, e.g. "organization" (accepts UUID or name)
	ShortFlag string   // short flag, e.g. "o"
	Aliases   []string // flag aliases, e.g. ["org"]
	IDFlag    string   // hidden backwards-compat alias, e.g. "organization-id"
	RawID     bool     // parent IDs are not UUIDs (e.g. numeric job IDs); use the flag value as the ID directly, no name resolution
	Optional  bool     // parent is optional (e.g. workspace on notification-configuration)
	Resolver  func(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error)
}

// FieldDef maps a CLI flag to a struct field.
type FieldDef struct {
	StructField string
	Flag        string
	Short       string
	Type        FieldType
	Required    bool
	Description string
	Hidden      bool // left out of --help: only executor or super-service tokens may set it
	// FileFlag, for a String field, adds --<FileFlag> <path> that reads the value
	// from a file instead (template --content-file). The two flags exclude each
	// other, and a Required field needs one of them.
	FileFlag string

	// For Relationship fields: StructField names the relation field (its jsonapi
	// tag gives the relationship name), RelType is the JSON:API type of the target,
	// and Resolver turns a name into an ID (nil accepts UUIDs only). A String field
	// with a Resolver holds an ID in an attribute; a name given to it is resolved.
	RelType  string
	Resolver func(ctx context.Context, c *terrakube.Client, parentIDs []string, name string) (string, error)
}

// EnvAnnotation is the flag annotation listing the flag names whose
// TERRAKUBE_* environment variables may fill a parent flag when it is not set.
const EnvAnnotation = "terrakube_env"

// Runtime provides access to CLI infrastructure.
type Runtime struct {
	NewClient  func() (*terrakube.Client, error)
	GetContext func() context.Context
	GetOutput  func() string
}

// Config defines a resource for CLI command generation.
type Config[T any] struct {
	Runtime
	Name    string
	Aliases []string
	Parents []ParentScope
	Fields  []FieldDef

	List func(ctx context.Context, c *terrakube.Client, parentIDs []string, opts *terrakube.ListOptions) ([]*T, error)
	Get  func(ctx context.Context, c *terrakube.Client, parentIDs []string, id string) (*T, error)
	// Create and Update receive only the attributes (and relationships) the user set,
	// so server defaults apply to the rest and explicit false/0/"" values are kept.
	Create func(ctx context.Context, c *terrakube.Client, parentIDs []string, p terrakube.Patch) (*T, error)
	Update func(ctx context.Context, c *terrakube.Client, parentIDs []string, id string, p terrakube.Patch) (*T, error)
	Delete func(ctx context.Context, c *terrakube.Client, parentIDs []string, id string) error

	// SoftDelete, when set, makes delete set an attribute through Update instead of
	// sending DELETE (the server's hooks only clean up on the attribute).
	SoftDelete *SoftDelete

	// HiddenVerbs lists subcommands (e.g. "delete") left out of --help because the
	// server allows them only to executor or super-service tokens. They still run.
	HiddenVerbs []string

	// CreateFlags adds extra flags to create; AfterCreate then gets the created
	// resource and replaces the default rendering (e.g. job create --wait).
	// ListFilters add list flags that filter on an attribute (job list --status);
	// they are ANDed with --filter.
	ListFilters []ListFilter

	CreateFlags func(cmd *cobra.Command)
	AfterCreate func(cmd *cobra.Command, ctx context.Context, c *terrakube.Client, parentIDs []string, created *T) error
}

// ListFilter is a list flag that adds Attribute=="value" to the RSQL filter.
type ListFilter struct {
	Flag        string
	Attribute   string
	Description string
}

// QuoteRSQL quotes v as an RSQL string, so spaces, ';', ',', quotes and
// parentheses cannot alter the filter.
func QuoteRSQL(v string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
}

// SoftDelete configures a delete that sets Attribute to true.
type SoftDelete struct {
	Attribute string // e.g. "deleted" or "disabled"
	Done      string // printed after the name, e.g. "marked as deleted"
	Hard      bool   // add --hard, which sends DELETE through Config.Delete
	Confirm   bool   // ask for the resource name on a TTY, or require --yes
}

// IsTerminal reports whether stdin is a terminal; Stdin is where the delete
// confirmation is read. Both are variables so tests can replace them.
var (
	IsTerminal           = func() bool { return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) }
	Stdin      io.Reader = os.Stdin
)

// Register creates Cobra commands for the resource and adds them to root.
func Register[T any](root *cobra.Command, cfg Config[T]) *cobra.Command {
	checkFields[T](cfg.Fields)
	parentCmd := &cobra.Command{
		Use:     cfg.Name + " list|get|create|update|delete [FLAGS]",
		Short:   fmt.Sprintf("manage %s resources", cfg.Name),
		Aliases: cfg.Aliases,
	}
	root.AddCommand(parentCmd)

	if cfg.List != nil {
		parentCmd.AddCommand(newListCmd(cfg))
	}
	if cfg.Get != nil {
		parentCmd.AddCommand(newGetCmd(cfg))
	}
	if cfg.Create != nil {
		parentCmd.AddCommand(newCreateCmd(cfg))
	}
	if cfg.Update != nil {
		parentCmd.AddCommand(newUpdateCmd(cfg))
	}
	if cfg.Delete != nil || cfg.SoftDelete != nil {
		parentCmd.AddCommand(newDeleteCmd(cfg))
	}
	for _, sub := range parentCmd.Commands() {
		if slices.Contains(cfg.HiddenVerbs, sub.Name()) {
			sub.Hidden = true
		}
	}
	return parentCmd
}

func newListCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        fmt.Sprintf("list %s resources", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cfg.NewClient()
			if err != nil {
				return err
			}
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			var opts *terrakube.ListOptions
			filter, _ := cmd.Flags().GetString("filter")
			parts := []string{}
			if filter != "" {
				parts = append(parts, filter)
			}
			for _, lf := range cfg.ListFilters {
				if v, _ := cmd.Flags().GetString(lf.Flag); v != "" {
					parts = append(parts, lf.Attribute+"=="+QuoteRSQL(v))
				}
			}
			if len(parts) > 0 {
				opts = &terrakube.ListOptions{Filter: strings.Join(parts, ";")}
			}

			result, err := cfg.List(ctx, client, parentIDs, opts)
			if err != nil {
				return err
			}

			return output.Render(os.Stdout, result, cfg.GetOutput())
		},
	}

	addParentFlags(cmd, cfg.Parents)
	cmd.Flags().String("filter", "", "RSQL filter expression")
	for _, lf := range cfg.ListFilters {
		cmd.Flags().String(lf.Flag, "", lf.Description)
	}
	return cmd
}

func newGetCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        fmt.Sprintf("get a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cfg.NewClient()
			if err != nil {
				return err
			}
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			id, _ := cmd.Flags().GetString("id")
			result, err := cfg.Get(ctx, client, parentIDs, id)
			if err != nil {
				return err
			}

			return output.Render(os.Stdout, result, cfg.GetOutput())
		},
	}

	addParentFlags(cmd, cfg.Parents)
	cmd.Flags().String("id", "", fmt.Sprintf("%s ID", cfg.Name))
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newCreateCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        fmt.Sprintf("create a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cfg.NewClient()
			if err != nil {
				return err
			}
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			resource := new(T)
			if err := populateChangedFields(cmd, cfg.Fields, resource); err != nil {
				return err
			}

			patch := changedPatch(cmd, cfg.Fields, resource)
			if err := addRelationships[T](ctx, client, cmd, cfg.Fields, parentIDs, &patch); err != nil {
				return err
			}
			result, err := cfg.Create(ctx, client, parentIDs, patch)
			if err != nil {
				return err
			}
			if cfg.AfterCreate != nil {
				return cfg.AfterCreate(cmd, ctx, client, parentIDs, result)
			}

			return output.Render(os.Stdout, result, cfg.GetOutput())
		},
	}

	addParentFlags(cmd, cfg.Parents)
	addFieldFlags(cmd, cfg.Fields, true)
	if cfg.CreateFlags != nil {
		cfg.CreateFlags(cmd)
	}
	return cmd
}

func newUpdateCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update",
		Short:        fmt.Sprintf("update a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cfg.NewClient()
			if err != nil {
				return err
			}
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			id, _ := cmd.Flags().GetString("id")
			resource := new(T)
			setStructField(resource, "ID", id)
			if err := populateChangedFields(cmd, cfg.Fields, resource); err != nil {
				return err
			}

			patch := changedPatch(cmd, cfg.Fields, resource)
			if err := addRelationships[T](ctx, client, cmd, cfg.Fields, parentIDs, &patch); err != nil {
				return err
			}
			if len(patch.Attributes) == 0 && len(patch.Relationships) == 0 {
				return errors.New("nothing to update: set at least one field flag, see --help")
			}
			result, err := cfg.Update(ctx, client, parentIDs, id, patch)
			if err != nil {
				return err
			}
			// Elide may answer PATCH with 204, leaving result empty: show the stored
			// state, or the sent fields if it can't be read back (e.g. soft-deleted).
			if result == nil || reflect.ValueOf(result).Elem().FieldByName("ID").String() == "" {
				result = resource
				if cfg.Get != nil {
					if fetched, getErr := cfg.Get(ctx, client, parentIDs, id); getErr == nil {
						result = fetched
					}
				}
			}

			return output.Render(os.Stdout, result, cfg.GetOutput())
		},
	}

	addParentFlags(cmd, cfg.Parents)
	cmd.Flags().String("id", "", fmt.Sprintf("%s ID", cfg.Name))
	_ = cmd.MarkFlagRequired("id")
	addFieldFlags(cmd, cfg.Fields, false)
	return cmd
}

func newDeleteCmd[T any](cfg Config[T]) *cobra.Command {
	soft := cfg.SoftDelete
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        fmt.Sprintf("delete a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			confirm := false
			if soft != nil && soft.Confirm {
				yes, _ := cmd.Flags().GetBool("yes")
				if !yes && !IsTerminal() {
					return fmt.Errorf("%s delete needs confirmation: run it in a terminal or pass --yes", cfg.Name)
				}
				confirm = !yes
			}

			client, err := cfg.NewClient()
			if err != nil {
				return err
			}
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			id, _ := cmd.Flags().GetString("id")
			if confirm {
				if err := confirmByName(ctx, cmd, cfg, client, parentIDs, id); err != nil {
					return err
				}
			}

			hard, _ := cmd.Flags().GetBool("hard")
			if soft == nil || hard {
				if err := cfg.Delete(ctx, client, parentIDs, id); err != nil {
					return err
				}
				_, err = fmt.Fprintf(os.Stdout, "%s deleted\n", cfg.Name)
				return err
			}

			p := terrakube.Patch{Attributes: map[string]any{soft.Attribute: true}}
			if _, err := cfg.Update(ctx, client, parentIDs, id, p); err != nil {
				return err
			}
			_, err = fmt.Fprintf(os.Stdout, "%s %s\n", cfg.Name, soft.Done)
			return err
		},
	}

	addParentFlags(cmd, cfg.Parents)
	cmd.Flags().String("id", "", fmt.Sprintf("%s ID", cfg.Name))
	_ = cmd.MarkFlagRequired("id")
	if soft != nil && soft.Hard {
		cmd.Flags().Bool("hard", false, fmt.Sprintf("send DELETE instead of setting %s=true", soft.Attribute))
	}
	if soft != nil && soft.Confirm {
		cmd.Flags().Bool("yes", false, "skip the confirmation prompt")
	}
	return cmd
}

// confirmByName reads the resource's name from the terminal and fails unless it
// matches the stored name.
func confirmByName[T any](ctx context.Context, cmd *cobra.Command, cfg Config[T], client *terrakube.Client, parentIDs []string, id string) error {
	current, err := cfg.Get(ctx, client, parentIDs, id)
	if err != nil {
		return err
	}
	name := reflect.ValueOf(current).Elem().FieldByName("Name").String()
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Type the %s name (%s) to confirm: ", cfg.Name, name)
	typed, _ := bufio.NewReader(Stdin).ReadString('\n')
	if strings.TrimSpace(typed) != name {
		return fmt.Errorf("confirmation did not match %q: nothing changed", name)
	}
	return nil
}

var timeType = reflect.TypeFor[time.Time]()

// checkFields panics at startup if a FieldDef names a struct field that does
// not exist or cannot hold its flag type, e.g. after a terrakube-go upgrade.
func checkFields[T any](fields []FieldDef) {
	t := reflect.TypeFor[T]()
	for _, f := range fields {
		sf, ok := t.FieldByName(f.StructField)
		if !ok {
			panic(fmt.Sprintf("%s has no field %q (flag --%s)", t.Name(), f.StructField, f.Flag))
		}
		ft := sf.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		var fits bool
		switch f.Type {
		case String:
			fits = ft.Kind() == reflect.String || ft == timeType
		case Bool:
			fits = ft.Kind() == reflect.Bool
		case Int:
			fits = ft.Kind() >= reflect.Int && ft.Kind() <= reflect.Int64
		case StringSlice:
			fits = ft == reflect.TypeFor[[]string]()
		case Relationship, RelationshipMany:
			kind := reflect.Struct
			if f.Type == RelationshipMany {
				kind = reflect.Slice // of pointers to the target struct
			}
			tag := strings.Split(sf.Tag.Get("jsonapi"), ",")
			fits = ft.Kind() == kind && len(tag) > 1 && tag[0] == "relation" && f.RelType != ""
		case StringMap:
			fits = ft.Kind() == reflect.Map && ft.Key().Kind() == reflect.String
		}
		if f.FileFlag != "" && f.Type != String {
			fits = false
		}
		if !fits {
			panic(fmt.Sprintf("%s.%s is %s, cannot hold flag --%s", t.Name(), f.StructField, sf.Type, f.Flag))
		}
	}
}

func addParentFlags(cmd *cobra.Command, parents []ParentScope) {
	aliases := make(map[string]string)
	for _, p := range parents {
		usage := fmt.Sprintf("%s ID or name", p.Name)
		if p.RawID {
			usage = fmt.Sprintf("%s ID", p.Name)
		}
		cmd.Flags().StringP(p.Flag, p.ShortFlag, "", usage)
		envNames := []string{p.Flag}
		if p.IDFlag != "" && p.IDFlag != p.Flag {
			aliases[p.IDFlag] = p.Flag
			envNames = append(envNames, p.IDFlag)
		}
		_ = cmd.Flags().SetAnnotation(p.Flag, EnvAnnotation, envNames)
		for _, a := range p.Aliases {
			aliases[a] = p.Flag
		}
	}
	if len(aliases) > 0 {
		prev := cmd.Flags().GetNormalizeFunc()
		cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
			if mapped, ok := aliases[name]; ok {
				return pflag.NormalizedName(mapped)
			}
			return prev(f, name)
		})
	}
}

func addFieldFlags(cmd *cobra.Command, fields []FieldDef, forCreate bool) {
	for _, f := range fields {
		desc := f.Description
		if desc == "" {
			desc = f.Flag
		}

		switch f.Type {
		case String, Relationship:
			cmd.Flags().StringP(f.Flag, f.Short, "", desc)
		case Bool:
			cmd.Flags().BoolP(f.Flag, f.Short, false, desc)
		case Int:
			cmd.Flags().IntP(f.Flag, f.Short, 0, desc)
		case StringSlice, RelationshipMany:
			cmd.Flags().StringSliceP(f.Flag, f.Short, nil, desc)
		case StringMap:
			cmd.Flags().StringArrayP(f.Flag, f.Short, nil, desc+" (KEY=VALUE, repeatable)")
		}

		if f.FileFlag != "" {
			cmd.Flags().String(f.FileFlag, "", fmt.Sprintf("read --%s from this file", f.Flag))
			cmd.MarkFlagsMutuallyExclusive(f.Flag, f.FileFlag)
		}
		switch {
		case forCreate && f.Required && f.FileFlag != "":
			cmd.MarkFlagsOneRequired(f.Flag, f.FileFlag)
		case forCreate && f.Required:
			_ = cmd.MarkFlagRequired(f.Flag)
		}
		if f.Hidden {
			_ = cmd.Flags().MarkHidden(f.Flag)
		}
	}
}

// changed reports whether f's flag, or its FileFlag, was set on cmd.
func changed(cmd *cobra.Command, f FieldDef) bool {
	if f.FileFlag != "" && cmd.Flags().Changed(f.FileFlag) {
		return true
	}
	return cmd.Flags().Changed(f.Flag)
}

func populateChangedFields(cmd *cobra.Command, fields []FieldDef, obj any) error {
	for _, f := range fields {
		if !changed(cmd, f) {
			continue
		}
		if err := setFieldFromFlag(cmd, f, obj); err != nil {
			return err
		}
	}
	return nil
}

func setFieldFromFlag(cmd *cobra.Command, f FieldDef, obj any) error {
	v := reflect.ValueOf(obj).Elem()
	field := v.FieldByName(f.StructField)
	if !field.IsValid() {
		return fmt.Errorf("struct has no field %q", f.StructField)
	}

	switch f.Type {
	case Relationship, RelationshipMany:
		// Sent as a relationship by addRelationships, not stored in the struct.
	case String:
		val, _ := cmd.Flags().GetString(f.Flag)
		if f.FileFlag != "" && cmd.Flags().Changed(f.FileFlag) {
			path, _ := cmd.Flags().GetString(f.FileFlag)
			b, err := os.ReadFile(path) //nolint:gosec // the user names the file to send
			if err != nil {
				return fmt.Errorf("--%s: %w", f.FileFlag, err)
			}
			val = string(b)
		}
		if err := setStringField(field, val); err != nil {
			return fmt.Errorf("--%s: %w", f.Flag, err)
		}
	case Bool:
		val, _ := cmd.Flags().GetBool(f.Flag)
		setBoolField(field, val)
	case Int:
		val, _ := cmd.Flags().GetInt(f.Flag)
		setIntField(field, val)
	case StringSlice:
		val, _ := cmd.Flags().GetStringSlice(f.Flag)
		setStringSliceField(field, val)
	case StringMap:
		pairs, _ := cmd.Flags().GetStringArray(f.Flag)
		m := reflect.MakeMap(field.Type())
		for _, kv := range pairs {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("--%s: %q is not KEY=VALUE", f.Flag, kv)
			}
			m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v).Convert(field.Type().Elem()))
		}
		field.Set(m)
	}
	return nil
}

func setStringField(field reflect.Value, val string) error {
	if field.Kind() == reflect.Pointer {
		if val == "" {
			return nil
		}
		ptr := reflect.New(field.Type().Elem())
		if err := setStringField(ptr.Elem(), val); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}
	if field.Type() == timeType {
		if val == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return fmt.Errorf("want an RFC 3339 timestamp like 2026-01-02T15:04:05Z: %w", err)
		}
		// Same form jsonapi's rfc3339 tag sends on create, so update matches it.
		field.Set(reflect.ValueOf(t.UTC().Truncate(time.Second)))
		return nil
	}
	field.SetString(val)
	return nil
}

func setStringSliceField(field reflect.Value, val []string) {
	if len(val) == 0 {
		return
	}
	field.Set(reflect.ValueOf(val))
}

func setBoolField(field reflect.Value, val bool) {
	if field.Kind() == reflect.Pointer {
		ptr := reflect.New(field.Type().Elem())
		ptr.Elem().SetBool(val)
		field.Set(ptr)
	} else {
		field.SetBool(val)
	}
}

func setIntField(field reflect.Value, val int) {
	if field.Kind() == reflect.Pointer {
		ptr := reflect.New(field.Type().Elem())
		ptr.Elem().SetInt(int64(val))
		field.Set(ptr)
	} else {
		field.SetInt(int64(val))
	}
}

func setStructField(obj any, name string, val string) {
	v := reflect.ValueOf(obj).Elem()
	field := v.FieldByName(name)
	if field.IsValid() && field.CanSet() {
		field.SetString(val)
	}
}

// changedPatch returns the JSON:API attributes of the flags set on cmd, with
// the typed values already parsed into obj by populateChangedFields.
func changedPatch(cmd *cobra.Command, fields []FieldDef, obj any) terrakube.Patch {
	p := terrakube.Patch{Attributes: map[string]any{}}
	rv := reflect.ValueOf(obj).Elem()
	for _, f := range fields {
		sf, ok := rv.Type().FieldByName(f.StructField)
		if !changed(cmd, f) || !ok {
			continue
		}
		if tag := strings.Split(sf.Tag.Get("jsonapi"), ","); len(tag) > 1 && tag[0] == "attr" {
			p.Attributes[tag[1]] = rv.FieldByIndex(sf.Index).Interface()
		}
	}
	return p
}

// addRelationships adds the Relationship flags set on cmd to p: an ID as is, a
// name through the field's Resolver, and "" as a cleared link. It also resolves
// names given to String fields that have a Resolver (attributes holding an ID).
func addRelationships[T any](ctx context.Context, client *terrakube.Client, cmd *cobra.Command, fields []FieldDef, parentIDs []string, p *terrakube.Patch) error {
	for _, f := range fields {
		flag := cmd.Flags().Lookup(f.Flag)
		if flag == nil || !flag.Changed || f.Resolver == nil && f.Type != Relationship && f.Type != RelationshipMany {
			continue
		}
		sf, _ := reflect.TypeFor[T]().FieldByName(f.StructField) // checked by checkFields
		name := strings.Split(sf.Tag.Get("jsonapi"), ",")[1]
		if f.Type == String {
			// An attribute holding an ID (e.g. defaultTemplate): resolve a name to it.
			if val := flag.Value.String(); val != "" && !IsUUID(val) {
				id, err := f.Resolver(ctx, client, parentIDs, val)
				if err != nil {
					return err
				}
				p.Attributes[name] = id
			}
			continue
		}
		if p.Relationships == nil {
			p.Relationships = map[string]terrakube.Relationship{}
		}
		if f.Type == RelationshipMany {
			vals, _ := cmd.Flags().GetStringSlice(f.Flag)
			rel := terrakube.Relationship{Type: f.RelType, Many: true}
			for _, val := range vals {
				id, err := resolveRef(ctx, client, f, parentIDs, val)
				if err != nil {
					return err
				}
				rel.IDs = append(rel.IDs, id)
			}
			p.Relationships[name] = rel
			continue
		}
		val := flag.Value.String()
		if val == "" {
			p.Relationships[name] = terrakube.Relationship{}
			continue
		}
		id, err := resolveRef(ctx, client, f, parentIDs, val)
		if err != nil {
			return err
		}
		p.Relationships[name] = terrakube.Relationship{Type: f.RelType, ID: &id}
	}
	return nil
}

// resolveRef returns val if it is a UUID, else the ID f's Resolver finds for it.
func resolveRef(ctx context.Context, client *terrakube.Client, f FieldDef, parentIDs []string, val string) (string, error) {
	switch {
	case IsUUID(val):
		return val, nil
	case f.Resolver == nil:
		return "", fmt.Errorf("--%s: %q is not a UUID", f.Flag, val)
	default:
		return f.Resolver(ctx, client, parentIDs, val)
	}
}
