package client

import "encoding/json"

// Nullable is a request field that is omitted when unset (with the omitzero
// tag), and otherwise sent as its value or, when Value is nil, as JSON null.
type Nullable[T any] struct {
	Set   bool
	Value *T
}

// Null returns a Nullable that is sent as JSON null.
func Null[T any]() Nullable[T] { return Nullable[T]{Set: true} }

// NullableOf returns a Nullable that is sent as v, or as JSON null when v is nil.
func NullableOf[T any](v *T) Nullable[T] { return Nullable[T]{Set: true, Value: v} }

func (n Nullable[T]) IsZero() bool { return !n.Set }

func (n Nullable[T]) MarshalJSON() ([]byte, error) { return json.Marshal(n.Value) }

type PaginatedResponse[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type BuildpackConfig struct {
	Order  int    `json:"order"`
	Source string `json:"source"`
}

type Application struct {
	ID                         string            `json:"id"`
	CompanyID                  *string           `json:"company_id"`
	ProjectID                  *string           `json:"project_id"`
	Name                       string            `json:"name"`
	Namespace                  *string           `json:"namespace"`
	DisplayName                string            `json:"display_name"`
	Source                     string            `json:"source"`
	Type                       string            `json:"type"`
	Status                     *string           `json:"status"`
	BuildCacheEnabled          bool              `json:"build_cache_enabled"`
	GitLFSEnabled              bool              `json:"git_lfs_enabled"`
	HibernationEnabled         bool              `json:"hibernation_enabled"`
	HibernateAfterSeconds      *int64            `json:"hibernate_after_seconds"`
	AutoDeploy                 bool              `json:"auto_deploy"`
	IsSuspended                bool              `json:"is_suspended"`
	SuspensionStatus           string            `json:"suspension_status"`
	IsCDNEnabled               *bool             `json:"is_cdn_enabled"`
	IsEdgeCacheEnabled         *bool             `json:"is_edge_cache_enabled"`
	ClusterID                  *string           `json:"cluster_id"`
	GitType                    *string           `json:"git_type"`
	RepoURL                    *string           `json:"repo_url"`
	DefaultBranch              *string           `json:"default_branch"`
	DockerImage                *string           `json:"docker_image"`
	DockerRegistryCredentialID *string           `json:"docker_registry_credential_id"`
	BuildType                  string            `json:"build_type"`
	BuildPath                  *string           `json:"build_path"`
	PackBuilder                *string           `json:"pack_builder"`
	NixpacksVersion            *string           `json:"nixpacks_version"`
	DockerfilePath             *string           `json:"dockerfile_path"`
	DockerContext              *string           `json:"docker_context"`
	AllowDeployPaths           []string          `json:"allow_deploy_paths"`
	IgnoreDeployPaths          []string          `json:"ignore_deploy_paths"`
	Buildpacks                 []BuildpackConfig `json:"buildpacks"`
	CreatedBy                  *string           `json:"created_by"`
	CreatedAt                  string            `json:"created_at"`
	UpdatedAt                  string            `json:"updated_at"`
}

type CreateApplicationRequest struct {
	DisplayName                string  `json:"display_name"`
	ClusterID                  string  `json:"cluster_id"`
	Source                     string  `json:"source"`
	ProjectID                  *string `json:"project_id,omitempty"`
	GitType                    *string `json:"git_type,omitempty"`
	RepoURL                    *string `json:"repo_url,omitempty"`
	DefaultBranch              *string `json:"default_branch,omitempty"`
	DockerImage                *string `json:"docker_image,omitempty"`
	DockerRegistryCredentialID *string `json:"docker_registry_credential_id,omitempty"`
}

type UpdateApplicationRequest struct {
	DisplayName                *string `json:"display_name,omitempty"`
	AutoDeploy                 *bool   `json:"auto_deploy,omitempty"`
	DefaultBranch              *string `json:"default_branch,omitempty"`
	HibernationEnabled         *bool   `json:"hibernation_enabled,omitempty"`
	HibernateAfterSeconds      *int64  `json:"hibernate_after_seconds,omitempty"`
	BuildType                  *string `json:"build_type,omitempty"`
	BuildPath                  *string `json:"build_path,omitempty"`
	DockerfilePath             *string `json:"dockerfile_path,omitempty"`
	DockerContext              *string `json:"docker_context,omitempty"`
	BuildCacheEnabled          *bool   `json:"build_cache_enabled,omitempty"`
	GitLFSEnabled              *bool   `json:"git_lfs_enabled,omitempty"`
	DockerRegistryCredentialID *string `json:"docker_registry_credential_id,omitempty"`
	PackBuilder                *string `json:"pack_builder,omitempty"`
	NixpacksVersion            *string `json:"nixpacks_version,omitempty"`
	// A non-nil empty slice is sent as [] to clear the paths.
	AllowDeployPaths  []string          `json:"allow_deploy_paths,omitzero"`
	IgnoreDeployPaths []string          `json:"ignore_deploy_paths,omitzero"`
	Buildpacks        []BuildpackConfig `json:"buildpacks,omitempty"`
	Source            *string           `json:"source,omitempty"`
	GitType           *string           `json:"git_type,omitempty"`
	RepoURL           *string           `json:"repo_url,omitempty"`
	DockerImage       *string           `json:"docker_image,omitempty"`
}

type TriggerDeploymentRequest struct {
	Branch         *string `json:"branch,omitempty"`
	CommitSHA      *string `json:"commit_sha,omitempty"`
	DockerImage    *string `json:"docker_image,omitempty"`
	IsRestart      bool    `json:"is_restart"`
	SkipBuildCache bool    `json:"skip_build_cache,omitempty"`
}

type Deployment struct {
	ID     string  `json:"id"`
	Status *string `json:"status"`
}

type ApplicationListItem struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	DisplayName      string  `json:"display_name"`
	Status           *string `json:"status"`
	Type             string  `json:"type"`
	Source           string  `json:"source"`
	IsSuspended      bool    `json:"is_suspended"`
	SuspensionStatus string  `json:"suspension_status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type APIKeyPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

type APIKeyRole struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type Cluster struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName *string `json:"display_name"`
	Location    string  `json:"location"`
	Provider    string  `json:"provider"`
}

type ProcessResourceType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CPULimit    int64  `json:"cpu_limit"`
	MemoryLimit int64  `json:"memory_limit"`
	Category    string `json:"category"`
}

type DatabaseResourceType struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CPULimit     int64  `json:"cpu_limit"`
	MemoryLimit  int64  `json:"memory_limit"`
	StorageLimit int64  `json:"storage_limit"`
}
