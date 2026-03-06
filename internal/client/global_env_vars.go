package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GlobalEnvironmentVariable struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsRuntime   *bool  `json:"is_runtime"`
	IsBuildtime *bool  `json:"is_buildtime"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateGlobalEnvVarRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsRuntime   *bool  `json:"is_runtime,omitempty"`
	IsBuildtime *bool  `json:"is_buildtime,omitempty"`
}

type UpdateGlobalEnvVarRequest struct {
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
	IsBuildtime *bool   `json:"is_buildtime,omitempty"`
}

func (c *SevallaClient) ListGlobalEnvVars(ctx context.Context) ([]GlobalEnvironmentVariable, error) {
	return Paginate[GlobalEnvironmentVariable](ctx, c, "/applications/global-env-vars")
}

func (c *SevallaClient) CreateGlobalEnvVar(ctx context.Context, input *CreateGlobalEnvVarRequest) (*GlobalEnvironmentVariable, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create global env var request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/applications/global-env-vars"), bytes.NewReader(body))
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

	var envVar GlobalEnvironmentVariable
	if err := json.NewDecoder(resp.Body).Decode(&envVar); err != nil {
		return nil, fmt.Errorf("decoding global env var response: %w", err)
	}

	return &envVar, nil
}

func (c *SevallaClient) UpdateGlobalEnvVar(ctx context.Context, id string, input *UpdateGlobalEnvVarRequest) (*GlobalEnvironmentVariable, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update global env var request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url("/applications/global-env-vars/%s", id), bytes.NewReader(body))
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

	var envVar GlobalEnvironmentVariable
	if err := json.NewDecoder(resp.Body).Decode(&envVar); err != nil {
		return nil, fmt.Errorf("decoding global env var response: %w", err)
	}

	return &envVar, nil
}

func (c *SevallaClient) DeleteGlobalEnvVar(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/applications/global-env-vars/%s", id), nil)
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
