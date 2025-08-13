package client

import (
	"fmt"
	"net/http"
	"terrakube/client/models"
)

type OrganizationClient struct {
	Client *Client
}

func (c *OrganizationClient) List(filter string) ([]*models.Organization, error) {
	req, err := c.Client.newRequestWithFilter(http.MethodGet, "organization", filter, nil)
	if err != nil {
		return nil, err
	}
	var organizationResp models.GetBodyOrganization
	_, err = c.Client.do(req, &organizationResp)
	return organizationResp.Data, err
}

func (c *OrganizationClient) Create(organization models.Organization) (*models.Organization, error) {
	reqBody := models.PostBodyOrganization{
		Data: &organization,
	}

	req, err := c.Client.newRequest(http.MethodPost, "organization", reqBody)
	if err != nil {
		return nil, err
	}
	var organizationResp models.PostBodyOrganization
	_, err = c.Client.do(req, &organizationResp)
	return organizationResp.Data, err
}

func (c *OrganizationClient) Delete(organizationID string) error {
	req, err := c.Client.newRequest(http.MethodDelete, fmt.Sprintf("organization/%v", organizationID), nil)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}

func (c *OrganizationClient) Update(organization models.Organization) error {
	reqBody := models.PostBodyOrganization{
		Data: &organization,
	}

	req, err := c.Client.newRequest(http.MethodPatch, fmt.Sprintf("organization/%v", organization.ID), reqBody)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}
