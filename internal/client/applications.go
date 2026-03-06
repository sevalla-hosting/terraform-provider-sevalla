package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *SevallaClient) ListApplications(ctx context.Context) ([]ApplicationListItem, error) {
	return Paginate[ApplicationListItem](ctx, c, "/applications")
}

func (c *SevallaClient) GetApplication(ctx context.Context, id string) (*Application, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/applications/%s", id), nil)
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

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decoding application response: %w", err)
	}

	return &app, nil
}

func (c *SevallaClient) CreateApplication(ctx context.Context, input *CreateApplicationRequest) (*Application, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create application request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications"), bytes.NewReader(body))
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

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decoding application response: %w", err)
	}

	return &app, nil
}

func (c *SevallaClient) UpdateApplication(ctx context.Context, id string, input *UpdateApplicationRequest) (*Application, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update application request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/applications/%s", id), bytes.NewReader(body))
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

	var app Application
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decoding application response: %w", err)
	}

	return &app, nil
}

func (c *SevallaClient) DeleteApplication(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/applications/%s", id), nil)
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

func (c *SevallaClient) TriggerDeployment(ctx context.Context, appID string, input *TriggerDeploymentRequest) (string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("marshaling trigger deployment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications/%s/deployments", appID), bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", parseErrorResponse(resp)
	}

	var deployment TriggerDeploymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return "", fmt.Errorf("decoding deployment response: %w", err)
	}

	return deployment.ID, nil
}

func (c *SevallaClient) SuspendApplication(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications/%s/suspend", id), nil)
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

func (c *SevallaClient) ActivateApplication(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications/%s/activate", id), nil)
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

func (c *SevallaClient) ToggleApplicationCDN(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications/%s/cdn/toggle", id), nil)
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
