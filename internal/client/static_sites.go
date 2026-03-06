package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type StaticSite struct {
	ID                 string  `json:"id"`
	CompanyID          *string `json:"company_id"`
	ProjectID          *string `json:"project_id"`
	Name               string  `json:"name"`
	DisplayName        string  `json:"display_name"`
	Status             *string `json:"status"`
	Source             string  `json:"source"`
	RepoURL            *string `json:"repo_url"`
	DefaultBranch      *string `json:"default_branch"`
	AutoDeploy         bool    `json:"auto_deploy"`
	IsPreviewEnabled   bool    `json:"is_preview_enabled"`
	GitType            *string `json:"git_type"`
	Hostname           *string `json:"hostname"`
	InstallCommand     *string `json:"install_command"`
	BuildCommand       *string `json:"build_command"`
	PublishedDirectory *string `json:"published_directory"`
	RootDirectory      *string `json:"root_directory"`
	NodeVersion        *string `json:"node_version"`
	IndexFile          *string `json:"index_file"`
	ErrorFile          *string `json:"error_file"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type CreateStaticSiteRequest struct {
	DisplayName        string  `json:"display_name"`
	RepoURL            string  `json:"repo_url"`
	DefaultBranch      string  `json:"default_branch"`
	Source             *string `json:"source,omitempty"`
	GitType            *string `json:"git_type,omitempty"`
	AutoDeploy         *bool   `json:"auto_deploy,omitempty"`
	IsPreviewEnabled   *bool   `json:"is_preview_enabled,omitempty"`
	InstallCommand     *string `json:"install_command,omitempty"`
	BuildCommand       *string `json:"build_command,omitempty"`
	PublishedDirectory *string `json:"published_directory,omitempty"`
	RootDirectory      *string `json:"root_directory,omitempty"`
	NodeVersion        *string `json:"node_version,omitempty"`
	IndexFile          *string `json:"index_file,omitempty"`
	ErrorFile          *string `json:"error_file,omitempty"`
	ProjectID          *string `json:"project_id,omitempty"`
}

type UpdateStaticSiteRequest struct {
	DisplayName        *string `json:"display_name,omitempty"`
	AutoDeploy         *bool   `json:"auto_deploy,omitempty"`
	DefaultBranch      *string `json:"default_branch,omitempty"`
	BuildCommand       *string `json:"build_command,omitempty"`
	NodeVersion        *string `json:"node_version,omitempty"`
	PublishedDirectory *string `json:"published_directory,omitempty"`
	IsPreviewEnabled   *bool   `json:"is_preview_enabled,omitempty"`
	Source             *string `json:"source,omitempty"`
	GitType            *string `json:"git_type,omitempty"`
	RepoURL            *string `json:"repo_url,omitempty"`
	InstallCommand     *string `json:"install_command,omitempty"`
	RootDirectory      *string `json:"root_directory,omitempty"`
	IndexFile          *string `json:"index_file,omitempty"`
	ErrorFile          *string `json:"error_file,omitempty"`
}

type StaticSiteListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Status      *string `json:"status"`
	Source      string  `json:"source"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (c *SevallaClient) ListStaticSites(ctx context.Context) ([]StaticSiteListItem, error) {
	return Paginate[StaticSiteListItem](ctx, c, "/static-sites")
}

func (c *SevallaClient) GetStaticSite(ctx context.Context, id string) (*StaticSite, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/static-sites/%s", id), nil)
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

	var site StaticSite
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("decoding static site response: %w", err)
	}

	return &site, nil
}

func (c *SevallaClient) CreateStaticSite(ctx context.Context, input *CreateStaticSiteRequest) (*StaticSite, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create static site request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/static-sites"), bytes.NewReader(body))
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

	var site StaticSite
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("decoding static site response: %w", err)
	}

	return &site, nil
}

func (c *SevallaClient) UpdateStaticSite(ctx context.Context, id string, input *UpdateStaticSiteRequest) (*StaticSite, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update static site request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/static-sites/%s", id), bytes.NewReader(body))
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

	var site StaticSite
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("decoding static site response: %w", err)
	}

	return &site, nil
}

func (c *SevallaClient) DeleteStaticSite(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/static-sites/%s", id), nil)
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
