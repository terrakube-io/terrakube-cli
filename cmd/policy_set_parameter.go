package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

// policySetScope is the organization and policy set of policy-set children.
var policySetScope = []resource.ParentScope{orgScope[0], {
	Name:     "policy-set",
	Flag:     "policy-set",
	IDFlag:   "policy-set-id",
	Resolver: policySetResolver,
}}

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.PolicySetParameter]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-set-parameter",
		Aliases: []string{"policy-set-parameters"},
		Parents: policySetScope,
		Fields: []resource.FieldDef{
			{StructField: "Key", Flag: "key", Short: "k", Type: resource.String, Required: true, Description: "Parameter key"},
			{StructField: "Value", Flag: "value", Short: "v", Type: resource.String, Required: true, Description: "Parameter value"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Parameter description"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.PolicySetParameter, error) {
			return c.PolicySetParameters.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicySetParameter, error) {
			return c.PolicySetParameters.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.PolicySetParameter, error) {
			return c.PolicySetParameters.CreateFields(ctx, pIDs[0], pIDs[1], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.PolicySetParameter, error) {
			return c.PolicySetParameters.UpdateFields(ctx, pIDs[0], pIDs[1], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.PolicySetParameters.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
