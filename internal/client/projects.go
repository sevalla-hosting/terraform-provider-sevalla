package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Project struct {
	ID          string           `json:"id"`
	CompanyID   *string          `json:"company_id"`
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Services    []ProjectService `json:"services"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

type ProjectListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateProjectRequest struct {
	DisplayName string `json:"display_name"`
}

type UpdateProjectRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
}

type ProjectService struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type AddProjectServiceRequest struct {
	ServiceID   string `json:"service_id"`
	ServiceType string `json:"service_type"`
}

func (c *SevallaClient) ListProjects(ctx context.Context) ([]ProjectListItem, error) {
	return Paginate[ProjectListItem](ctx, c, "/projects")
}

func (c *SevallaClient) GetProject(ctx context.Context, id string) (*Project, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/projects/%s", id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("decoding project response: %w", err)
	}

	return &project, nil
}

func (c *SevallaClient) CreateProject(ctx context.Context, input *CreateProjectRequest) (*Project, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/projects"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("decoding project response: %w", err)
	}

	return &project, nil
}

func (c *SevallaClient) UpdateProject(ctx context.Context, id string, input *UpdateProjectRequest) (*Project, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/projects/%s", id), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("decoding project response: %w", err)
	}

	return &project, nil
}

func (c *SevallaClient) DeleteProject(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/projects/%s", id), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp)
	}

	return nil
}

func (c *SevallaClient) AddProjectService(ctx context.Context, projectID string, input *AddProjectServiceRequest) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshaling add project service request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/projects/%s/services", projectID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	return nil
}


func (c *SevallaClient) RemoveProjectService(ctx context.Context, projectID string, serviceID string, serviceType string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/projects/%s/services/%s", projectID, serviceID), nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Set("service_type", serviceType)
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp)
	}

	return nil
}
