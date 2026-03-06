package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type PipelineListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreatePipelineRequest struct {
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type UpdatePipelineRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	Type        *string  `json:"type,omitempty"`
	StageOrder  []string `json:"stage_order,omitempty"`
	ProjectID   *string  `json:"project_id,omitempty"`
}

type CreatePipelineStageRequest struct {
	DisplayName  string  `json:"display_name"`
	InsertBefore int     `json:"insert_before"`
	Branch       *string `json:"branch,omitempty"`
}

type AddPipelineStageAppRequest struct {
	ApplicationID string `json:"app_id"`
}

type UpdatePipelinePreviewRequest struct{}

func (c *SevallaClient) ListPipelines(ctx context.Context) ([]PipelineListItem, error) {
	return Paginate[PipelineListItem](ctx, c, "/pipelines")
}

func (c *SevallaClient) GetPipeline(ctx context.Context, id string) (*Pipeline, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/pipelines/%s", id), nil)
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

	var pipeline Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, fmt.Errorf("decoding pipeline response: %w", err)
	}

	return &pipeline, nil
}

func (c *SevallaClient) CreatePipeline(ctx context.Context, input *CreatePipelineRequest) (*Pipeline, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create pipeline request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/pipelines"), bytes.NewReader(body))
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

	var pipeline Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, fmt.Errorf("decoding pipeline response: %w", err)
	}

	return &pipeline, nil
}

func (c *SevallaClient) UpdatePipeline(ctx context.Context, id string, input *UpdatePipelineRequest) (*Pipeline, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update pipeline request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/pipelines/%s", id), bytes.NewReader(body))
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

	var pipeline Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, fmt.Errorf("decoding pipeline response: %w", err)
	}

	return &pipeline, nil
}

func (c *SevallaClient) DeletePipeline(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/pipelines/%s", id), nil)
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

func (c *SevallaClient) CreatePipelineStage(ctx context.Context, pipelineID string, input *CreatePipelineStageRequest) (*PipelineStage, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create pipeline stage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/pipelines/%s/stages", pipelineID), bytes.NewReader(body))
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

	var stage PipelineStage
	if err := json.NewDecoder(resp.Body).Decode(&stage); err != nil {
		return nil, fmt.Errorf("decoding pipeline stage response: %w", err)
	}

	return &stage, nil
}

func (c *SevallaClient) DeletePipelineStage(ctx context.Context, pipelineID string, stageID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/pipelines/%s/stages/%s", pipelineID, stageID), nil)
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

func (c *SevallaClient) AddPipelineStageApp(ctx context.Context, pipelineID string, stageID string, input *AddPipelineStageAppRequest) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshaling add pipeline stage app request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/pipelines/%s/stages/%s/apps", pipelineID, stageID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp)
	}

	return nil
}

func (c *SevallaClient) RemovePipelineStageApp(ctx context.Context, pipelineID string, stageID string, appID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/pipelines/%s/stages/%s/apps/%s", pipelineID, stageID, appID), nil)
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

func (c *SevallaClient) EnablePipelinePreview(ctx context.Context, pipelineID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/pipelines/%s/preview/enable", pipelineID), nil)
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

func (c *SevallaClient) DisablePipelinePreview(ctx context.Context, pipelineID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/pipelines/%s/preview/disable", pipelineID), nil)
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

func (c *SevallaClient) UpdatePipelinePreview(ctx context.Context, pipelineID string, input *UpdatePipelinePreviewRequest) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshaling update pipeline preview request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url("/pipelines/%s/preview", pipelineID), bytes.NewReader(body))
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
