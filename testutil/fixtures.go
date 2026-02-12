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

func boolPtr(b bool) *bool { return &b }

// --- OrganizationVariable ---

func FixtureOrganizationVariable() *terrakube.OrganizationVariable {
	return &terrakube.OrganizationVariable{
		ID:          "ov1a2b3c-d4e5-6789-abcd-ef1234567890",
		Key:         "TF_LOG",
		Value:       "DEBUG",
		Description: "Terraform log level",
		Category:    "ENV",
		Sensitive:   boolPtr(false),
		Hcl:         false,
	}
}

func FixtureOrganizationVariableList() []*terrakube.OrganizationVariable {
	return []*terrakube.OrganizationVariable{
		FixtureOrganizationVariable(),
		{
			ID:          "ov2b3c4d-e5f6-7890-bcde-f12345678901",
			Key:         "AWS_DEFAULT_REGION",
			Value:       "us-west-2",
			Description: "Default AWS region",
			Category:    "ENV",
			Sensitive:   boolPtr(false),
			Hcl:         false,
		},
	}
}

// --- Tag ---

func FixtureTag() *terrakube.Tag {
	return &terrakube.Tag{
		ID:   "tg1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name: "production",
	}
}

func FixtureTagList() []*terrakube.Tag {
	return []*terrakube.Tag{
		FixtureTag(),
		{
			ID:   "tg2b3c4d-e5f6-7890-bcde-f12345678901",
			Name: "staging",
		},
	}
}

// --- SSH ---

func FixtureSSH() *terrakube.SSH {
	return &terrakube.SSH{
		ID:          "ss1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name:        "deploy-key",
		Description: strPtr("Deployment SSH key"),
		PrivateKey:  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
		SSHType:     "rsa",
	}
}

func FixtureSSHList() []*terrakube.SSH {
	return []*terrakube.SSH{
		FixtureSSH(),
		{
			ID:         "ss2b3c4d-e5f6-7890-bcde-f12345678901",
			Name:       "backup-key",
			PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
			SSHType:    "ed25519",
		},
	}
}

// --- Agent ---

func FixtureAgent() *terrakube.Agent {
	return &terrakube.Agent{
		ID:          "ag1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name:        "k8s-runner",
		Description: "Kubernetes-based runner agent",
		URL:         "https://agent.example.com",
	}
}

func FixtureAgentList() []*terrakube.Agent {
	return []*terrakube.Agent{
		FixtureAgent(),
		{
			ID:          "ag2b3c4d-e5f6-7890-bcde-f12345678901",
			Name:        "docker-runner",
			Description: "Docker-based runner agent",
			URL:         "https://docker-agent.example.com",
		},
	}
}

// --- Collection ---

func FixtureCollection() *terrakube.Collection {
	return &terrakube.Collection{
		ID:          "cl1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name:        "shared-vars",
		Description: strPtr("Shared variable collection"),
		Priority:    10,
	}
}

func FixtureCollectionList() []*terrakube.Collection {
	return []*terrakube.Collection{
		FixtureCollection(),
		{
			ID:       "cl2b3c4d-e5f6-7890-bcde-f12345678901",
			Name:     "env-config",
			Priority: 5,
		},
	}
}

// --- CollectionItem ---

func FixtureCollectionItem() *terrakube.CollectionItem {
	return &terrakube.CollectionItem{
		ID:          "ci1a2b3c-d4e5-6789-abcd-ef1234567890",
		Key:         "DB_HOST",
		Value:       "db.example.com",
		Description: strPtr("Database hostname"),
		Category:    "ENV",
		Sensitive:   false,
		Hcl:         false,
	}
}

func FixtureCollectionItemList() []*terrakube.CollectionItem {
	return []*terrakube.CollectionItem{
		FixtureCollectionItem(),
		{
			ID:        "ci2b3c4d-e5f6-7890-bcde-f12345678901",
			Key:       "DB_PORT",
			Value:     "5432",
			Category:  "ENV",
			Sensitive: false,
			Hcl:       false,
		},
	}
}

// --- CollectionReference ---

func FixtureCollectionReference() *terrakube.CollectionReference {
	return &terrakube.CollectionReference{
		ID:          "cr1a2b3c-d4e5-6789-abcd-ef1234567890",
		Description: strPtr("Production workspace reference"),
	}
}

func FixtureCollectionReferenceList() []*terrakube.CollectionReference {
	return []*terrakube.CollectionReference{
		FixtureCollectionReference(),
		{
			ID:          "cr2b3c4d-e5f6-7890-bcde-f12345678901",
			Description: strPtr("Staging workspace reference"),
		},
	}
}

// --- WorkspaceAccess ---

func FixtureWorkspaceAccess() *terrakube.WorkspaceAccess {
	return &terrakube.WorkspaceAccess{
		ID:              "wa1a2b3c-d4e5-6789-abcd-ef1234567890",
		ManageState:     true,
		ManageWorkspace: true,
		ManageJob:       false,
		Name:            "platform-engineering",
	}
}

func FixtureWorkspaceAccessList() []*terrakube.WorkspaceAccess {
	return []*terrakube.WorkspaceAccess{
		FixtureWorkspaceAccess(),
		{
			ID:              "wa2b3c4d-e5f6-7890-bcde-f12345678901",
			ManageState:     false,
			ManageWorkspace: false,
			ManageJob:       true,
			Name:            "developers",
		},
	}
}

// --- WorkspaceSchedule ---

func FixtureWorkspaceSchedule() *terrakube.WorkspaceSchedule {
	return &terrakube.WorkspaceSchedule{
		ID:         "ws1a2b3c-d4e5-6789-abcd-ef1234567890",
		Schedule:   "0 0 * * *",
		TemplateID: "tpl-abc-123",
	}
}

func FixtureWorkspaceScheduleList() []*terrakube.WorkspaceSchedule {
	return []*terrakube.WorkspaceSchedule{
		FixtureWorkspaceSchedule(),
		{
			ID:         "ws2b3c4d-e5f6-7890-bcde-f12345678901",
			Schedule:   "0 12 * * MON-FRI",
			TemplateID: "tpl-def-456",
		},
	}
}

// --- Webhook ---

func FixtureWebhook() *terrakube.Webhook {
	return &terrakube.Webhook{
		ID:           "wh1a2b3c-d4e5-6789-abcd-ef1234567890",
		Path:         "/",
		Branch:       "main",
		TemplateID:   "tpl-abc-123",
		RemoteHookID: "gh-hook-456",
		Event:        "PUSH",
	}
}

func FixtureWebhookList() []*terrakube.Webhook {
	return []*terrakube.Webhook{
		FixtureWebhook(),
		{
			ID:           "wh2b3c4d-e5f6-7890-bcde-f12345678901",
			Path:         "/modules",
			Branch:       "develop",
			TemplateID:   "tpl-def-456",
			RemoteHookID: "gh-hook-789",
			Event:        "TAG",
		},
	}
}

// --- WebhookEvent ---

func FixtureWebhookEvent() *terrakube.WebhookEvent {
	return &terrakube.WebhookEvent{
		ID:         "we1a2b3c-d4e5-6789-abcd-ef1234567890",
		Branch:     "main",
		Event:      "PUSH",
		Path:       "/",
		Priority:   1,
		TemplateID: "tpl-abc-123",
	}
}

func FixtureWebhookEventList() []*terrakube.WebhookEvent {
	return []*terrakube.WebhookEvent{
		FixtureWebhookEvent(),
		{
			ID:         "we2b3c4d-e5f6-7890-bcde-f12345678901",
			Branch:     "develop",
			Event:      "TAG",
			Path:       "/modules",
			Priority:   2,
			TemplateID: "tpl-def-456",
		},
	}
}

// --- History ---

func FixtureHistory() *terrakube.History {
	return &terrakube.History{
		ID:           "hi1a2b3c-d4e5-6789-abcd-ef1234567890",
		JobReference: "job-ref-001",
		Output:       "state output data",
		Serial:       1,
		Md5:          strPtr("d41d8cd98f00b204e9800998ecf8427e"),
		Lineage:      strPtr("lineage-abc-123"),
	}
}

func FixtureHistoryList() []*terrakube.History {
	return []*terrakube.History{
		FixtureHistory(),
		{
			ID:           "hi2b3c4d-e5f6-7890-bcde-f12345678901",
			JobReference: "job-ref-002",
			Output:       "updated state",
			Serial:       2,
		},
	}
}

// --- Action ---

func FixtureAction() *terrakube.Action {
	return &terrakube.Action{
		ID:              "ac1a2b3c-d4e5-6789-abcd-ef1234567890",
		Action:          "terraform-plan",
		Active:          true,
		Category:        "terraform",
		Description:     strPtr("Run terraform plan"),
		DisplayCriteria: strPtr("always"),
		Label:           "Plan",
		Name:            "tf-plan",
		Type:            "built-in",
		Version:         strPtr("1.0.0"),
	}
}

func FixtureActionList() []*terrakube.Action {
	return []*terrakube.Action{
		FixtureAction(),
		{
			ID:       "ac2b3c4d-e5f6-7890-bcde-f12345678901",
			Action:   "terraform-apply",
			Active:   false,
			Category: "terraform",
			Label:    "Apply",
			Name:     "tf-apply",
			Type:     "built-in",
		},
	}
}

// --- Step ---

func FixtureStep() *terrakube.Step {
	return &terrakube.Step{
		ID:         "st1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name:       "Plan Step",
		Output:     strPtr("Terraform plan output..."),
		Status:     "completed",
		StepNumber: 1,
	}
}

func FixtureStepList() []*terrakube.Step {
	return []*terrakube.Step{
		FixtureStep(),
		{
			ID:         "st2b3c4d-e5f6-7890-bcde-f12345678901",
			Name:       "Apply Step",
			Status:     "running",
			StepNumber: 2,
		},
	}
}

// --- Provider ---

func FixtureProvider() *terrakube.Provider {
	return &terrakube.Provider{
		ID:          "pr1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name:        "aws",
		Description: strPtr("AWS provider"),
	}
}

func FixtureProviderList() []*terrakube.Provider {
	return []*terrakube.Provider{
		FixtureProvider(),
		{
			ID:          "pr2b3c4d-e5f6-7890-bcde-f12345678901",
			Name:        "azurerm",
			Description: strPtr("Azure provider"),
		},
	}
}

// --- ProviderVersion ---

func FixtureProviderVersion() *terrakube.ProviderVersion {
	return &terrakube.ProviderVersion{
		ID:            "pv1a2b3c-d4e5-6789-abcd-ef1234567890",
		VersionNumber: "5.0.0",
		Protocols:     strPtr("5.0"),
	}
}

func FixtureProviderVersionList() []*terrakube.ProviderVersion {
	return []*terrakube.ProviderVersion{
		FixtureProviderVersion(),
		{
			ID:            "pv2b3c4d-e5f6-7890-bcde-f12345678901",
			VersionNumber: "4.67.0",
			Protocols:     strPtr("5.0"),
		},
	}
}

// --- Implementation ---

func FixtureImplementation() *terrakube.Implementation {
	return &terrakube.Implementation{
		ID:       "im1a2b3c-d4e5-6789-abcd-ef1234567890",
		Os:       "linux",
		Arch:     "amd64",
		Filename: "terraform-provider-aws_5.0.0_linux_amd64.zip",
	}
}

func FixtureImplementationList() []*terrakube.Implementation {
	return []*terrakube.Implementation{
		FixtureImplementation(),
		{
			ID:       "im2b3c4d-e5f6-7890-bcde-f12345678901",
			Os:       "darwin",
			Arch:     "arm64",
			Filename: "terraform-provider-aws_5.0.0_darwin_arm64.zip",
		},
	}
}

// --- ModuleVersion ---

func FixtureModuleVersion() *terrakube.ModuleVersion {
	return &terrakube.ModuleVersion{
		ID:      "mv1a2b3c-d4e5-6789-abcd-ef1234567890",
		Version: "1.0.0",
		Commit:  strPtr("abc123def"),
	}
}

func FixtureModuleVersionList() []*terrakube.ModuleVersion {
	return []*terrakube.ModuleVersion{
		FixtureModuleVersion(),
		{
			ID:      "mv2b3c4d-e5f6-7890-bcde-f12345678901",
			Version: "0.9.0",
		},
	}
}

// --- Address ---

func FixtureAddress() *terrakube.Address {
	return &terrakube.Address{
		ID:   "ad1a2b3c-d4e5-6789-abcd-ef1234567890",
		Name: "aws_vpc.main",
		Type: "aws_vpc",
	}
}

func FixtureAddressList() []*terrakube.Address {
	return []*terrakube.Address{
		FixtureAddress(),
		{
			ID:   "ad2b3c4d-e5f6-7890-bcde-f12345678901",
			Name: "aws_subnet.public",
			Type: "aws_subnet",
		},
	}
}

// --- GithubAppToken ---

func FixtureGithubAppToken() *terrakube.GithubAppToken {
	return &terrakube.GithubAppToken{
		ID:             "ga1a2b3c-d4e5-6789-abcd-ef1234567890",
		AppID:          "12345",
		InstallationID: "67890",
		Owner:          "acme-corp",
	}
}

func FixtureGithubAppTokenList() []*terrakube.GithubAppToken {
	return []*terrakube.GithubAppToken{
		FixtureGithubAppToken(),
		{
			ID:             "ga2b3c4d-e5f6-7890-bcde-f12345678901",
			AppID:          "54321",
			InstallationID: "09876",
			Owner:          "globex-corp",
		},
	}
}
