package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.Job]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "job",
		Aliases: []string{"jobs"},
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
			{StructField: "Command", Flag: "command", Short: "c", Type: resource.String, Required: true, Description: "Command to execute (plan, apply, destroy)"},
			{StructField: "Output", Flag: "output", Type: resource.String, Description: "Job output log"},
			{StructField: "Status", Flag: "status", Type: resource.String, Description: "Job status"},
			{StructField: "CommitID", Flag: "commit-id", Type: resource.String, Description: "VCS commit ID"},
			{StructField: "OverrideBranch", Flag: "override-branch", Type: resource.String, Description: "Override branch"},
			{StructField: "OverrideSource", Flag: "override-source", Type: resource.String, Description: "Override source repository"},
			{StructField: "PlanChanges", Flag: "plan-changes", Type: resource.Bool, Description: "Whether plan contains changes"},
			{StructField: "Refresh", Flag: "refresh", Type: resource.Bool, Description: "Whether to refresh state"},
			{StructField: "RefreshOnly", Flag: "refresh-only", Type: resource.Bool, Description: "Refresh state only"},
			{StructField: "TargetAddrs", Flag: "target-addrs", Type: resource.StringSlice, Description: "Target resource addresses"},
			{StructField: "ReplaceAddrs", Flag: "replace-addrs", Type: resource.StringSlice, Description: "Replace resource addresses"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Job, error) {
			return c.Jobs.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Job, error) {
			return c.Jobs.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, j *terrakube.Job) (*terrakube.Job, error) {
			if len(pIDs) > 1 && pIDs[1] != "" {
				j.Workspace = &terrakube.Workspace{ID: pIDs[1]}
			}
			return c.Jobs.Create(ctx, pIDs[0], j)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, j *terrakube.Job) (*terrakube.Job, error) {
			return c.Jobs.Update(ctx, pIDs[0], j)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Jobs.Delete(ctx, pIDs[0], id)
		},
	})
}
