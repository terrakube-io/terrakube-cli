package client

import (
	"fmt"
	"net/http"
	"terrakube/client/models"
)

type WorkspaceClient struct {
	Client *Client
}

func (c *WorkspaceClient) List(organizationId string, filter string) ([]*models.Workspace, error) {
	req, err := c.Client.newRequestWithFilter(http.MethodGet, fmt.Sprintf("organization/%v/workspace", organizationId), filter, nil)
	if err != nil {
		return nil, err
	}
	var resp models.GetBodyWorkspace
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *WorkspaceClient) Create(organizationId string, workspace models.Workspace) (*models.Workspace, error) {
	reqBody := models.PostBodyWorkspace{
		Data: &workspace,
	}

	req, err := c.Client.newRequest(http.MethodPost, fmt.Sprintf("organization/%v/workspace", organizationId), reqBody)
	if err != nil {
		return nil, err
	}
	var resp models.PostBodyWorkspace
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *WorkspaceClient) Delete(organizationID string, workspaceId string) error {
	req, err := c.Client.newRequest(http.MethodDelete, fmt.Sprintf("organization/%v/workspace/%v", organizationID, workspaceId), nil)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}

func (c *WorkspaceClient) Update(organizationId string, workspace models.Workspace) error {
	reqBody := models.PostBodyWorkspace{
		Data: &workspace,
	}

	req, err := c.Client.newRequest(http.MethodPatch, fmt.Sprintf("organization/%v/workspace/%v", organizationId, workspace.ID), reqBody)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}
