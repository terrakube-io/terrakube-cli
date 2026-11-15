package cmd

import (
	"context"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	wsCmd := resource.Register(rootCmd, resource.Config[terrakube.Workspace]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "workspace",
		Aliases: []string{"ws", "workspaces"},
		Parents: []resource.ParentScope{
			{
				Name:      "organization",
				Flag:      "organization",
				ShortFlag: "o",
				Aliases:   []string{"org"},
				IDFlag:    "organization-id",
				Resolver:  orgResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Name", Flag: "name", Short: "n", Type: resource.String, Required: true, Description: "Workspace name"},
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Workspace description"},
			{StructField: "Source", Flag: "source", Short: "s", Type: resource.String, Description: "Repository source URL"},
			{StructField: "Branch", Flag: "branch", Short: "b", Type: resource.String, Description: "VCS branch"},
			{StructField: "Folder", Flag: "folder", Short: "f", Type: resource.String, Description: "Workspace working folder"},
			{StructField: "IaCType", Flag: "iac-type", Short: "t", Type: resource.String, Description: "IaC type (terraform, tofu)"},
			{StructField: "IaCVersion", Flag: "iac-version", Short: "v", Type: resource.String, Description: "Terraform/Tofu version"},
			{StructField: "ExecutionMode", Flag: "execution-mode", Short: "e", Type: resource.String, Description: "Execution mode (remote, local)"},
			{StructField: "Deleted", Flag: "deleted", Type: resource.Bool, Description: "Mark workspace as deleted"},
			{StructField: "Vcs", Flag: "vcs", Type: resource.Relationship, RelType: "vcs", Resolver: vcsResolver, Description: "VCS connection, by name or ID (\"\" clears it)"},
			{StructField: "SSH", Flag: "ssh", Type: resource.Relationship, RelType: "ssh", Resolver: sshResolver, Description: "SSH key, by name or ID (\"\" clears it)"},
			{StructField: "Agent", Flag: "agent", Type: resource.Relationship, RelType: "agent", Resolver: agentResolver, Description: "Agent pool, by name or ID (\"\" clears it)"},
			{StructField: "Project", Flag: "project", Type: resource.Relationship, RelType: "project", Resolver: projectResolver, Description: "Project, by name or ID (\"\" clears it)"},
			{StructField: "TemplateID", Flag: "template", Type: resource.String, Resolver: templateResolver, Description: "Default template, by name or ID (\"\" clears it)"},
			{StructField: "RequireApproveToApply", Flag: "require-approve-to-apply", Type: resource.Bool, Description: "Require approval before an apply"},
			{StructField: "RestrictApplyToConfiguredBranch", Flag: "restrict-apply-to-configured-branch", Type: resource.Bool, Description: "Apply only runs of the workspace's branch"},
			{StructField: "AllowRemoteApply", Flag: "allow-remote-apply", Type: resource.Bool, Description: "Allow applies started from a remote CLI run"},
			{StructField: "GlobalRemoteState", Flag: "global-remote-state", Type: resource.Bool, Description: "Let every workspace in the organization read this state (server default: true)"},
			{StructField: "SharedIDs", Flag: "shared-ids", Type: resource.String, Description: "Comma-separated IDs of the workspaces that may read this state when --global-remote-state=false"},
			{StructField: "ModuleSSHKey", Flag: "module-ssh-key", Type: resource.String, Resolver: sshResolver, Description: "SSH key for private module sources, by name or ID (\"\" clears it)"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.Workspace, error) {
			return c.Workspaces.List(ctx, pIDs[0], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) (*terrakube.Workspace, error) {
			return c.Workspaces.Get(ctx, pIDs[0], id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, p terrakube.Patch) (*terrakube.Workspace, error) {
			return c.Workspaces.CreateFields(ctx, pIDs[0], p)
		},
		Update: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string, p terrakube.Patch) (*terrakube.Workspace, error) {
			return c.Workspaces.UpdateFields(ctx, pIDs[0], id, p)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, pIDs []string, id string) error {
			return c.Workspaces.Delete(ctx, pIDs[0], id)
		},
		// The server stops a workspace's schedules and jobs only when deleted=true is set.
		SoftDelete: &resource.SoftDelete{Attribute: "deleted", Done: "marked as deleted", Hard: true},
	})
	wsCmd.AddCommand(newWorkspaceLockCmd(true), newWorkspaceLockCmd(false))
}
