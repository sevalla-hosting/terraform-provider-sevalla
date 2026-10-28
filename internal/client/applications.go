package client

import (
	"context"
	"net/http"
)

func (c *SevallaClient) ListApplications(ctx context.Context) ([]ApplicationListItem, error) {
	return Paginate[ApplicationListItem](ctx, c, "/applications")
}

func (c *SevallaClient) GetApplication(ctx context.Context, id string) (*Application, error) {
	var app Application
	if err := c.do(ctx, http.MethodGet, path("/applications/%s", id), nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *SevallaClient) CreateApplication(ctx context.Context, input *CreateApplicationRequest) (*Application, error) {
	var app Application
	if err := c.do(ctx, http.MethodPost, "/applications", input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApplication returns the API response, which omits cluster_id,
// buildpacks and docker_registry_credential_id; use GetApplication for state.
func (c *SevallaClient) UpdateApplication(ctx context.Context, id string, input *UpdateApplicationRequest) (*Application, error) {
	var app Application
	if err := c.do(ctx, http.MethodPatch, path("/applications/%s", id), input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *SevallaClient) DeleteApplication(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/%s", id), nil, nil)
}

func (c *SevallaClient) TriggerDeployment(ctx context.Context, appID string, input *TriggerDeploymentRequest) (*Deployment, error) {
	var deployment Deployment
	if err := c.do(ctx, http.MethodPost, path("/applications/%s/deployments", appID), input, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (c *SevallaClient) GetDeployment(ctx context.Context, appID, deploymentID string) (*Deployment, error) {
	var deployment Deployment
	if err := c.do(ctx, http.MethodGet, path("/applications/%s/deployments/%s", appID, deploymentID), nil, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

type toggleResponse struct {
	IsTurnedOn bool `json:"is_turned_on"`
}

// ToggleApplicationCDN flips the CDN setting and returns the new state.
func (c *SevallaClient) ToggleApplicationCDN(ctx context.Context, id string) (bool, error) {
	var out toggleResponse
	err := c.do(ctx, http.MethodPost, path("/applications/%s/cdn/toggle", id), nil, &out)
	return out.IsTurnedOn, err
}

// ToggleApplicationEdgeCache flips the edge cache setting and returns the new state.
func (c *SevallaClient) ToggleApplicationEdgeCache(ctx context.Context, id string) (bool, error) {
	var out toggleResponse
	err := c.do(ctx, http.MethodPost, path("/applications/%s/edge-cache/toggle", id), nil, &out)
	return out.IsTurnedOn, err
}

func (c *SevallaClient) GetApplicationIPRestriction(ctx context.Context, appID string) (*IPRestriction, error) {
	var out IPRestriction
	if err := c.do(ctx, http.MethodGet, path("/applications/%s/ip-restriction", appID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateApplicationIPRestriction(ctx context.Context, appID string, input *IPRestriction) (*IPRestriction, error) {
	var out IPRestriction
	if err := c.do(ctx, http.MethodPut, path("/applications/%s/ip-restriction", appID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type DeploymentHook struct {
	URL string `json:"url"`
}

func (c *SevallaClient) EnableDeploymentHook(ctx context.Context, appID string) (*DeploymentHook, error) {
	var out DeploymentHook
	if err := c.do(ctx, http.MethodPost, path("/applications/%s/deployment-hook", appID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) GetDeploymentHook(ctx context.Context, appID string) (*DeploymentHook, error) {
	var out DeploymentHook
	if err := c.do(ctx, http.MethodGet, path("/applications/%s/deployment-hook", appID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DisableDeploymentHook(ctx context.Context, appID string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/%s/deployment-hook", appID), nil, nil)
}

type TCPProxy struct {
	ID           string `json:"id"`
	AppID        string `json:"app_id"`
	ProcessID    string `json:"process_id"`
	Port         int64  `json:"port"`
	ExternalPort int64  `json:"external_port"`
	Hostname     string `json:"hostname"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type PrivatePort struct {
	ID        string `json:"id"`
	AppID     string `json:"app_id"`
	ProcessID string `json:"process_id"`
	Port      int64  `json:"port"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreatePortRequest struct {
	ProcessID string `json:"process_id"`
	Port      int64  `json:"port"`
}

func (c *SevallaClient) ListTCPProxies(ctx context.Context, appID string) ([]TCPProxy, error) {
	return Paginate[TCPProxy](ctx, c, path("/applications/%s/tcp-proxies", appID))
}

func (c *SevallaClient) CreateTCPProxy(ctx context.Context, appID string, input *CreatePortRequest) (*TCPProxy, error) {
	var out TCPProxy
	if err := c.do(ctx, http.MethodPost, path("/applications/%s/tcp-proxies", appID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteTCPProxy(ctx context.Context, appID, proxyID string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/%s/tcp-proxies/%s", appID, proxyID), nil, nil)
}

func (c *SevallaClient) ListPrivatePorts(ctx context.Context, appID string) ([]PrivatePort, error) {
	return Paginate[PrivatePort](ctx, c, path("/applications/%s/private-ports", appID))
}

func (c *SevallaClient) CreatePrivatePort(ctx context.Context, appID string, input *CreatePortRequest) (*PrivatePort, error) {
	var out PrivatePort
	if err := c.do(ctx, http.MethodPost, path("/applications/%s/private-ports", appID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeletePrivatePort(ctx context.Context, appID, portID string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/%s/private-ports/%s", appID, portID), nil, nil)
}
