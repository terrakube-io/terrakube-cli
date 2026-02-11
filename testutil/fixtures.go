package testutil

import (
	"terrakube/client/models"
)

func strPtr(s string) *string { return &s }

// --- Organization ---

func FixtureOrganization() *models.Organization {
	return &models.Organization{
		ID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Type: "organization",
		Attributes: &models.OrganizationAttributes{
			Name:          "acme-corp",
			Description:   strPtr("ACME Corporation infrastructure"),
			ExecutionMode: strPtr("remote"),
			Icon:          strPtr("https://cdn.example.com/icons/acme.png"),
		},
	}
}

func FixtureOrganizationList() []*models.Organization {
	return []*models.Organization{
		FixtureOrganization(),
		{
			ID:   "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			Type: "organization",
			Attributes: &models.OrganizationAttributes{
				Name:          "globex-corp",
				Description:   strPtr("Globex Corporation"),
				ExecutionMode: strPtr("local"),
			},
		},
		{
			ID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
			Type: "organization",
			Attributes: &models.OrganizationAttributes{
				Name:          "initech",
				Description:   strPtr("Initech platform team"),
				ExecutionMode: strPtr("remote"),
			},
		},
	}
}

func FixtureGetBodyOrganization() models.GetBodyOrganization {
	return models.GetBodyOrganization{Data: FixtureOrganizationList()}
}

func FixturePostBodyOrganization() models.PostBodyOrganization {
	return models.PostBodyOrganization{Data: FixtureOrganization()}
}

// --- Workspace ---

func FixtureWorkspace() *models.Workspace {
	return &models.Workspace{
		ID:   "d4e5f6a7-b8c9-0123-defa-234567890123",
		Type: "workspace",
		Attributes: &models.WorkspaceAttributes{
			Name:             "production-vpc",
			Description:      "Production VPC infrastructure",
			Source:           "https://github.com/acme-corp/infra.git",
			Folder:           "/",
			ExecutionMode:    "remote",
			Branch:           "main",
			IacType:          "terraform",
			TerraformVersion: "1.5.7",
		},
		Relationships: &models.WorkspaceRelationships{
			Organization: &models.WorkspaceRelationshipsOrganization{
				ID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Type: "organization",
			},
		},
	}
}

func FixtureWorkspaceList() []*models.Workspace {
	return []*models.Workspace{
		FixtureWorkspace(),
		{
			ID:   "e5f6a7b8-c9d0-1234-efab-345678901234",
			Type: "workspace",
			Attributes: &models.WorkspaceAttributes{
				Name:             "staging-vpc",
				Description:      "Staging VPC infrastructure",
				Source:           "https://github.com/acme-corp/infra.git",
				Folder:           "/staging",
				ExecutionMode:    "remote",
				Branch:           "develop",
				IacType:          "terraform",
				TerraformVersion: "1.5.7",
			},
		},
	}
}

func FixtureGetBodyWorkspace() models.GetBodyWorkspace {
	return models.GetBodyWorkspace{Data: FixtureWorkspaceList()}
}

func FixturePostBodyWorkspace() models.PostBodyWorkspace {
	return models.PostBodyWorkspace{Data: FixtureWorkspace()}
}

// --- Module ---

func FixtureModule() *models.Module {
	tagPrefix := "v"
	folder := "/modules/vpc"
	return &models.Module{
		ID:   "f6a7b8c9-d0e1-2345-fabc-456789012345",
		Type: "module",
		Attributes: &models.ModuleAttributes{
			Name:         "vpc",
			Description:  "AWS VPC module",
			Provider:     "aws",
			Source:       "https://github.com/acme-corp/terraform-aws-vpc.git",
			TagPrefix:    &tagPrefix,
			Folder:       &folder,
			RegistryPath: "acme-corp/vpc/aws",
			Versions:     []string{"1.0.0", "1.1.0", "2.0.0"},
		},
		Relationships: &models.ModuleRelationships{
			Organization: &models.ModuleRelationshipsOrganization{
				ID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Type: "organization",
			},
		},
	}
}

func FixtureModuleList() []*models.Module {
	return []*models.Module{
		FixtureModule(),
		{
			ID:   "a7b8c9d0-e1f2-3456-abcd-567890123456",
			Type: "module",
			Attributes: &models.ModuleAttributes{
				Name:         "s3-bucket",
				Description:  "AWS S3 bucket module",
				Provider:     "aws",
				Source:       "https://github.com/acme-corp/terraform-aws-s3.git",
				RegistryPath: "acme-corp/s3-bucket/aws",
				Versions:     []string{"1.0.0"},
			},
		},
	}
}

func FixtureGetBodyModule() models.GetBodyModule {
	return models.GetBodyModule{Data: FixtureModuleList()}
}

func FixturePostBodyModule() models.PostBodyModule {
	return models.PostBodyModule{Data: FixtureModule()}
}

// --- Variable ---

func FixtureVariable() *models.Variable {
	return &models.Variable{
		ID:   "b8c9d0e1-f2a3-4567-bcde-678901234567",
		Type: "variable",
		Attributes: &models.VariableAttributes{
			Key:         "AWS_REGION",
			Value:       "us-east-1",
			Description: "AWS region for deployments",
			Category:    "ENV",
			Sensitive:   false,
			Hcl:         false,
		},
		Relationships: &models.VariableRelationships{
			Workspace: &models.VariableRelationshipsWorkspace{
				ID:   "d4e5f6a7-b8c9-0123-defa-234567890123",
				Type: "workspace",
			},
		},
	}
}

func FixtureVariableList() []*models.Variable {
	return []*models.Variable{
		FixtureVariable(),
		{
			ID:   "c9d0e1f2-a3b4-5678-cdef-789012345678",
			Type: "variable",
			Attributes: &models.VariableAttributes{
				Key:         "DB_PASSWORD",
				Value:       "",
				Description: "Database password",
				Category:    "ENV",
				Sensitive:   true,
				Hcl:         false,
			},
		},
	}
}

func FixtureGetBodyVariable() models.GetBodyVariable {
	return models.GetBodyVariable{Data: FixtureVariableList()}
}

func FixturePostBodyVariable() models.PostBodyVariable {
	return models.PostBodyVariable{Data: FixtureVariable()}
}

// --- Job ---

func FixtureJob() *models.Job {
	return &models.Job{
		ID:   "d0e1f2a3-b4c5-6789-defa-890123456789",
		Type: "job",
		Attributes: &models.JobAttributes{
			Command: "plan",
			Output:  "Terraform will perform the following actions...",
			Status:  "completed",
		},
		Relationships: &models.JobRelationships{
			Organization: &models.JobRelationshipsOrganization{
				ID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Type: "organization",
			},
			Workspace: &models.JobRelationshipsWorkspace{
				Data: &models.JobRelationshipsWorkspaceData{
					ID:   "d4e5f6a7-b8c9-0123-defa-234567890123",
					Type: "workspace",
				},
			},
		},
	}
}

func FixtureJobList() []*models.Job {
	return []*models.Job{
		FixtureJob(),
		{
			ID:   "e1f2a3b4-c5d6-7890-efab-901234567890",
			Type: "job",
			Attributes: &models.JobAttributes{
				Command: "apply",
				Output:  "Apply complete! Resources: 3 added, 0 changed, 0 destroyed.",
				Status:  "completed",
			},
		},
	}
}

func FixtureGetBodyJob() models.GetBodyJob {
	return models.GetBodyJob{Data: FixtureJobList()}
}

func FixturePostBodyJob() models.PostBodyJob {
	return models.PostBodyJob{Data: FixtureJob()}
}

// --- Team ---

func FixtureTeam() *models.Team {
	return &models.Team{
		ID:   "f2a3b4c5-d6e7-8901-fabc-012345678901",
		Type: "team",
		Attributes: &models.TeamAttributes{
			Name:             "platform-engineering",
			ManageWorkspace:  true,
			ManageModule:     true,
			ManageProvider:   false,
			ManageState:      true,
			ManageCollection: false,
			ManageVcs:        true,
			ManageTemplate:   false,
		},
	}
}

func FixtureTeamList() []*models.Team {
	return []*models.Team{
		FixtureTeam(),
		{
			ID:   "a3b4c5d6-e7f8-9012-abcd-123456789012",
			Type: "team",
			Attributes: &models.TeamAttributes{
				Name:             "developers",
				ManageWorkspace:  true,
				ManageModule:     false,
				ManageProvider:   false,
				ManageState:      false,
				ManageCollection: false,
				ManageVcs:        false,
				ManageTemplate:   false,
			},
		},
	}
}

func FixtureGetBodyTeam() models.GetBodyTeam {
	return models.GetBodyTeam{Data: FixtureTeamList()}
}

func FixturePostBodyTeam() models.PostBodyTeam {
	return models.PostBodyTeam{Data: FixtureTeam()}
}
