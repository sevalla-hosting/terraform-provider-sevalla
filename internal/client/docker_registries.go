package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type DockerRegistryListItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Registry  *string `json:"registry"`
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

func (c *SevallaClient) ListDockerRegistries(ctx context.Context) ([]DockerRegistryListItem, error) {
	return Paginate[DockerRegistryListItem](ctx, c, "/docker-registries")
}

func (c *SevallaClient) GetDockerRegistry(ctx context.Context, id string) (*DockerRegistry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/docker-registries/%s", id), nil)
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

	var registry DockerRegistry
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, fmt.Errorf("decoding docker registry response: %w", err)
	}

	return &registry, nil
}

func (c *SevallaClient) CreateDockerRegistry(ctx context.Context, input *CreateDockerRegistryRequest) (*DockerRegistry, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create docker registry request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/docker-registries"), bytes.NewReader(body))
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

	var registry DockerRegistry
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, fmt.Errorf("decoding docker registry response: %w", err)
	}

	return &registry, nil
}

func (c *SevallaClient) UpdateDockerRegistry(ctx context.Context, id string, input *UpdateDockerRegistryRequest) (*DockerRegistry, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update docker registry request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/docker-registries/%s", id), bytes.NewReader(body))
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

	var registry DockerRegistry
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, fmt.Errorf("decoding docker registry response: %w", err)
	}

	return &registry, nil
}

func (c *SevallaClient) DeleteDockerRegistry(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/docker-registries/%s", id), nil)
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
