package client

import (
	"context"
	"net/http"
)

func (c *SevallaClient) ListClusters(ctx context.Context) ([]Cluster, error) {
	var out []Cluster
	err := c.do(ctx, http.MethodGet, "/resources/clusters", nil, &out)
	return out, err
}

func (c *SevallaClient) ListProcessResourceTypes(ctx context.Context) ([]ProcessResourceType, error) {
	var out []ProcessResourceType
	err := c.do(ctx, http.MethodGet, "/resources/process-resource-types", nil, &out)
	return out, err
}

func (c *SevallaClient) ListAPIKeyPermissions(ctx context.Context) ([]APIKeyPermission, error) {
	var out []APIKeyPermission
	err := c.do(ctx, http.MethodGet, "/resources/rbac/api-key-permissions", nil, &out)
	return out, err
}

func (c *SevallaClient) ListAPIKeyRoles(ctx context.Context) ([]APIKeyRole, error) {
	var out []APIKeyRole
	err := c.do(ctx, http.MethodGet, "/resources/rbac/api-key-roles", nil, &out)
	return out, err
}

func (c *SevallaClient) ListDatabaseResourceTypes(ctx context.Context) ([]DatabaseResourceType, error) {
	var out []DatabaseResourceType
	err := c.do(ctx, http.MethodGet, "/resources/database-resource-types", nil, &out)
	return out, err
}
