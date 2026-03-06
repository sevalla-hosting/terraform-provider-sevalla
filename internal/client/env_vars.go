package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EnvironmentVariable struct {
	ID           string  `json:"id"`
	Key          string  `json:"key"`
	Value        string  `json:"value"`
	IsRuntime    *bool   `json:"is_runtime"`
	IsBuildtime  *bool   `json:"is_buildtime"`
	IsProduction *bool   `json:"is_production"`
	IsPreview    *bool   `json:"is_preview"`
	Branch       *string `json:"branch"`
	CreatedAt    *string `json:"created_at"`
	UpdatedAt    *string `json:"updated_at"`
}

type CreateEnvironmentVariableRequest struct {
	Key          string  `json:"key"`
	Value        string  `json:"value"`
	IsRuntime    *bool   `json:"is_runtime,omitempty"`
	IsBuildtime  *bool   `json:"is_buildtime,omitempty"`
	IsProduction *bool   `json:"is_production,omitempty"`
	IsPreview    *bool   `json:"is_preview,omitempty"`
	Branch       *string `json:"branch,omitempty"`
}

type UpdateEnvironmentVariableRequest struct {
	Key          *string `json:"key,omitempty"`
	Value        *string `json:"value,omitempty"`
	IsRuntime    *bool   `json:"is_runtime,omitempty"`
	IsBuildtime  *bool   `json:"is_buildtime,omitempty"`
	IsProduction *bool   `json:"is_production,omitempty"`
	IsPreview    *bool   `json:"is_preview,omitempty"`
	Branch       *string `json:"branch,omitempty"`
}

func (c *SevallaClient) ListEnvironmentVariables(ctx context.Context, serviceType string, serviceID string) ([]EnvironmentVariable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("%s/%s/env-vars", serviceType, serviceID), nil)
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

	// The API may return a bare array or a wrapped object {data: [...]}.
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding environment variables response: %w", err)
	}

	var envVars []EnvironmentVariable
	if err := json.Unmarshal(raw, &envVars); err != nil {
		// Try wrapped format: {"data": [...]}
		var wrapper struct {
			Data []EnvironmentVariable `json:"data"`
		}
		if err2 := json.Unmarshal(raw, &wrapper); err2 != nil {
			return nil, fmt.Errorf("decoding environment variables response: %w", err)
		}
		envVars = wrapper.Data
	}

	return envVars, nil
}

func (c *SevallaClient) CreateEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, input *CreateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create environment variable request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("%s/%s/env-vars", serviceType, serviceID), bytes.NewReader(body))
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

	var envVar EnvironmentVariable
	if err := json.NewDecoder(resp.Body).Decode(&envVar); err != nil {
		return nil, fmt.Errorf("decoding environment variable response: %w", err)
	}

	return &envVar, nil
}

func (c *SevallaClient) UpdateEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, envVarID string, input *UpdateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update environment variable request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url("%s/%s/env-vars/%s", serviceType, serviceID, envVarID), bytes.NewReader(body))
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

	var envVar EnvironmentVariable
	if err := json.NewDecoder(resp.Body).Decode(&envVar); err != nil {
		return nil, fmt.Errorf("decoding environment variable response: %w", err)
	}

	return &envVar, nil
}

func (c *SevallaClient) DeleteEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, envVarID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("%s/%s/env-vars/%s", serviceType, serviceID, envVarID), nil)
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
