package client

import (
	"fmt"
	"net/http"
	"terrakube/client/models"
)

type TeamClient struct {
	Client *Client
}

func (c *TeamClient) List(organizationId string, filter string) ([]*models.Team, error) {
	req, err := c.Client.newRequestWithFilter(http.MethodGet, fmt.Sprintf("organization/%v/team", organizationId), filter, nil)
	if err != nil {
		return nil, err
	}
	var resp models.GetBodyTeam
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *TeamClient) Create(organizationId string, team models.Team) (*models.Team, error) {
	reqBody := models.PostBodyTeam{
		Data: &team,
	}

	req, err := c.Client.newRequest(http.MethodPost, fmt.Sprintf("organization/%v/team", organizationId), reqBody)
	if err != nil {
		return nil, err
	}
	var resp models.PostBodyTeam
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *TeamClient) Delete(organizationID string, teamId string) error {
	req, err := c.Client.newRequest(http.MethodDelete, fmt.Sprintf("organization/%v/team/%v", organizationID, teamId), nil)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}

func (c *TeamClient) Update(organizationId string, team models.Team) error {
	reqBody := models.PostBodyTeam{
		Data: &team,
	}

	req, err := c.Client.newRequest(http.MethodPatch, fmt.Sprintf("organization/%v/team/%v", organizationId, team.ID), reqBody)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}
