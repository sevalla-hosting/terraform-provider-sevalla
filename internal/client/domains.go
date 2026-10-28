package client

import (
	"context"
	"encoding/json"
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

// serviceType is the collection path of the owning service, e.g. "/applications".
func (c *SevallaClient) GetDomain(ctx context.Context, serviceType string, serviceID string, domainID string) (*Domain, error) {
	var out Domain
	if err := c.do(ctx, http.MethodGet, path(serviceType+"/%s/domains/%s", serviceID, domainID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateDomain(ctx context.Context, serviceType string, serviceID string, input *CreateDomainRequest) (*Domain, error) {
	var out Domain
	if err := c.do(ctx, http.MethodPost, path(serviceType+"/%s/domains", serviceID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteDomain(ctx context.Context, serviceType string, serviceID string, domainID string) error {
	return c.do(ctx, http.MethodDelete, path(serviceType+"/%s/domains/%s", serviceID, domainID), nil, nil)
}
