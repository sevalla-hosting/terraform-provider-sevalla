package client

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Database struct {
	ID                 string  `json:"id"`
	CompanyID          *string `json:"company_id"`
	ProjectID          *string `json:"project_id"`
	Name               string  `json:"name"`
	DisplayName        string  `json:"display_name"`
	Type               string  `json:"type"`
	Version            *string `json:"version"`
	Status             *string `json:"status"`
	IsSuspended        bool    `json:"is_suspended"`
	ClusterID          string  `json:"cluster_id"`
	ClusterDisplayName *string `json:"cluster_display_name"`
	ClusterLocation    *string `json:"cluster_location"`
	ResourceTypeID     *string `json:"resource_type_id"`
	ResourceTypeName   *string `json:"resource_type_name"`
	CPULimit           *int64  `json:"cpu_limit"`
	MemoryLimit        *int64  `json:"memory_limit"`
	StorageSize        *int64  `json:"storage_size"`
	DbName             *string `json:"db_name"`
	DbUser             *string `json:"db_user"`
	InternalHostname   *string `json:"internal_hostname"`
	InternalPort       *string `json:"internal_port"`
	ExternalHostname   *string `json:"external_hostname"`
	ExternalPort       *string `json:"external_port"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type DatabaseExtensions struct {
	EnableVector  *bool `json:"enable_vector,omitempty"`
	EnablePostgis *bool `json:"enable_postgis,omitempty"`
	EnableCron    *bool `json:"enable_cron,omitempty"`
}

type CreateDatabaseRequest struct {
	DisplayName    string              `json:"display_name"`
	Type           string              `json:"type"`
	Version        string              `json:"version"`
	ClusterID      string              `json:"cluster_id"`
	ResourceTypeID string              `json:"resource_type_id"`
	DbName         string              `json:"db_name"`
	DbPassword     string              `json:"db_password"`
	DbUser         *string             `json:"db_user,omitempty"`
	ProjectID      *string             `json:"project_id,omitempty"`
	Extensions     *DatabaseExtensions `json:"extensions,omitempty"`
}

type UpdateDatabaseRequest struct {
	DisplayName    *string          `json:"display_name,omitempty"`
	ResourceTypeID *string          `json:"resource_type_id,omitempty"`
	ProjectID      Nullable[string] `json:"project_id,omitzero"`
}

type DatabaseListItem struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	DisplayName      string  `json:"display_name"`
	Type             string  `json:"type"`
	Version          string  `json:"version"`
	ResourceTypeName *string `json:"resource_type_name"`
	Status           *string `json:"status"`
	IsSuspended      bool    `json:"is_suspended"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type InternalConnection struct {
	ID                string  `json:"id"`
	SourceID          string  `json:"source_id"`
	SourceType        *string `json:"source_type"`
	SourceName        *string `json:"source_name"`
	SourceDisplayName *string `json:"source_display_name"`
	TargetID          string  `json:"target_id"`
	TargetType        *string `json:"target_type"`
	TargetName        *string `json:"target_name"`
	TargetDisplayName *string `json:"target_display_name"`
}

type CreateInternalConnectionRequest struct {
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
}

type IPRestriction struct {
	Type      string   `json:"type"`
	IsEnabled bool     `json:"is_enabled"`
	IPList    []string `json:"ip_list"`
}

// WaitForDatabaseStatus polls the database until it reaches one of the target
// statuses, and fails when it reaches error or a *Failed status instead.
func (c *SevallaClient) WaitForDatabaseStatus(ctx context.Context, id string, targetStatuses []string, maxWait time.Duration) (*Database, error) {
	var db *Database
	err := Poll(ctx, maxWait, func() (bool, error) {
		var err error
		if db, err = c.GetDatabase(ctx, id); err != nil {
			return false, err
		}
		status := ""
		if db.Status != nil {
			status = *db.Status
		}
		switch {
		case slices.Contains(targetStatuses, status):
			return true, nil
		case status == "error" || strings.HasSuffix(status, "Failed"):
			return false, fmt.Errorf("database %s is in the %s state", id, status)
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (c *SevallaClient) ListDatabases(ctx context.Context) ([]DatabaseListItem, error) {
	return Paginate[DatabaseListItem](ctx, c, "/databases")
}

func (c *SevallaClient) GetDatabase(ctx context.Context, id string) (*Database, error) {
	var db Database
	if err := c.do(ctx, http.MethodGet, path("/databases/%s", id), nil, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

// CreateDatabase returns a partial database; poll GetDatabase for the full state.
func (c *SevallaClient) CreateDatabase(ctx context.Context, input *CreateDatabaseRequest) (*Database, error) {
	var db Database
	if err := c.do(ctx, http.MethodPost, "/databases", input, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (c *SevallaClient) UpdateDatabase(ctx context.Context, id string, input *UpdateDatabaseRequest) (*Database, error) {
	var db Database
	if err := c.do(ctx, http.MethodPatch, path("/databases/%s", id), input, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (c *SevallaClient) DeleteDatabase(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/databases/%s", id), nil, nil)
}

// ResetDatabasePassword changes the password of the database user.
func (c *SevallaClient) ResetDatabasePassword(ctx context.Context, id string, newPassword string) error {
	return c.do(ctx, http.MethodPost, path("/databases/%s/reset-password", id), map[string]string{"new_password": newPassword}, nil)
}

func (c *SevallaClient) ListInternalConnections(ctx context.Context, dbID string) ([]InternalConnection, error) {
	var out []InternalConnection
	err := c.do(ctx, http.MethodGet, path("/databases/%s/internal-connections", dbID), nil, &out)
	return out, err
}

// CreateInternalConnection returns only the new connection's ID.
func (c *SevallaClient) CreateInternalConnection(ctx context.Context, dbID string, input *CreateInternalConnectionRequest) (*InternalConnection, error) {
	var conn InternalConnection
	if err := c.do(ctx, http.MethodPost, path("/databases/%s/internal-connections", dbID), input, &conn); err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *SevallaClient) DeleteInternalConnection(ctx context.Context, dbID string, connID string) error {
	return c.do(ctx, http.MethodDelete, path("/databases/%s/internal-connections/%s", dbID, connID), nil, nil)
}

func (c *SevallaClient) GetDatabaseIPRestriction(ctx context.Context, dbID string) (*IPRestriction, error) {
	var out IPRestriction
	if err := c.do(ctx, http.MethodGet, path("/databases/%s/ip-restriction", dbID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateDatabaseIPRestriction(ctx context.Context, dbID string, input *IPRestriction) (*IPRestriction, error) {
	var out IPRestriction
	if err := c.do(ctx, http.MethodPut, path("/databases/%s/ip-restriction", dbID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
