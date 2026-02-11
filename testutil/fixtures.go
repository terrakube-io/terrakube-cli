package testutil

import (
	terrakube "github.com/denniswebb/terrakube-go"
)

func strPtr(s string) *string { return &s }

// --- Organization ---

func FixtureOrganization() *terrakube.Organization {
	return &terrakube.Organization{
		ID:            "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Name:          "acme-corp",
		Description:   strPtr("ACME Corporation infrastructure"),
		ExecutionMode: "remote",
		Icon:          strPtr("https://cdn.example.com/icons/acme.png"),
	}
}

func FixtureOrganizationList() []*terrakube.Organization {
	return []*terrakube.Organization{
		FixtureOrganization(),
		{
			ID:            "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			Name:          "globex-corp",
			Description:   strPtr("Globex Corporation"),
			ExecutionMode: "local",
		},
		{
			ID:            "c3d4e5f6-a7b8-9012-cdef-123456789012",
			Name:          "initech",
			Description:   strPtr("Initech platform team"),
			ExecutionMode: "remote",
		},
	}
}

// --- Workspace ---

func FixtureWorkspace() *terrakube.Workspace {
	return &terrakube.Workspace{
		ID:            "d4e5f6a7-b8c9-0123-defa-234567890123",
		Name:          "production-vpc",
		Description:   strPtr("Production VPC infrastructure"),
		Source:        "https://github.com/acme-corp/infra.git",
		Folder:        "/",
		ExecutionMode: "remote",
		Branch:        "main",
		IaCType:       "terraform",
		IaCVersion:    "1.5.7",
	}
}

func FixtureWorkspaceList() []*terrakube.Workspace {
	return []*terrakube.Workspace{
		FixtureWorkspace(),
		{
			ID:            "e5f6a7b8-c9d0-1234-efab-345678901234",
			Name:          "staging-vpc",
			Description:   strPtr("Staging VPC infrastructure"),
			Source:        "https://github.com/acme-corp/infra.git",
			Folder:        "/staging",
			ExecutionMode: "remote",
			Branch:        "develop",
			IaCType:       "terraform",
			IaCVersion:    "1.5.7",
		},
	}
}

// --- Module ---

func FixtureModule() *terrakube.Module {
	return &terrakube.Module{
		ID:           "f6a7b8c9-d0e1-2345-fabc-456789012345",
		Name:         "vpc",
		Description:  "AWS VPC module",
		Provider:     "aws",
		Source:       "https://github.com/acme-corp/terraform-aws-vpc.git",
		TagPrefix:    strPtr("v"),
		Folder:       strPtr("/modules/vpc"),
		RegistryPath: strPtr("acme-corp/vpc/aws"),
	}
}

func FixtureModuleList() []*terrakube.Module {
	return []*terrakube.Module{
		FixtureModule(),
		{
			ID:           "a7b8c9d0-e1f2-3456-abcd-567890123456",
			Name:         "s3-bucket",
			Description:  "AWS S3 bucket module",
			Provider:     "aws",
			Source:       "https://github.com/acme-corp/terraform-aws-s3.git",
			RegistryPath: strPtr("acme-corp/s3-bucket/aws"),
		},
	}
}

// --- Variable ---

func FixtureVariable() *terrakube.Variable {
	return &terrakube.Variable{
		ID:          "b8c9d0e1-f2a3-4567-bcde-678901234567",
		Key:         "AWS_REGION",
		Value:       "us-east-1",
		Description: "AWS region for deployments",
		Category:    "ENV",
		Sensitive:   false,
		Hcl:         false,
	}
}

func FixtureVariableList() []*terrakube.Variable {
	return []*terrakube.Variable{
		FixtureVariable(),
		{
			ID:          "c9d0e1f2-a3b4-5678-cdef-789012345678",
			Key:         "DB_PASSWORD",
			Value:       "",
			Description: "Database password",
			Category:    "ENV",
			Sensitive:   true,
			Hcl:         false,
		},
	}
}

// --- Job ---

func FixtureJob() *terrakube.Job {
	return &terrakube.Job{
		ID:      "d0e1f2a3-b4c5-6789-defa-890123456789",
		Command: "plan",
		Output:  "Terraform will perform the following actions...",
		Status:  "completed",
		Workspace: &terrakube.Workspace{
			ID: "d4e5f6a7-b8c9-0123-defa-234567890123",
		},
	}
}

func FixtureJobList() []*terrakube.Job {
	return []*terrakube.Job{
		FixtureJob(),
		{
			ID:      "e1f2a3b4-c5d6-7890-efab-901234567890",
			Command: "apply",
			Output:  "Apply complete! Resources: 3 added, 0 changed, 0 destroyed.",
			Status:  "completed",
		},
	}
}

// --- Template ---

func FixtureTemplate() *terrakube.Template {
	return &terrakube.Template{
		ID:          "t1a2b3c4-d5e6-7890-abcd-ef1234567890",
		Name:        "standard-plan",
		Description: strPtr("Standard Terraform plan template"),
		Version:     strPtr("1.0.0"),
		Content:     "flow:\n  - type: terraformPlan",
	}
}

func FixtureTemplateList() []*terrakube.Template {
	return []*terrakube.Template{
		FixtureTemplate(),
		{
			ID:      "t2b3c4d5-e6f7-8901-bcde-f12345678901",
			Name:    "custom-apply",
			Version: strPtr("2.0.0"),
			Content: "flow:\n  - type: terraformApply",
		},
	}
}

// --- VCS ---

func FixtureVCS() *terrakube.VCS {
	return &terrakube.VCS{
		ID:             "v1a2b3c4-d5e6-7890-abcd-ef1234567890",
		Name:           "github-main",
		Description:    "Main GitHub connection",
		VcsType:        "GITHUB",
		ConnectionType: "OAUTH",
		ClientID:       "client-id-123",
		ClientSecret:   "client-secret-456",
		Endpoint:       "https://github.com",
		APIURL:         "https://api.github.com",
		Status:         "ACTIVE",
		Callback:       strPtr("https://app.example.com/callback"),
	}
}

func FixtureVCSList() []*terrakube.VCS {
	return []*terrakube.VCS{
		FixtureVCS(),
		{
			ID:             "v2b3c4d5-e6f7-8901-bcde-f12345678901",
			Name:           "gitlab-secondary",
			Description:    "GitLab secondary connection",
			VcsType:        "GITLAB",
			ConnectionType: "SSH",
			Endpoint:       "https://gitlab.com",
			APIURL:         "https://gitlab.com/api/v4",
			Status:         "INACTIVE",
		},
	}
}

// --- WorkspaceTag ---

func FixtureWorkspaceTag() *terrakube.WorkspaceTag {
	return &terrakube.WorkspaceTag{
		ID:    "wt1a2b3c-d4e5-6789-abcd-ef1234567890",
		TagID: "tag-prod-001",
	}
}

func FixtureWorkspaceTagList() []*terrakube.WorkspaceTag {
	return []*terrakube.WorkspaceTag{
		FixtureWorkspaceTag(),
		{
			ID:    "wt2b3c4d-e5f6-7890-bcde-f12345678901",
			TagID: "tag-staging-002",
		},
	}
}

// --- Team ---

func FixtureTeam() *terrakube.Team {
	return &terrakube.Team{
		ID:               "f2a3b4c5-d6e7-8901-fabc-012345678901",
		Name:             "platform-engineering",
		ManageWorkspace:  true,
		ManageModule:     true,
		ManageProvider:   false,
		ManageState:      true,
		ManageCollection: false,
		ManageVcs:        true,
		ManageTemplate:   false,
	}
}

func FixtureTeamList() []*terrakube.Team {
	return []*terrakube.Team{
		FixtureTeam(),
		{
			ID:               "a3b4c5d6-e7f8-9012-abcd-123456789012",
			Name:             "developers",
			ManageWorkspace:  true,
			ManageModule:     false,
			ManageProvider:   false,
			ManageState:      false,
			ManageCollection: false,
			ManageVcs:        false,
			ManageTemplate:   false,
		},
	}
}
