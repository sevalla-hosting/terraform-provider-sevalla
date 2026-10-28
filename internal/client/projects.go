package client

import (
	"context"
	"net/http"
	"net/url"
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

func (c *SevallaClient) GetProject(ctx context.Context, id string) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodGet, path("/projects/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateProject(ctx context.Context, input *CreateProjectRequest) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodPost, "/projects", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateProject(ctx context.Context, id string, input *UpdateProjectRequest) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodPatch, path("/projects/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteProject(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/projects/%s", id), nil, nil)
}

// AddProjectService adds a service to the project. Adding a pipeline also
// moves the applications in its stages into the project.
func (c *SevallaClient) AddProjectService(ctx context.Context, projectID string, input *AddProjectServiceRequest) error {
	return c.do(ctx, http.MethodPost, path("/projects/%s/services", projectID), input, nil)
}

func (c *SevallaClient) RemoveProjectService(ctx context.Context, projectID string, serviceID string, serviceType string) error {
	return c.do(ctx, http.MethodDelete, path("/projects/%s/services/%s", projectID, serviceID)+"?service_type="+url.QueryEscape(serviceType), nil, nil)
}
