package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/jsonapi"
)

type Client struct {
	BaseURL      *url.URL
	Token        string
	Organization *OrganizationClient
	Module       *ModuleClient
	Workspace    *WorkspaceClient
	Variable     *VariableClient
	Job          *JobClient
	Team         *TeamClient
	HttpClient   *http.Client
	BasePath     string
}

var defaultPath string = "/api/v1/"

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: c.BasePath + path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		jsonStr, _ := json.Marshal(body)
		buf = bytes.NewBuffer(jsonStr)
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	bearer := "Bearer " + c.Token
	req.Header.Set("Authorization", bearer)
	if body != nil {
		req.Header.Set("Content-Type", jsonapi.MediaType)
	}

	return req, nil
}

func (c *Client) newRequestWithFilter(method, path string, query string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: c.BasePath + path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		jsonStr, _ := json.Marshal(body)
		buf = bytes.NewBuffer(jsonStr)
	}

	queryWithFilter := ""

	if len(query) == 0 {
		queryWithFilter = u.String()
	} else {
		queryWithFilter = fmt.Sprintf("%s?%s", u.String(), query)
	}

	req, err := http.NewRequest(method, queryWithFilter, buf)
	if err != nil {
		return nil, err
	}
	bearer := "Bearer " + c.Token
	req.Header.Set("Authorization", bearer)
	if body != nil {
		req.Header.Set("Content-Type", jsonapi.MediaType)
	}

	return req, nil
}

// TODO: do() does not check HTTP status codes. A 401, 403, 404, or 500
// response is treated identically to a 200. When v is non-nil, the only
// failure signal is a JSON decode error (e.g., empty body on 500 causes EOF).
// When v is nil (Update/Delete operations), errors are completely silent —
// the caller gets nil error even on 500. This needs:
//   - Check resp.StatusCode and return a structured error for non-2xx responses
//   - Read and include the response body in the error for debugging
//   - Update all callers that currently ignore the returned *http.Response
//   - Update tests in client_test.go (TestDo_500WithValidJSON, TestDo_401Response,
//     TestDo_404WithJSONError) to assert errors instead of documenting silent success
//   - Add server error tests for Update/Delete on all resources (currently only
//     Organization and Variable have them)
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error or handle it appropriately
			// For now, we'll ignore it as it's a cleanup operation
			_ = closeErr
		}
	}()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}

func NewClient(httpClient *http.Client, token string, baseUrl *url.URL) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{HttpClient: httpClient}
	c.BaseURL = baseUrl
	c.Token = token
	c.Organization = &OrganizationClient{Client: c}
	c.Module = &ModuleClient{Client: c}
	c.Workspace = &WorkspaceClient{Client: c}
	c.Variable = &VariableClient{Client: c}
	c.Team = &TeamClient{Client: c}
	c.Job = &JobClient{Client: c}

	// Handle base path
	if baseUrl.Path == "" {
		c.BasePath = defaultPath
	} else {
		c.BasePath = baseUrl.Path
		if !strings.HasSuffix(c.BasePath, "/") {
			c.BasePath += "/"
		}
		c.BasePath += defaultPath
	}

	return c
}
