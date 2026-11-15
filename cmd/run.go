package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	terrakube "github.com/terrakube-io/terrakube-go"

	outputpkg "terrakube/internal/output"
	"terrakube/internal/resource"
)

// runScope is the organization and workspace a run starts in.
var runScope = []resource.ParentScope{orgScope[0], {
	Name:      "workspace",
	Flag:      "workspace",
	ShortFlag: "w",
	Aliases:   []string{"ws"},
	IDFlag:    "workspace-id",
	Resolver:  workspaceResolver,
}}

func init() {
	runCmd := &cobra.Command{
		Use:   "run plan|apply|destroy",
		Short: "start a plan, apply or destroy on a workspace",
		Long: `Start a job on a workspace. What the job does depends only on its template:
-t takes a template name or ID, then TERRAKUBE_TEMPLATE_PLAN/APPLY/DESTROY. Without
either, apply uses the workspace's default template; plan and destroy need one.`,
	}
	for _, kind := range []string{"plan", "apply", "destroy"} {
		runCmd.AddCommand(newRunKindCmd(kind))
	}
	rootCmd.AddCommand(runCmd)
}

func newRunKindCmd(kind string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          kind,
		Short:        "start a " + kind + " job on a workspace",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout < 0 {
				return errors.New("--timeout must not be negative")
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			ctx := getContext()
			ids, err := resource.ResolveScope(ctx, client, cmd, runScope)
			if err != nil {
				return err
			}
			tpl, err := runTemplate(ctx, client, cmd, kind, ids)
			if err != nil {
				return err
			}

			p := terrakube.Patch{
				Attributes:    map[string]any{"templateReference": tpl},
				Relationships: map[string]terrakube.Relationship{"workspace": {Type: "workspace", ID: &ids[1]}},
			}
			if v, _ := cmd.Flags().GetStringArray("target"); len(v) > 0 {
				p.Attributes["targetAddrs"] = v
			}
			if v, _ := cmd.Flags().GetStringArray("replace"); len(v) > 0 {
				p.Attributes["replaceAddrs"] = v
			}
			if v, _ := cmd.Flags().GetString("branch"); v != "" {
				p.Attributes["overrideBranch"] = v
			}
			job, err := client.Jobs.CreateFields(ctx, ids[0], p)
			if err != nil {
				return err
			}
			if wait, _ := cmd.Flags().GetBool("wait"); wait {
				return waitForJob(ctx, client, ids[0], job.ID, timeout)
			}
			return outputpkg.Render(os.Stdout, job, output)
		},
	}
	resource.AddScopeFlags(cmd, runScope)
	cmd.Flags().StringP("template", "t", "", "template name or ID (default: TERRAKUBE_TEMPLATE_"+strings.ToUpper(kind)+runDefaultHint(kind)+")")
	cmd.Flags().StringArray("target", nil, "resource address to target (repeatable)")
	cmd.Flags().StringArray("replace", nil, "resource address to replace (repeatable)")
	cmd.Flags().String("branch", "", "run this branch instead of the workspace's")
	cmd.Flags().Bool("wait", false, "wait until the job reaches a final status (see: job wait)")
	cmd.Flags().Duration("timeout", 0, "with --wait, give up after this long (0 waits without a limit)")
	return cmd
}

func runDefaultHint(kind string) string {
	if kind == "apply" {
		return ", then the workspace's default template"
	}
	return ""
}

// runTemplate returns the template ID for a run: -t, then the env, then (for
// apply only) the workspace's default template. A name is resolved to its ID.
func runTemplate(ctx context.Context, c *terrakube.Client, cmd *cobra.Command, kind string, ids []string) (string, error) {
	envKey := "TERRAKUBE_TEMPLATE_" + strings.ToUpper(kind)
	val, _ := cmd.Flags().GetString("template")
	if val == "" {
		val = os.Getenv(envKey)
	}
	if val == "" {
		if kind != "apply" {
			return "", fmt.Errorf("run %s needs a template: pass -t <name|ID> or set %s", kind, envKey)
		}
		ws, err := c.Workspaces.Get(ctx, ids[0], ids[1])
		if err != nil {
			return "", err
		}
		if ws.TemplateID == "" {
			return "", fmt.Errorf("workspace %q has no default template: pass -t <name|ID> or set %s", ws.Name, envKey)
		}
		return ws.TemplateID, nil
	}
	if resource.IsUUID(val) {
		return val, nil
	}
	return templateResolver(ctx, c, ids, val)
}
