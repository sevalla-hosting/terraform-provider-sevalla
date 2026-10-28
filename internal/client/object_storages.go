package client

import (
	"context"
	"net/http"
)

type ObjectStorage struct {
	ID           string  `json:"id"`
	CompanyID    string  `json:"company_id"`
	ProjectID    *string `json:"project_id"`
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Location     string  `json:"location"`
	Jurisdiction string  `json:"jurisdiction"`
	Status       string  `json:"status"`
	Domain       *string `json:"domain"`
	Endpoint     *string `json:"endpoint"`
	// AccessKey and SecretKey are only returned by the create call.
	AccessKey  *string `json:"access_key"`
	SecretKey  *string `json:"secret_key"`
	BucketName string  `json:"bucket_name"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type CreateObjectStorageRequest struct {
	DisplayName  string  `json:"display_name"`
	Location     *string `json:"location,omitempty"`
	Jurisdiction *string `json:"jurisdiction,omitempty"`
	PublicAccess *bool   `json:"public_access,omitempty"`
	ProjectID    *string `json:"project_id,omitempty"`
}

type UpdateObjectStorageRequest struct {
	DisplayName *string          `json:"display_name,omitempty"`
	ProjectID   Nullable[string] `json:"project_id,omitzero"`
}

type CORSPolicy struct {
	ID             string   `json:"id"`
	AllowedOrigins []string `json:"origins"`
	AllowedMethods []string `json:"methods"`
	AllowedHeaders []string `json:"headers"`
}

type CORSPolicyRequest struct {
	AllowedOrigins []string `json:"origins"`
	AllowedMethods []string `json:"methods"`
	// A non-nil empty slice is sent as [] to clear the headers.
	AllowedHeaders []string `json:"headers,omitzero"`
}

func (c *SevallaClient) GetObjectStorage(ctx context.Context, id string) (*ObjectStorage, error) {
	var out ObjectStorage
	if err := c.do(ctx, http.MethodGet, path("/object-storage/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateObjectStorage(ctx context.Context, input *CreateObjectStorageRequest) (*ObjectStorage, error) {
	var out ObjectStorage
	if err := c.do(ctx, http.MethodPost, "/object-storage", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateObjectStorage(ctx context.Context, id string, input *UpdateObjectStorageRequest) (*ObjectStorage, error) {
	var out ObjectStorage
	if err := c.do(ctx, http.MethodPatch, path("/object-storage/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteObjectStorage starts an asynchronous deletion (HTTP 202); the bucket
// is emptied and removed in the background.
func (c *SevallaClient) DeleteObjectStorage(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/object-storage/%s", id), nil, nil)
}

// EnableObjectStorageCDN turns on the public CDN domain.
func (c *SevallaClient) EnableObjectStorageCDN(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, path("/object-storage/%s/domain", id), nil, nil)
}

// DisableObjectStorageCDN turns off the public CDN domain.
func (c *SevallaClient) DisableObjectStorageCDN(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/object-storage/%s/domain", id), nil, nil)
}

func (c *SevallaClient) ListCORSPolicies(ctx context.Context, osID string) ([]CORSPolicy, error) {
	var out struct {
		Data []CORSPolicy `json:"data"`
	}
	err := c.do(ctx, http.MethodGet, path("/object-storage/%s/cors-policies", osID), nil, &out)
	return out.Data, err
}

// CreateCORSPolicy returns no body; list the policies to find the new one.
func (c *SevallaClient) CreateCORSPolicy(ctx context.Context, osID string, input *CORSPolicyRequest) error {
	return c.do(ctx, http.MethodPost, path("/object-storage/%s/cors-policies", osID), input, nil)
}

func (c *SevallaClient) UpdateCORSPolicy(ctx context.Context, osID string, policyID string, input *CORSPolicyRequest) error {
	return c.do(ctx, http.MethodPatch, path("/object-storage/%s/cors-policies/%s", osID, policyID), input, nil)
}

func (c *SevallaClient) DeleteCORSPolicy(ctx context.Context, osID string, policyID string) error {
	return c.do(ctx, http.MethodDelete, path("/object-storage/%s/cors-policies/%s", osID, policyID), nil, nil)
}
