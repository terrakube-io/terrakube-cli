package cmd

import (
	"context"
	"fmt"

	terrakube "github.com/terrakube-io/terrakube-go"

	"terrakube/internal/resource"
)

func init() {
	resource.Register(rootCmd, resource.Config[terrakube.CollectionReference]{
		Runtime: resource.Runtime{
			NewClient:  newClient,
			GetContext: getContext,
			GetOutput:  func() string { return output },
		},
		Name:    "collection-reference",
		Aliases: []string{"collection-references", "collection-refs"},
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
				Name:     "collection",
				Flag:     "collection",
				IDFlag:   "collection-id",
				Resolver: collectionResolver,
			},
			{
				Name:      "workspace",
				Flag:      "workspace",
				ShortFlag: "w",
				Aliases:   []string{"ws"},
				IDFlag:    "workspace-id",
				Optional:  true,
				Resolver:  workspaceResolver,
			},
			{
				Name:     "project",
				Flag:     "project",
				IDFlag:   "project-id",
				Optional: true,
				Resolver: projectResolver,
			},
		},
		Fields: []resource.FieldDef{
			{StructField: "Description", Flag: "description", Short: "d", Type: resource.String, Description: "Reference description"},
			{StructField: "Scope", Flag: "scope", Type: resource.String, Description: "WORKSPACE, PROJECT or ORGANIZATION. Inferred from --workspace or --project; pass ORGANIZATION to apply the collection to every workspace in the organization"},
		},
		List: func(ctx context.Context, c *terrakube.Client, pIDs []string, opts *terrakube.ListOptions) ([]*terrakube.CollectionReference, error) {
			return c.CollectionReferences.List(ctx, pIDs[0], pIDs[1], opts)
		},
		Get: func(ctx context.Context, c *terrakube.Client, _ []string, id string) (*terrakube.CollectionReference, error) {
			return c.CollectionReferences.Get(ctx, id)
		},
		Create: func(ctx context.Context, c *terrakube.Client, pIDs []string, ref *terrakube.CollectionReference) (*terrakube.CollectionReference, error) {
			if err := applyCollectionReferenceTarget(ref, pIDs); err != nil {
				return nil, err
			}
			return c.CollectionReferences.Create(ctx, pIDs[0], pIDs[1], ref)
		},
		Update: func(ctx context.Context, c *terrakube.Client, _ []string, ref *terrakube.CollectionReference) (*terrakube.CollectionReference, error) {
			return c.CollectionReferences.Update(ctx, ref)
		},
		Delete: func(ctx context.Context, c *terrakube.Client, _ []string, id string) error {
			return c.CollectionReferences.Delete(ctx, id)
		},
	})
}

// applyCollectionReferenceTarget sets what the reference attaches to from --workspace, --project and --scope, and
// rejects combinations the API would refuse. A reference with neither target is only valid with --scope ORGANIZATION,
// so a forgotten --workspace can never apply a collection to the whole organization.
func applyCollectionReferenceTarget(ref *terrakube.CollectionReference, pIDs []string) error {
	workspaceID, projectID := "", ""
	if len(pIDs) > 2 {
		workspaceID = pIDs[2]
	}
	if len(pIDs) > 3 {
		projectID = pIDs[3]
	}
	scope := ""
	if ref.Scope != nil {
		scope = *ref.Scope
	}
	if scope == "" {
		switch {
		case projectID != "":
			scope = terrakube.CollectionReferenceScopeProject
		default:
			scope = terrakube.CollectionReferenceScopeWorkspace
		}
	}

	switch {
	case workspaceID != "" && projectID != "":
		return fmt.Errorf("set only one of --workspace and --project")
	case scope == terrakube.CollectionReferenceScopeWorkspace && workspaceID == "":
		return fmt.Errorf("set --workspace, --project, or --scope ORGANIZATION")
	case scope == terrakube.CollectionReferenceScopeProject && projectID == "":
		return fmt.Errorf("--scope PROJECT needs --project")
	case scope == terrakube.CollectionReferenceScopeOrganization && (workspaceID != "" || projectID != ""):
		return fmt.Errorf("--scope ORGANIZATION takes neither --workspace nor --project")
	case scope == terrakube.CollectionReferenceScopeWorkspace && projectID != "",
		scope == terrakube.CollectionReferenceScopeProject && workspaceID != "":
		return fmt.Errorf("--scope %s does not match the target given", scope)
	}

	switch scope {
	case terrakube.CollectionReferenceScopeProject:
		ref.Project = &terrakube.Project{ID: projectID}
		ref.Scope = &scope
	case terrakube.CollectionReferenceScopeOrganization:
		ref.Scope = &scope
	default:
		// Workspace references send no scope, so they keep working against Terrakube versions before 2.35.
		ref.Scope = nil
		ref.Workspace = &terrakube.Workspace{ID: workspaceID}
	}
	return nil
}
