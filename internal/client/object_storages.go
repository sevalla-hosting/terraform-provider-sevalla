package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	Domain       *string `json:"domain"`
	Endpoint     *string `json:"endpoint"`
	AccessKey    *string `json:"access_key"`
	SecretKey    *string `json:"secret_key"`
	BucketName   string  `json:"bucket_name"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type CreateObjectStorageRequest struct {
	DisplayName  string  `json:"display_name"`
	Location     *string `json:"location,omitempty"`
	Jurisdiction *string `json:"jurisdiction,omitempty"`
	PublicAccess *bool   `json:"public_access,omitempty"`
	ProjectID    *string `json:"project_id,omitempty"`
}

type UpdateObjectStorageRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type ObjectStorageListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Status      *string `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CORSPolicy struct {
	ID             string   `json:"id"`
	AllowedOrigins []string `json:"origins"`
	AllowedMethods []string `json:"methods"`
	AllowedHeaders []string `json:"headers"`
}

type CreateCORSPolicyRequest struct {
	AllowedOrigins []string `json:"origins"`
	AllowedMethods []string `json:"methods"`
	AllowedHeaders []string `json:"headers,omitempty"`
}

type UpdateCORSPolicyRequest struct {
	AllowedOrigins []string `json:"origins,omitempty"`
	AllowedMethods []string `json:"methods,omitempty"`
	AllowedHeaders []string `json:"headers,omitempty"`
}

func (c *SevallaClient) ListObjectStorages(ctx context.Context) ([]ObjectStorageListItem, error) {
	return Paginate[ObjectStorageListItem](ctx, c, "/object-storage")
}

func (c *SevallaClient) GetObjectStorage(ctx context.Context, id string) (*ObjectStorage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/object-storage/%s", id), nil)
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

	var os ObjectStorage
	if err := json.NewDecoder(resp.Body).Decode(&os); err != nil {
		return nil, fmt.Errorf("decoding object storage response: %w", err)
	}

	return &os, nil
}

func (c *SevallaClient) CreateObjectStorage(ctx context.Context, input *CreateObjectStorageRequest) (*ObjectStorage, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create object storage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/object-storage"), bytes.NewReader(body))
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

	var storage ObjectStorage
	if err := json.NewDecoder(resp.Body).Decode(&storage); err != nil {
		return nil, fmt.Errorf("decoding object storage response: %w", err)
	}

	return &storage, nil
}

func (c *SevallaClient) UpdateObjectStorage(ctx context.Context, id string, input *UpdateObjectStorageRequest) (*ObjectStorage, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update object storage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/object-storage/%s", id), bytes.NewReader(body))
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

	var storage ObjectStorage
	if err := json.NewDecoder(resp.Body).Decode(&storage); err != nil {
		return nil, fmt.Errorf("decoding object storage response: %w", err)
	}

	return &storage, nil
}

func (c *SevallaClient) DeleteObjectStorage(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/object-storage/%s", id), nil)
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

func (c *SevallaClient) EnableObjectStorageCDN(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/object-storage/%s/domain", id), nil)
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

func (c *SevallaClient) DisableObjectStorageCDN(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/object-storage/%s/domain", id), nil)
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

func (c *SevallaClient) ListCORSPolicies(ctx context.Context, osID string) ([]CORSPolicy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/object-storage/%s/cors-policies", osID), nil)
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

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding cors policies response: %w", err)
	}

	var policies []CORSPolicy
	if err := json.Unmarshal(raw, &policies); err != nil {
		var wrapper struct {
			Data []CORSPolicy `json:"data"`
		}
		if err2 := json.Unmarshal(raw, &wrapper); err2 != nil {
			return nil, fmt.Errorf("decoding cors policies response: %w", err)
		}
		policies = wrapper.Data
	}

	return policies, nil
}

func (c *SevallaClient) CreateCORSPolicy(ctx context.Context, osID string, input *CreateCORSPolicyRequest) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshaling create cors policy request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/object-storage/%s/cors-policies", osID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	return nil
}

func (c *SevallaClient) UpdateCORSPolicy(ctx context.Context, osID string, policyID string, input *UpdateCORSPolicyRequest) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshaling update cors policy request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/object-storage/%s/cors-policies/%s", osID, policyID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	return nil
}

func (c *SevallaClient) DeleteCORSPolicy(ctx context.Context, osID string, policyID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/object-storage/%s/cors-policies/%s", osID, policyID), nil)
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
