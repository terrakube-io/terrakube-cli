package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.FederatedClaim]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "federated-claim",
		Aliases: []string{"federated-claims", "claim"},
		Parents: []resource.ParentScope{
			{
				Name:   "federated",
				Flag:   "federated",
				IDFlag: "federated-id",
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "ClaimKey", Flag: "claim-key", Type: resource.String, Required: true, Description: "Claim key name"},
			{StructField: "ClaimValue", Flag: "claim-value", Type: resource.String, Required: true, Description: "Claim value"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.FederatedClaim, error) {
			return c.FederatedClaims.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.FederatedClaim, error) {
			return c.FederatedClaims.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, fc *terrakube.FederatedClaim) (*terrakube.FederatedClaim, error) {
			return c.FederatedClaims.Create(ctx, pIDs[0], fc)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, fc *terrakube.FederatedClaim) (*terrakube.FederatedClaim, error) {
			return c.FederatedClaims.Update(ctx, pIDs[0], fc)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.FederatedClaims.Delete(ctx, pIDs[0], id)
		},
	})
}
