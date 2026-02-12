package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.OrganizationVariable]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "organization-variable",
		Aliases: []string{"org-var", "org-vars", "organization-variables"},
		Parents: []resource.ParentScope{{
			Name:      "organization",
			Flag:      "organization",
			ShortFlag: "o",
			Aliases:   []string{"org"},
			IDFlag:    "organization-id",
			Resolver:  orgResolver,
		}},
		Fields: []resource.FieldDef{
			{StructField: "Key", Flag: "key", Short: "k", Type: resource.String, Required: true, Description: "Variable key"},
			{StructField: "Value", Flag: "value", Short: "v", Type: resource.String, Required: true, Description: "Variable value"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Variable description"},
			{StructField: "Category", Flag: "category", Type: resource.String, Required: true, Description: "Variable category (ENV, TERRAFORM)"},
			{StructField: "Sensitive", Flag: "sensitive", Type: resource.Bool, Description: "Whether the variable is sensitive"},
			{StructField: "Hcl", Flag: "hcl", Type: resource.Bool, Description: "Whether the variable value is HCL"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.OrganizationVariable, error) {
			return c.OrganizationVariables.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.OrganizationVariable, error) {
			return c.OrganizationVariables.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.OrganizationVariable) (*terrakube.OrganizationVariable, error) {
			return c.OrganizationVariables.Create(ctx, pIDs[0], v)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, v *terrakube.OrganizationVariable) (*terrakube.OrganizationVariable, error) {
			return c.OrganizationVariables.Update(ctx, pIDs[0], v)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.OrganizationVariables.Delete(ctx, pIDs[0], id)
		},
	})
}
