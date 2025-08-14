package models

type GetBodyTeam struct {

	// data
	Data []*Team `json:"data"`
}

type PostBodyTeam struct {

	// data
	Data *Team `json:"data"`
}

type Team struct {

	// attributes
	Attributes *TeamAttributes `json:"attributes,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

type TeamAttributes struct {

	// name
	Name string `json:"name,omitempty"`

	// manageWorkspace
	ManageWorkspace bool `json:"manageWorkspace,omitempty"`

	// manageModule
	ManageModule bool `json:"manageModule,omitempty"`

	// manageProvider
	ManageProvider bool `json:"manageProvider,omitempty"`

	// manageState
	ManageState bool `json:"manageState,omitempty"`

	// manageCollection
	ManageCollection bool `json:"manageCollection,omitempty"`

	// manageVcs
	ManageVcs bool `json:"manageVcs,omitempty"`

	// manageTemplate
	ManageTemplate bool `json:"manageTemplate,omitempty"`
}
