package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Domain struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	IsPrimary  bool             `json:"is_primary"`
	IsWildcard bool             `json:"is_wildcard"`
	IsEnabled  bool             `json:"is_enabled"`
	Status     *string          `json:"status"`
	DnsRecords *json.RawMessage `json:"dns_records"`
	Errors     *json.RawMessage `json:"errors"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}

type CreateDomainRequest struct {
	DomainName    string  `json:"domain_name"`
	IsWildcard    *bool   `json:"is_wildcard,omitempty"`
	CustomSSLCert *string `json:"custom_ssl_cert,omitempty"`
	CustomSSLKey  *string `json:"custom_ssl_key,omitempty"`
}

func (c *SevallaClient) ListDomains(ctx context.Context, serviceType string, serviceID string) ([]Domain, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("%s/%s/domains", serviceType, serviceID), nil)
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

	var domains []Domain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, fmt.Errorf("decoding domains response: %w", err)
	}

	return domains, nil
}

func (c *SevallaClient) GetDomain(ctx context.Context, serviceType string, serviceID string, domainID string) (*Domain, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("%s/%s/domains/%s", serviceType, serviceID, domainID), nil)
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

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("decoding domain response: %w", err)
	}

	return &domain, nil
}

func (c *SevallaClient) CreateDomain(ctx context.Context, serviceType string, serviceID string, input *CreateDomainRequest) (*Domain, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create domain request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("%s/%s/domains", serviceType, serviceID), bytes.NewReader(body))
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

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("decoding domain response: %w", err)
	}

	return &domain, nil
}

func (c *SevallaClient) DeleteDomain(ctx context.Context, serviceType string, serviceID string, domainID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("%s/%s/domains/%s", serviceType, serviceID, domainID), nil)
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

func (c *SevallaClient) SetPrimaryDomain(ctx context.Context, serviceType string, serviceID string, domainID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("%s/%s/domains/%s/set-primary", serviceType, serviceID, domainID), nil)
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

func (c *SevallaClient) ToggleDomain(ctx context.Context, serviceType string, serviceID string, domainID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("%s/%s/domains/%s/toggle", serviceType, serviceID, domainID), nil)
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

func (c *SevallaClient) RefreshDomainStatus(ctx context.Context, serviceType string, serviceID string, domainID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("%s/%s/domains/%s/refresh-status", serviceType, serviceID, domainID), nil)
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
