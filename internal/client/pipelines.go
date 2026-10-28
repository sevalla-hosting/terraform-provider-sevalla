package client

import (
	"context"
	"net/http"
)

type Pipeline struct {
	ID          string          `json:"id"`
	CompanyID   *string         `json:"company_id"`
	ProjectID   *string         `json:"project_id"`
	DisplayName string          `json:"display_name"`
	Type        string          `json:"type"`
	Stages      []PipelineStage `json:"stages"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type PipelineStage struct {
	ID              string  `json:"id"`
	PipelineID      *string `json:"pipeline_id"`
	DisplayName     string  `json:"display_name"`
	Type            string  `json:"type"`
	Order           int     `json:"order"`
	Branch          *string `json:"branch"`
	AutoCreateApp   *bool   `json:"auto_create_app"`
	DeleteStaleApps *bool   `json:"delete_stale_apps"`
	StaleAppDays    *int    `json:"stale_app_days"`
	CreatedAt       *string `json:"created_at"`
	UpdatedAt       *string `json:"updated_at"`
}

type CreatePipelineRequest struct {
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type UpdatePipelineRequest struct {
	DisplayName *string          `json:"display_name,omitempty"`
	Type        *string          `json:"type,omitempty"`
	ProjectID   Nullable[string] `json:"project_id,omitzero"`
}

type CreatePipelineStageRequest struct {
	DisplayName  string  `json:"display_name"`
	InsertBefore int     `json:"insert_before"`
	Branch       *string `json:"branch,omitempty"`
}

type AddPipelineStageAppRequest struct {
	ApplicationID string `json:"app_id"`
}

func (c *SevallaClient) GetPipeline(ctx context.Context, id string) (*Pipeline, error) {
	var out Pipeline
	if err := c.do(ctx, http.MethodGet, path("/pipelines/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreatePipeline(ctx context.Context, input *CreatePipelineRequest) (*Pipeline, error) {
	var out Pipeline
	if err := c.do(ctx, http.MethodPost, "/pipelines", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdatePipeline(ctx context.Context, id string, input *UpdatePipelineRequest) (*Pipeline, error) {
	var out Pipeline
	if err := c.do(ctx, http.MethodPatch, path("/pipelines/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeletePipeline(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/pipelines/%s", id), nil, nil)
}

func (c *SevallaClient) CreatePipelineStage(ctx context.Context, pipelineID string, input *CreatePipelineStageRequest) (*PipelineStage, error) {
	var out PipelineStage
	if err := c.do(ctx, http.MethodPost, path("/pipelines/%s/stages", pipelineID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeletePipelineStage(ctx context.Context, pipelineID string, stageID string) error {
	return c.do(ctx, http.MethodDelete, path("/pipelines/%s/stages/%s", pipelineID, stageID), nil, nil)
}

func (c *SevallaClient) AddPipelineStageApp(ctx context.Context, pipelineID string, stageID string, input *AddPipelineStageAppRequest) error {
	return c.do(ctx, http.MethodPost, path("/pipelines/%s/stages/%s/apps", pipelineID, stageID), input, nil)
}

func (c *SevallaClient) RemovePipelineStageApp(ctx context.Context, pipelineID string, stageID string, appID string) error {
	return c.do(ctx, http.MethodDelete, path("/pipelines/%s/stages/%s/apps/%s", pipelineID, stageID, appID), nil, nil)
}
