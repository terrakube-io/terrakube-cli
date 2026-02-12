package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.History]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name: "history",
		Parents: []resource.ParentScope{
			{
				Name:     "organization",
				IDFlag:   "organization-id",
				NameFlag: "organization-name",
				Resolver: orgResolver,
			},
			{
				Name:     "workspace",
				IDFlag:   "workspace-id",
				NameFlag: "workspace-name",
				Resolver: workspaceResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "JobReference", Flag: "job-reference", Type: resource.String, Description: "Job reference"},
			{StructField: "Output", Flag: "output", Type: resource.String, Description: "Output data"},
			{StructField: "Serial", Flag: "serial", Type: resource.Int, Description: "Serial number"},
			{StructField: "Md5", Flag: "md5", Type: resource.String, Description: "MD5 hash"},
			{StructField: "Lineage", Flag: "lineage", Type: resource.String, Description: "Lineage identifier"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.History, error) {
			return c.History.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.History, error) {
			return c.History.Get(ctx, pIDs[0], pIDs[1], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, h *terrakube.History) (*terrakube.History, error) {
			return c.History.Create(ctx, pIDs[0], pIDs[1], h)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, h *terrakube.History) (*terrakube.History, error) {
			return c.History.Update(ctx, pIDs[0], pIDs[1], h)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.History.Delete(ctx, pIDs[0], pIDs[1], id)
		},
	})
}
