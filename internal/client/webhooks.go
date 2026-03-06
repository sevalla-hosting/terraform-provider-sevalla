package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type WebhookListItem struct {
	ID            string   `json:"id"`
	Endpoint      string   `json:"endpoint"`
	AllowedEvents []string `json:"allowed_events"`
	IsEnabled     bool     `json:"is_enabled"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type CreateWebhookRequest struct {
	AllowedEvents []string `json:"allowed_events"`
	Endpoint      string   `json:"endpoint"`
	Description   *string  `json:"description,omitempty"`
}

type UpdateWebhookRequest struct {
	AllowedEvents []string `json:"allowed_events,omitempty"`
	Endpoint      *string  `json:"endpoint,omitempty"`
	Description   *string  `json:"description,omitempty"`
}

func (c *SevallaClient) ListWebhooks(ctx context.Context) ([]WebhookListItem, error) {
	return Paginate[WebhookListItem](ctx, c, "/webhooks")
}

func (c *SevallaClient) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/webhooks/%s", id), nil)
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

	var webhook Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("decoding webhook response: %w", err)
	}

	return &webhook, nil
}

func (c *SevallaClient) CreateWebhook(ctx context.Context, input *CreateWebhookRequest) (*Webhook, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create webhook request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/webhooks"), bytes.NewReader(body))
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

	var webhook Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("decoding webhook response: %w", err)
	}

	return &webhook, nil
}

func (c *SevallaClient) UpdateWebhook(ctx context.Context, id string, input *UpdateWebhookRequest) (*Webhook, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update webhook request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/webhooks/%s", id), bytes.NewReader(body))
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

	var webhook Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("decoding webhook response: %w", err)
	}

	return &webhook, nil
}

func (c *SevallaClient) DeleteWebhook(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/webhooks/%s", id), nil)
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

func (c *SevallaClient) ToggleWebhook(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/webhooks/%s/toggle", id), nil)
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

func (c *SevallaClient) RollWebhookSecret(ctx context.Context, id string) (*Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/webhooks/%s/roll-secret", id), nil)
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

	var webhook Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("decoding webhook response: %w", err)
	}

	return &webhook, nil
}
