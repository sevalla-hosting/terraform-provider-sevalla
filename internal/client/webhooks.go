package client

import (
	"context"
	"net/http"
)

type Webhook struct {
	ID                 string   `json:"id"`
	CompanyID          *string  `json:"company_id"`
	Endpoint           string   `json:"endpoint"`
	AllowedEvents      []string `json:"allowed_events"`
	IsEnabled          bool     `json:"is_enabled"`
	Secret             *string  `json:"secret"`
	Description        *string  `json:"description"`
	OldSecret          *string  `json:"old_secret"`
	OldSecretExpiredAt *string  `json:"old_secret_expired_at"`
	CreatedBy          *string  `json:"created_by"`
	UpdatedBy          *string  `json:"updated_by"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type CreateWebhookRequest struct {
	AllowedEvents []string `json:"allowed_events"`
	Endpoint      string   `json:"endpoint"`
	Description   *string  `json:"description,omitempty"`
}

type UpdateWebhookRequest struct {
	AllowedEvents []string         `json:"allowed_events,omitempty"`
	Endpoint      *string          `json:"endpoint,omitempty"`
	Description   Nullable[string] `json:"description,omitzero"`
}

func (c *SevallaClient) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	var out Webhook
	if err := c.do(ctx, http.MethodGet, path("/webhooks/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateWebhook(ctx context.Context, input *CreateWebhookRequest) (*Webhook, error) {
	var out Webhook
	if err := c.do(ctx, http.MethodPost, "/webhooks", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateWebhook(ctx context.Context, id string, input *UpdateWebhookRequest) (*Webhook, error) {
	var out Webhook
	if err := c.do(ctx, http.MethodPatch, path("/webhooks/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteWebhook(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/webhooks/%s", id), nil, nil)
}

// ToggleWebhook flips the enabled state of the webhook and returns it.
func (c *SevallaClient) ToggleWebhook(ctx context.Context, id string) (*Webhook, error) {
	var out Webhook
	if err := c.do(ctx, http.MethodPost, path("/webhooks/%s/toggle", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
