package client

import (
	"context"
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
	var out GlobalEnvironmentVariable
	if err := c.do(ctx, http.MethodPost, "/applications/global-env-vars", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateGlobalEnvVar(ctx context.Context, id string, input *UpdateGlobalEnvVarRequest) (*GlobalEnvironmentVariable, error) {
	var out GlobalEnvironmentVariable
	if err := c.do(ctx, http.MethodPut, path("/applications/global-env-vars/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteGlobalEnvVar(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/global-env-vars/%s", id), nil, nil)
}
