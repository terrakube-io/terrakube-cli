package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

var variableCmd = resource.Register(workspaceCmd, resource.Config[terrakube.Variable]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "variable",
		Aliases: []string{"variables", "var"},
		Parents: []resource.ParentScope{
			{
				Name:      "organization",
				Flag:      "organization",
				ShortFlag: "o",
				Aliases:   []string{"org"},
				IDFlag:    "organization-id",
				Resolver:  orgResolver,
			},
			{
				Name:      "workspace",
				Flag:      "workspace",
				ShortFlag: "w",
				Aliases:   []string{"ws"},
				IDFlag:    "workspace-id",
				Resolver:  workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Key", Flag: "key", Short: "k", Type: resource.String, Required: true, Description: "Variable key"},
			{StructField: "Value", Flag: "value", Short: "v", Type: resource.String, Description: "Variable value"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Variable description"},
			{StructField: "Category", Flag: "category", Short: "c", Type: resource.String, Required: true, Description: "Category (TERRAFORM or ENV)"},
			{StructField: "Sensitive", Flag: "sensitive", Type: resource.Bool, Description: "Mark variable as sensitive"},
			{StructField: "Hcl", Flag: "hcl", Type: resource.Bool, Description: "Parse value as HCL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Variable, error) {
			return c.Variables.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Variable, error) {
			return c.Variables.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.Variable) (*terrakube.Variable, error) {
			return c.Variables.Create(ctx, pIDs[0], pIDs[1], v)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.Variable) (*terrakube.Variable, error) {
			return c.Variables.Update(ctx, pIDs[0], pIDs[1], v)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Variables.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
