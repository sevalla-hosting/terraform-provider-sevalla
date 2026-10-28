package client

import (
	"context"
	"net/http"
)

// EnvironmentVariable covers both application variables (is_runtime,
// is_buildtime) and static site variables (is_production, is_preview, branch).
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
	// Branch is only supported by static sites; Null() clears it.
	Branch Nullable[string] `json:"branch,omitzero"`
}

// serviceType is the collection path of the owning service, e.g. "/applications".
func (c *SevallaClient) ListEnvironmentVariables(ctx context.Context, serviceType string, serviceID string) ([]EnvironmentVariable, error) {
	return Paginate[EnvironmentVariable](ctx, c, path(serviceType+"/%s/env-vars", serviceID))
}

func (c *SevallaClient) CreateEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, input *CreateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	var out EnvironmentVariable
	if err := c.do(ctx, http.MethodPost, path(serviceType+"/%s/env-vars", serviceID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, envVarID string, input *UpdateEnvironmentVariableRequest) (*EnvironmentVariable, error) {
	var out EnvironmentVariable
	if err := c.do(ctx, http.MethodPut, path(serviceType+"/%s/env-vars/%s", serviceID, envVarID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteEnvironmentVariable(ctx context.Context, serviceType string, serviceID string, envVarID string) error {
	return c.do(ctx, http.MethodDelete, path(serviceType+"/%s/env-vars/%s", serviceID, envVarID), nil, nil)
}
