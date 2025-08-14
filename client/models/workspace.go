package models

type GetBodyWorkspace struct {

	// data
	Data []*Workspace `json:"data"`
}

type PostBodyWorkspace struct {

	// data
	Data *Workspace `json:"data"`
}

type Workspace struct {

	// attributes
	Attributes *WorkspaceAttributes `json:"attributes,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// relationships
	Relationships *WorkspaceRelationships `json:"relationships,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

type WorkspaceAttributes struct {

	// name
	Name string `json:"name,omitempty"`

	// source
	Source string `json:"source,omitempty"`

	// branch
	Branch string `json:"branch,omitempty"`

	// terraformVersion
	TerraformVersion string `json:"terraformVersion,omitempty"`
}

type WorkspaceRelationships struct {
	// job
	Job *WorkspaceRelationshipsJob `json:"job,omitempty"`

	// organization
	Organization *WorkspaceRelationshipsOrganization `json:"organization,omitempty"`

	// variable
	Variable *WorkspaceRelationshipsVariable `json:"variable,omitempty"`
}

type WorkspaceRelationshipsJob struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [job]
	Type string `json:"type,omitempty"`
}

type WorkspaceRelationshipsOrganization struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [organization]
	Type string `json:"type,omitempty"`
}

type WorkspaceRelationshipsVariable struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [variable]
	Type string `json:"type,omitempty"`
}
