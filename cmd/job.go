package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

var jobCmd = resource.Register(rootCmd, resource.Config[terrakube.Job]{
	Runtime: resource.Runtime{
		NewClient:  newClient,
		GetContext: getContext,
		GetOutput:  func() string { return output },
	},
	Name:    "job",
	Aliases: []string{"jobs"},
	Parents: []resource.ParentScope{{
		Name:      "organization",
		Flag:      "organization",
		ShortFlag: "o",
		Aliases:   []string{"org"},
		IDFlag:    "organization-id",
		Resolver:  orgResolver,
	}},
	Fields: []resource.FieldDef{
		{StructField: "Command", Flag: "command", Short: "c", Type: resource.String, Required: true, Description: "Command (e.g. plan, apply, destroy)"},
		{StructField: "Status", Flag: "status", Type: resource.String, Description: "Job status"},
		{StructField: "Comments", Flag: "comments", Type: resource.String, Description: "Job comments"},
		{StructField: "CommitID", Flag: "commit-id", Type: resource.String, Description: "Git commit ID"},
		{StructField: "OverrideBranch", Flag: "override-branch", Type: resource.String, Description: "Override branch"},
		{StructField: "OverrideSource", Flag: "override-source", Type: resource.String, Description: "Override source repo URL"},
		{StructField: "PlanChanges", Flag: "plan-changes", Type: resource.Bool, Description: "Plan changes flag"},
		{StructField: "Refresh", Flag: "refresh", Type: resource.Bool, Description: "Refresh flag"},
		{StructField: "RefreshOnly", Flag: "refresh-only", Type: resource.Bool, Description: "Refresh only flag"},
		{StructField: "PRNumber", Flag: "pr-number", Type: resource.Int, Description: "Pull request number"},
		{StructField: "PRCommentError", Flag: "pr-comment-error", Type: resource.String, Description: "PR comment error"},
		{StructField: "TemplateReference", Flag: "template-reference", Type: resource.String, Description: "Template reference ID"},
	},
	List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Job, error) {
		return c.Jobs.List(ctx, pIDs[0], opts)
	},
	Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Job, error) {
		return c.Jobs.Get(ctx, pIDs[0], id)
	},
	Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, j *terrakube.Job) (*terrakube.Job, error) {
		wsVal := viper.GetString("workspace")
		if wsVal == "" {
			wsVal = viper.GetString("workspace-id")
		}
		if wsVal == "" {
			return nil, fmt.Errorf("--workspace is required")
		}
		if resource.IsUUID(wsVal) {
			j.Workspace = &terrakube.Workspace{ID: wsVal}
		} else {
			wsID, err := workspaceResolver(ctx, c, pIDs, wsVal)
			if err != nil {
				return nil, err
			}
			j.Workspace = &terrakube.Workspace{ID: wsID}
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

var cancelJobCmd = &cobra.Command{
	Use:   "cancel",
	Short: "cancel a running job",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()
		orgID, _ := cmd.Flags().GetString("organization-id")
		jobID, _ := cmd.Flags().GetString("id")

		job := &terrakube.Job{ID: jobID, Status: "cancelled"}
		updated, err := client.Jobs.Update(ctx, orgID, job)
		if err != nil {
			return err
		}
		renderOutput(updated, output)
		return nil
	},
}

var approveJobCmd = &cobra.Command{
	Use:   "approve",
	Short: "approve a job awaiting approval",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()
		orgID, _ := cmd.Flags().GetString("organization-id")
		jobID, _ := cmd.Flags().GetString("id")

		job := &terrakube.Job{ID: jobID, Status: "approved"}
		updated, err := client.Jobs.Update(ctx, orgID, job)
		if err != nil {
			return err
		}
		renderOutput(updated, output)
		return nil
	},
}

var rejectJobCmd = &cobra.Command{
	Use:   "reject",
	Short: "reject a job awaiting approval",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client := newClient()
		ctx := getContext()
		orgID, _ := cmd.Flags().GetString("organization-id")
		jobID, _ := cmd.Flags().GetString("id")

		job := &terrakube.Job{ID: jobID, Status: "rejected"}
		updated, err := client.Jobs.Update(ctx, orgID, job)
		if err != nil {
			return err
		}
		renderOutput(updated, output)
		return nil
	},
}

func init() {
	for _, subCmd := range []*cobra.Command{cancelJobCmd, approveJobCmd, rejectJobCmd} {
		subCmd.Flags().StringP("organization", "o", "", "organization ID or name")
		subCmd.Flags().String("organization-id", "", "organization ID")
		subCmd.Flags().String("id", "", "Job ID")
		_ = subCmd.MarkFlagRequired("id")
		jobCmd.AddCommand(subCmd)
	}
	for _, sub := range jobCmd.Commands() {
		if sub.Name() == "create" {
			sub.Flags().StringP("workspace", "w", "", "workspace ID or name (required)")
			sub.Flags().String("workspace-id", "", "workspace ID")
		}
	}
	fmt.Sprintf("") // ensure fmt package imported
}
