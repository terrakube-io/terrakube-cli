package models

type GetBodyModule struct {

	// data
	Data []*Module `json:"data"`
}

type PostBodyModule struct {

	// data
	Data *Module `json:"data"`
}

type Module struct {

	// attributes
	Attributes *ModuleAttributes `json:"attributes,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// relationships
	Relationships *ModuleRelationships `json:"relationships,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

type ModuleAttributes struct {

	// description
	Description string `json:"description,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// provider
	Provider string `json:"provider,omitempty"`

	// source
	Source string `json:"source,omitempty"`

	// tagprefix
	TagPrefix *string `json:"tagPrefix,omitempty"`

	// folder
	Folder *string `json:"folder,omitempty"`

	// registryPath
	RegistryPath string `json:"registryPath,omitempty"`

	// versions
	Versions []string `json:"versions,omitempty"`
}

type ModuleRelationships struct {

	// definition
	Definition *ModuleRelationshipsDefinition `json:"definition,omitempty"`

	// organization
	Organization *ModuleRelationshipsOrganization `json:"organization,omitempty"`
}

type ModuleRelationshipsDefinition struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [definition]
	Type string `json:"type,omitempty"`
}

type ModuleRelationshipsOrganization struct {

	// id
	ID string `json:"id,omitempty"`

	// type
	// Enum: [organization]
	Type string `json:"type,omitempty"`
}
