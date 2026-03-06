package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type APIKey struct {
	ID           string             `json:"id"`
	CompanyID    *string            `json:"company_id"`
	Name         string             `json:"name"`
	Token        *string            `json:"token"`
	Enabled      bool               `json:"is_enabled"`
	ExpiresAt    *string            `json:"expired_at"`
	Capabilities []json.RawMessage  `json:"capabilities"`
	Roles        []json.RawMessage  `json:"roles"`
	Source       *string            `json:"source"`
	LastUsedAt   *string            `json:"last_used_at"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
}

type APIKeyListItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Enabled   bool    `json:"is_enabled"`
	ExpiresAt *string `json:"expired_at"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type APIKeyCapabilityRequest struct {
	Permission string  `json:"permission"`
	IDResource *string `json:"id_resource,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name         string                    `json:"name"`
	ExpiresAt    *string                   `json:"expired_at,omitempty"`
	Capabilities []APIKeyCapabilityRequest `json:"capabilities,omitempty"`
	RoleIDs      []string                  `json:"role_ids,omitempty"`
}

type UpdateAPIKeyRequest struct {
	Name         *string                   `json:"name,omitempty"`
	Capabilities []APIKeyCapabilityRequest `json:"capabilities,omitempty"`
	RoleIDs      []string                  `json:"role_ids,omitempty"`
}

func (c *SevallaClient) ListAPIKeys(ctx context.Context) ([]APIKeyListItem, error) {
	return Paginate[APIKeyListItem](ctx, c, "/api-keys")
}

func (c *SevallaClient) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/api-keys/%s", id), nil)
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

	var apiKey APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKey); err != nil {
		return nil, fmt.Errorf("decoding api key response: %w", err)
	}

	return &apiKey, nil
}

func (c *SevallaClient) CreateAPIKey(ctx context.Context, input *CreateAPIKeyRequest) (*APIKey, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create api key request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api-keys"), bytes.NewReader(body))
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

	var apiKey APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKey); err != nil {
		return nil, fmt.Errorf("decoding api key response: %w", err)
	}

	return &apiKey, nil
}

func (c *SevallaClient) UpdateAPIKey(ctx context.Context, id string, input *UpdateAPIKeyRequest) (*APIKey, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update api key request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/api-keys/%s", id), bytes.NewReader(body))
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

	var apiKey APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKey); err != nil {
		return nil, fmt.Errorf("decoding api key response: %w", err)
	}

	return &apiKey, nil
}

func (c *SevallaClient) DeleteAPIKey(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/api-keys/%s", id), nil)
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

func (c *SevallaClient) ToggleAPIKey(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api-keys/%s/toggle", id), nil)
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

func (c *SevallaClient) RotateAPIKey(ctx context.Context, id string) (*APIKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api-keys/%s/rotate", id), nil)
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

	var apiKey APIKey
	if err := json.NewDecoder(resp.Body).Decode(&apiKey); err != nil {
		return nil, fmt.Errorf("decoding api key response: %w", err)
	}

	return &apiKey, nil
}
