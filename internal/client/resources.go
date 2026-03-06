package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *SevallaClient) ListClusters(ctx context.Context) ([]Cluster, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/resources/clusters"), nil)
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

	var clusters []Cluster
	if err := json.NewDecoder(resp.Body).Decode(&clusters); err != nil {
		return nil, fmt.Errorf("decoding clusters response: %w", err)
	}

	return clusters, nil
}

func (c *SevallaClient) ListProcessResourceTypes(ctx context.Context) ([]ProcessResourceType, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/resources/process-resource-types"), nil)
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

	var types []ProcessResourceType
	if err := json.NewDecoder(resp.Body).Decode(&types); err != nil {
		return nil, fmt.Errorf("decoding process resource types response: %w", err)
	}

	return types, nil
}

func (c *SevallaClient) ListAPIKeyPermissions(ctx context.Context) ([]APIKeyPermission, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/resources/rbac/api-key-permissions"), nil)
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

	var permissions []APIKeyPermission
	if err := json.NewDecoder(resp.Body).Decode(&permissions); err != nil {
		return nil, fmt.Errorf("decoding api key permissions response: %w", err)
	}

	return permissions, nil
}

func (c *SevallaClient) ListAPIKeyRoles(ctx context.Context) ([]APIKeyRole, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/resources/rbac/api-key-roles"), nil)
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

	var roles []APIKeyRole
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("decoding api key roles response: %w", err)
	}

	return roles, nil
}

func (c *SevallaClient) ListDatabaseResourceTypes(ctx context.Context) ([]DatabaseResourceType, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/resources/database-resource-types"), nil)
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

	var types []DatabaseResourceType
	if err := json.NewDecoder(resp.Body).Decode(&types); err != nil {
		return nil, fmt.Errorf("decoding database resource types response: %w", err)
	}

	return types, nil
}
