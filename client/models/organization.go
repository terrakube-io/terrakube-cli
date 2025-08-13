package models

type GetBodyOrganization struct {

	// data
	Data []*Organization `json:"data"`
}

type PostBodyOrganization struct {

	// data
	Data *Organization `json:"data"`
}
type Organization struct {

	// attributes
	Attributes *OrganizationAttributes `json:"attributes,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// relationships
	Relationships *OrganizationRelationships `json:"relationships,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

type OrganizationAttributes struct {

	// description
	Description *string `json:"description,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// executionMode
	ExecutionMode *string `json:"executionMode,omitempty"`

	// icon
	Icon *string `json:"icon,omitempty"`
}

type OrganizationRelationships struct {

	// job
	Job *OrganizationRelationshipsJob `json:"job,omitempty"`

	// module
	Module *OrganizationRelationshipsModule `json:"module,omitempty"`

	// workspace
	Workspace *OrganizationRelationshipsWorkspace `json:"workspace,omitempty"`
}

type OrganizationRelationshipsJob struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [job]
	Type string `json:"type,omitempty"`
}

type OrganizationRelationshipsModule struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [module]
	Type string `json:"type,omitempty"`
}

type OrganizationRelationshipsWorkspace struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [workspace]
	Type string `json:"type,omitempty"`
}
