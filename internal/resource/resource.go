package resource

import (
	"context"
	"fmt"
	"os"
	"reflect"

	terrakube "github.com/denniswebb/terrakube-go"
	"github.com/spf13/cobra"

	"terrakube/internal/output"
)

// FieldType represents the type of a CLI flag.
type FieldType int

const (
	String FieldType = iota
	Bool
	Int
)

// ParentScope defines a parent resource for name-to-ID resolution.
type ParentScope struct {
	Name     string // e.g. "organization"
	IDFlag   string // e.g. "organization-id"
	NameFlag string // e.g. "organization-name" (empty disables name resolution)
	Resolver func(ctx context.Context, c *terrakube.Client, resolvedParentIDs []string, name string) (string, error)
}

// FieldDef maps a CLI flag to a struct field.
type FieldDef struct {
	StructField string
	Flag        string
	Short       string
	Type        FieldType
	Required    bool
	Description string
}

// Runtime provides access to CLI infrastructure.
type Runtime struct {
	NewClient  func() *terrakube.Client
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

	List   func(ctx context.Context, c *terrakube.Client, parentIDs []string, opts *terrakube.ListOptions) ([]*T, error)
	Get    func(ctx context.Context, c *terrakube.Client, parentIDs []string, id string) (*T, error)
	Create func(ctx context.Context, c *terrakube.Client, parentIDs []string, resource *T) (*T, error)
	Update func(ctx context.Context, c *terrakube.Client, parentIDs []string, resource *T) (*T, error)
	Delete func(ctx context.Context, c *terrakube.Client, parentIDs []string, id string) error
}

// Register creates Cobra commands for the resource and adds them to root.
func Register[T any](root *cobra.Command, cfg Config[T]) {
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
	if cfg.Delete != nil {
		parentCmd.AddCommand(newDeleteCmd(cfg))
	}
}

func newListCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        fmt.Sprintf("list %s resources", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cfg.NewClient()
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			var opts *terrakube.ListOptions
			filter, _ := cmd.Flags().GetString("filter")
			if filter != "" {
				opts = &terrakube.ListOptions{Filter: filter}
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
	return cmd
}

func newGetCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        fmt.Sprintf("get a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cfg.NewClient()
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
			client := cfg.NewClient()
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			resource := new(T)
			if err := populateFields(cmd, cfg.Fields, resource); err != nil {
				return err
			}

			result, err := cfg.Create(ctx, client, parentIDs, resource)
			if err != nil {
				return err
			}

			return output.Render(os.Stdout, result, cfg.GetOutput())
		},
	}

	addParentFlags(cmd, cfg.Parents)
	addFieldFlags(cmd, cfg.Fields, true)
	return cmd
}

func newUpdateCmd[T any](cfg Config[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update",
		Short:        fmt.Sprintf("update a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cfg.NewClient()
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

			result, err := cfg.Update(ctx, client, parentIDs, resource)
			if err != nil {
				return err
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
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        fmt.Sprintf("delete a %s resource", cfg.Name),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cfg.NewClient()
			ctx := cfg.GetContext()

			parentIDs, err := resolveParents(ctx, client, cmd, cfg.Parents)
			if err != nil {
				return err
			}

			id, _ := cmd.Flags().GetString("id")
			if err := cfg.Delete(ctx, client, parentIDs, id); err != nil {
				return err
			}

			_, err = fmt.Fprintf(os.Stdout, "%s deleted\n", cfg.Name)
			return err
		},
	}

	addParentFlags(cmd, cfg.Parents)
	cmd.Flags().String("id", "", fmt.Sprintf("%s ID", cfg.Name))
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func addParentFlags(cmd *cobra.Command, parents []ParentScope) {
	for _, p := range parents {
		cmd.Flags().String(p.IDFlag, "", fmt.Sprintf("%s ID", p.Name))
		if p.NameFlag != "" {
			cmd.Flags().String(p.NameFlag, "", fmt.Sprintf("%s name (resolved to ID)", p.Name))
		}
	}
}

func addFieldFlags(cmd *cobra.Command, fields []FieldDef, forCreate bool) {
	for _, f := range fields {
		desc := f.Description
		if desc == "" {
			desc = f.Flag
		}

		switch f.Type {
		case String:
			cmd.Flags().StringP(f.Flag, f.Short, "", desc)
		case Bool:
			cmd.Flags().BoolP(f.Flag, f.Short, false, desc)
		case Int:
			cmd.Flags().IntP(f.Flag, f.Short, 0, desc)
		}

		if forCreate && f.Required {
			_ = cmd.MarkFlagRequired(f.Flag)
		}
	}
}

func populateFields(cmd *cobra.Command, fields []FieldDef, obj any) error {
	for _, f := range fields {
		if err := setFieldFromFlag(cmd, f, obj); err != nil {
			return err
		}
	}
	return nil
}

func populateChangedFields(cmd *cobra.Command, fields []FieldDef, obj any) error {
	for _, f := range fields {
		flag := cmd.Flags().Lookup(f.Flag)
		if flag == nil || !flag.Changed {
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
	case String:
		val, _ := cmd.Flags().GetString(f.Flag)
		setStringField(field, val)
	case Bool:
		val, _ := cmd.Flags().GetBool(f.Flag)
		setBoolField(field, val)
	case Int:
		val, _ := cmd.Flags().GetInt(f.Flag)
		setIntField(field, val)
	}
	return nil
}

func setStringField(field reflect.Value, val string) {
	if field.Kind() == reflect.Pointer {
		if val == "" {
			return
		}
		ptr := reflect.New(field.Type().Elem())
		ptr.Elem().SetString(val)
		field.Set(ptr)
	} else {
		field.SetString(val)
	}
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
