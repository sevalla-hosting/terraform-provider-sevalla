package client

import (
	"context"
	"net/http"
)

type DockerRegistry struct {
	ID        string  `json:"id"`
	CompanyID *string `json:"company_id"`
	Name      string  `json:"name"`
	Registry  *string `json:"registry"`
	Username  *string `json:"username"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type CreateDockerRegistryRequest struct {
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Secret   string  `json:"secret"`
	Registry *string `json:"registry,omitempty"`
}

type UpdateDockerRegistryRequest struct {
	Name     *string `json:"name,omitempty"`
	Registry *string `json:"registry,omitempty"`
	Username *string `json:"username,omitempty"`
	Secret   *string `json:"secret,omitempty"`
}

func (c *SevallaClient) GetDockerRegistry(ctx context.Context, id string) (*DockerRegistry, error) {
	var out DockerRegistry
	if err := c.do(ctx, http.MethodGet, path("/docker-registries/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateDockerRegistry(ctx context.Context, input *CreateDockerRegistryRequest) (*DockerRegistry, error) {
	var out DockerRegistry
	if err := c.do(ctx, http.MethodPost, "/docker-registries", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateDockerRegistry(ctx context.Context, id string, input *UpdateDockerRegistryRequest) (*DockerRegistry, error) {
	var out DockerRegistry
	if err := c.do(ctx, http.MethodPatch, path("/docker-registries/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteDockerRegistry(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/docker-registries/%s", id), nil, nil)
}
