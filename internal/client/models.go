package client

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
	HibernationEnabled         bool              `json:"hibernation_enabled"`
	HibernateAfterSeconds      *int64            `json:"hibernate_after_seconds"`
	AutoDeploy                 bool              `json:"auto_deploy"`
	WaitForChecks              bool              `json:"wait_for_checks"`
	IsSuspended                bool              `json:"is_suspended"`
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
	DisplayName                *string           `json:"display_name,omitempty"`
	AutoDeploy                 *bool             `json:"auto_deploy,omitempty"`
	DefaultBranch              *string           `json:"default_branch,omitempty"`
	HibernationEnabled         *bool             `json:"hibernation_enabled,omitempty"`
	HibernateAfterSeconds      *int64            `json:"hibernate_after_seconds,omitempty"`
	BuildType                  *string           `json:"build_type,omitempty"`
	BuildPath                  *string           `json:"build_path,omitempty"`
	DockerfilePath             *string           `json:"dockerfile_path,omitempty"`
	DockerContext              *string           `json:"docker_context,omitempty"`
	BuildCacheEnabled          *bool             `json:"build_cache_enabled,omitempty"`
	DockerRegistryCredentialID *string           `json:"docker_registry_credential_id,omitempty"`
	PackBuilder                *string           `json:"pack_builder,omitempty"`
	NixpacksVersion            *string           `json:"nixpacks_version,omitempty"`
	AllowDeployPaths           []string          `json:"allow_deploy_paths,omitempty"`
	IgnoreDeployPaths          []string          `json:"ignore_deploy_paths,omitempty"`
	Buildpacks                 []BuildpackConfig `json:"buildpacks,omitempty"`
	Source                     *string           `json:"source,omitempty"`
	GitType                    *string           `json:"git_type,omitempty"`
	RepoURL                    *string           `json:"repo_url,omitempty"`
	DockerImage                *string           `json:"docker_image,omitempty"`
}

type TriggerDeploymentRequest struct {
	IsRestart bool `json:"is_restart"`
}

type TriggerDeploymentResponse struct {
	ID string `json:"id"`
}

type ApplicationListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Status      *string `json:"status"`
	Type        string  `json:"type"`
	Source      string  `json:"source"`
	IsSuspended bool    `json:"is_suspended"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
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
	Category     string `json:"category"`
}

