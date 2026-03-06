package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	DisplayName    *string `json:"display_name,omitempty"`
	ResourceTypeID *string `json:"resource_type_id,omitempty"`
	ProjectID      *string `json:"project_id,omitempty"`
}

type DatabaseListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	Status      *string `json:"status"`
	IsSuspended bool    `json:"is_suspended"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
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

// WaitForDatabaseStatus polls the database until it reaches one of the target statuses.
// It polls every 2 seconds for up to maxWait duration.
func (c *SevallaClient) WaitForDatabaseStatus(ctx context.Context, id string, targetStatuses []string, maxWait time.Duration) (*Database, error) {
	deadline := time.Now().Add(maxWait)
	for {
		db, err := c.GetDatabase(ctx, id)
		if err != nil {
			if IsNotFound(err) {
				return nil, err
			}
			return nil, fmt.Errorf("polling database status: %w", err)
		}

		if db.Status != nil {
			for _, s := range targetStatuses {
				if *db.Status == s {
					return db, nil
				}
			}
		}

		if time.Now().After(deadline) {
			currentStatus := "<nil>"
			if db.Status != nil {
				currentStatus = *db.Status
			}
			return nil, fmt.Errorf("database %s did not reach status %v within %s (current: %s)", id, targetStatuses, maxWait, currentStatus)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *SevallaClient) ListDatabases(ctx context.Context) ([]DatabaseListItem, error) {
	return Paginate[DatabaseListItem](ctx, c, "/databases")
}

func (c *SevallaClient) GetDatabase(ctx context.Context, id string) (*Database, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/databases/%s", id), nil)
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

	var db Database
	if err := json.NewDecoder(resp.Body).Decode(&db); err != nil {
		return nil, fmt.Errorf("decoding database response: %w", err)
	}

	return &db, nil
}

func (c *SevallaClient) CreateDatabase(ctx context.Context, input *CreateDatabaseRequest) (*Database, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create database request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/databases"), bytes.NewReader(body))
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

	var db Database
	if err := json.NewDecoder(resp.Body).Decode(&db); err != nil {
		return nil, fmt.Errorf("decoding database response: %w", err)
	}

	return &db, nil
}

func (c *SevallaClient) UpdateDatabase(ctx context.Context, id string, input *UpdateDatabaseRequest) (*Database, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update database request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/databases/%s", id), bytes.NewReader(body))
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

	var db Database
	if err := json.NewDecoder(resp.Body).Decode(&db); err != nil {
		return nil, fmt.Errorf("decoding database response: %w", err)
	}

	return &db, nil
}

func (c *SevallaClient) DeleteDatabase(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/databases/%s", id), nil)
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

func (c *SevallaClient) SuspendDatabase(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/databases/%s/suspend", id), nil)
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

func (c *SevallaClient) ActivateDatabase(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/databases/%s/activate", id), nil)
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

func (c *SevallaClient) ToggleExternalConnection(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/databases/%s/external-connection/toggle", id), nil)
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

func (c *SevallaClient) ListInternalConnections(ctx context.Context, dbID string) ([]InternalConnection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/databases/%s/internal-connections", dbID), nil)
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

	var connections []InternalConnection
	if err := json.NewDecoder(resp.Body).Decode(&connections); err != nil {
		return nil, fmt.Errorf("decoding internal connections response: %w", err)
	}

	return connections, nil
}

func (c *SevallaClient) CreateInternalConnection(ctx context.Context, dbID string, input *CreateInternalConnectionRequest) (*InternalConnection, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create internal connection request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/databases/%s/internal-connections", dbID), bytes.NewReader(body))
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

	var conn InternalConnection
	if err := json.NewDecoder(resp.Body).Decode(&conn); err != nil {
		return nil, fmt.Errorf("decoding internal connection response: %w", err)
	}

	return &conn, nil
}

func (c *SevallaClient) DeleteInternalConnection(ctx context.Context, dbID string, connID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/databases/%s/internal-connections/%s", dbID, connID), nil)
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

func (c *SevallaClient) GetDatabaseIPRestriction(ctx context.Context, dbID string) (*IPRestriction, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/databases/%s/ip-restriction", dbID), nil)
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

	var restriction IPRestriction
	if err := json.NewDecoder(resp.Body).Decode(&restriction); err != nil {
		return nil, fmt.Errorf("decoding ip restriction response: %w", err)
	}

	return &restriction, nil
}

func (c *SevallaClient) UpdateDatabaseIPRestriction(ctx context.Context, dbID string, input *IPRestriction) (*IPRestriction, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update ip restriction request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url("/databases/%s/ip-restriction", dbID), bytes.NewReader(body))
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

	var restriction IPRestriction
	if err := json.NewDecoder(resp.Body).Decode(&restriction); err != nil {
		return nil, fmt.Errorf("decoding ip restriction response: %w", err)
	}

	return &restriction, nil
}
