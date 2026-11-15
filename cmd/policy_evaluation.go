package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

// jobScope is the organization and job of a job's children; job IDs are numeric.
var jobScope = []resource.ParentScope{orgScope[0], {
	Name:   "job",
	Flag:   "job",
	IDFlag: "job-id",
	RawID:  true,
}}

func init() {
	// The executor writes evaluations; users only read them.
	resource.Register(rootCmd, resource.Config[terrakube.PolicyEvaluation]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "policy-evaluation",
		Aliases: []string{"policy-evaluations"},
		Parents: jobScope,
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.PolicyEvaluation, error) {
			return c.PolicyEvaluations.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.PolicyEvaluation, error) {
			return c.PolicyEvaluations.Get(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
