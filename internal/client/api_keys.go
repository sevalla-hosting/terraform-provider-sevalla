package client

import (
	"context"
	"net/http"
)

type APIKey struct {
	ID                string             `json:"id"`
	CompanyID         *string            `json:"company_id"`
	Name              string             `json:"name"`
	Enabled           bool               `json:"is_enabled"`
	Source            *string            `json:"source"`
	MaskedToken       *string            `json:"masked_token"`
	Roles             []APIKeyRoleRef    `json:"roles"`
	Capabilities      []APIKeyCapability `json:"capabilities"`
	LastUsedAt        *string            `json:"last_used_at"`
	ExpiresAt         *string            `json:"expired_at"`
	OldTokenExpiredAt *string            `json:"old_token_expired_at"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}

type APIKeyRoleRef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type APIKeyCapability struct {
	ID         string  `json:"id,omitempty"`
	Permission string  `json:"permission"`
	ResourceID *string `json:"resource_id,omitempty"`
}

// CreatedAPIKey is the create response; it is the only time the token is returned.
type CreatedAPIKey struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Name  string `json:"name"`
}

type CreateAPIKeyRequest struct {
	Name         string             `json:"name"`
	ExpiresAt    *string            `json:"expired_at,omitempty"`
	Capabilities []APIKeyCapability `json:"capabilities,omitempty"`
	RoleIDs      []string           `json:"role_ids,omitempty"`
}

type UpdateAPIKeyRequest struct {
	Name *string `json:"name,omitempty"`
	// Each list replaces the existing one; a non-nil empty slice clears it.
	Capabilities []APIKeyCapability `json:"capabilities,omitzero"`
	RoleIDs      []string           `json:"role_ids,omitzero"`
}

func (c *SevallaClient) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	var out APIKey
	if err := c.do(ctx, http.MethodGet, path("/api-keys/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateAPIKey(ctx context.Context, input *CreateAPIKeyRequest) (*CreatedAPIKey, error) {
	var out CreatedAPIKey
	if err := c.do(ctx, http.MethodPost, "/api-keys", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateAPIKey(ctx context.Context, id string, input *UpdateAPIKeyRequest) (*APIKey, error) {
	var out APIKey
	if err := c.do(ctx, http.MethodPatch, path("/api-keys/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteAPIKey(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/api-keys/%s", id), nil, nil)
}

// ToggleAPIKey flips the enabled state of the API key.
func (c *SevallaClient) ToggleAPIKey(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, path("/api-keys/%s/toggle", id), nil, nil)
}
