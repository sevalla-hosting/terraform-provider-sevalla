package client

import (
	"context"
	"net/http"
)

type StaticSite struct {
	ID                 string   `json:"id"`
	CompanyID          *string  `json:"company_id"`
	ProjectID          *string  `json:"project_id"`
	Name               string   `json:"name"`
	DisplayName        string   `json:"display_name"`
	Status             *string  `json:"status"`
	Source             string   `json:"source"`
	RepoURL            *string  `json:"repo_url"`
	DefaultBranch      *string  `json:"default_branch"`
	AutoDeploy         bool     `json:"auto_deploy"`
	IsPreviewEnabled   bool     `json:"is_preview_enabled"`
	GitLFSEnabled      bool     `json:"git_lfs_enabled"`
	GitType            *string  `json:"git_type"`
	Hostname           *string  `json:"hostname"`
	InstallCommand     *string  `json:"install_command"`
	BuildCommand       *string  `json:"build_command"`
	PublishedDirectory *string  `json:"published_directory"`
	RootDirectory      *string  `json:"root_directory"`
	NodeVersion        *string  `json:"node_version"`
	IndexFile          *string  `json:"index_file"`
	ErrorFile          *string  `json:"error_file"`
	AllowDeployPaths   []string `json:"allow_deploy_paths"`
	IgnoreDeployPaths  []string `json:"ignore_deploy_paths"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type CreateStaticSiteRequest struct {
	DisplayName        string   `json:"display_name"`
	RepoURL            string   `json:"repo_url"`
	DefaultBranch      string   `json:"default_branch"`
	Source             *string  `json:"source,omitempty"`
	GitType            *string  `json:"git_type,omitempty"`
	AutoDeploy         *bool    `json:"auto_deploy,omitempty"`
	IsPreviewEnabled   *bool    `json:"is_preview_enabled,omitempty"`
	InstallCommand     *string  `json:"install_command,omitempty"`
	BuildCommand       *string  `json:"build_command,omitempty"`
	PublishedDirectory *string  `json:"published_directory,omitempty"`
	RootDirectory      *string  `json:"root_directory,omitempty"`
	NodeVersion        *string  `json:"node_version,omitempty"`
	IndexFile          *string  `json:"index_file,omitempty"`
	ErrorFile          *string  `json:"error_file,omitempty"`
	AllowDeployPaths   []string `json:"allow_deploy_paths,omitempty"`
	IgnoreDeployPaths  []string `json:"ignore_deploy_paths,omitempty"`
	ProjectID          *string  `json:"project_id,omitempty"`
}

type UpdateStaticSiteRequest struct {
	DisplayName        *string          `json:"display_name,omitempty"`
	AutoDeploy         *bool            `json:"auto_deploy,omitempty"`
	DefaultBranch      *string          `json:"default_branch,omitempty"`
	BuildCommand       Nullable[string] `json:"build_command,omitzero"`
	NodeVersion        Nullable[string] `json:"node_version,omitzero"`
	PublishedDirectory Nullable[string] `json:"published_directory,omitzero"`
	IsPreviewEnabled   *bool            `json:"is_preview_enabled,omitempty"`
	GitLFSEnabled      *bool            `json:"git_lfs_enabled,omitempty"`
	Source             *string          `json:"source,omitempty"`
	GitType            Nullable[string] `json:"git_type,omitzero"`
	RepoURL            *string          `json:"repo_url,omitempty"`
	InstallCommand     Nullable[string] `json:"install_command,omitzero"`
	RootDirectory      Nullable[string] `json:"root_directory,omitzero"`
	IndexFile          *string          `json:"index_file,omitempty"`
	ErrorFile          Nullable[string] `json:"error_file,omitzero"`
	// A non-nil empty slice is sent as [] to clear the paths.
	AllowDeployPaths  []string `json:"allow_deploy_paths,omitzero"`
	IgnoreDeployPaths []string `json:"ignore_deploy_paths,omitzero"`
}

type StaticSiteListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Status      *string `json:"status"`
}

func (c *SevallaClient) ListStaticSites(ctx context.Context) ([]StaticSiteListItem, error) {
	return Paginate[StaticSiteListItem](ctx, c, "/static-sites")
}

func (c *SevallaClient) GetStaticSite(ctx context.Context, id string) (*StaticSite, error) {
	var out StaticSite
	if err := c.do(ctx, http.MethodGet, path("/static-sites/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateStaticSite(ctx context.Context, input *CreateStaticSiteRequest) (*StaticSite, error) {
	var out StaticSite
	if err := c.do(ctx, http.MethodPost, "/static-sites", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateStaticSite returns a partial static site; use GetStaticSite for state.
func (c *SevallaClient) UpdateStaticSite(ctx context.Context, id string, input *UpdateStaticSiteRequest) (*StaticSite, error) {
	var out StaticSite
	if err := c.do(ctx, http.MethodPatch, path("/static-sites/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteStaticSite(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/static-sites/%s", id), nil, nil)
}
