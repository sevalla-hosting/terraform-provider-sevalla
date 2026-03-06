package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type LoadBalancer struct {
	ID          string  `json:"id"`
	CompanyID   *string `json:"company_id"`
	ProjectID   *string `json:"project_id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Type        *string `json:"type,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateLoadBalancerRequest struct {
	DisplayName string  `json:"display_name"`
	Type        *string `json:"type,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type UpdateLoadBalancerRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Type        *string `json:"type,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type LoadBalancerListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Type        *string `json:"type,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type LoadBalancerDestination struct {
	ID          string   `json:"id"`
	ServiceType *string  `json:"service_type"`
	ServiceID   *string  `json:"service_id"`
	IsEnabled   bool     `json:"is_enabled"`
	Weight      *int     `json:"weight"`
	URL         *string  `json:"url"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	CreatedAt   *string  `json:"created_at"`
	UpdatedAt   *string  `json:"updated_at"`
}

type CreateLoadBalancerDestinationRequest struct {
	ServiceType string   `json:"service_type"`
	ServiceID   string   `json:"service_id,omitempty"`
	Weight      *int     `json:"weight,omitempty"`
	URL         *string  `json:"url,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

func (c *SevallaClient) ListLoadBalancers(ctx context.Context) ([]LoadBalancerListItem, error) {
	return Paginate[LoadBalancerListItem](ctx, c, "/load-balancers")
}

func (c *SevallaClient) GetLoadBalancer(ctx context.Context, id string) (*LoadBalancer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/load-balancers/%s", id), nil)
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

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("decoding load balancer response: %w", err)
	}

	return &lb, nil
}

func (c *SevallaClient) CreateLoadBalancer(ctx context.Context, input *CreateLoadBalancerRequest) (*LoadBalancer, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create load balancer request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/load-balancers"), bytes.NewReader(body))
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

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("decoding load balancer response: %w", err)
	}

	return &lb, nil
}

func (c *SevallaClient) UpdateLoadBalancer(ctx context.Context, id string, input *UpdateLoadBalancerRequest) (*LoadBalancer, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update load balancer request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url("/load-balancers/%s", id), bytes.NewReader(body))
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

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("decoding load balancer response: %w", err)
	}

	return &lb, nil
}

func (c *SevallaClient) DeleteLoadBalancer(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/load-balancers/%s", id), nil)
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

func (c *SevallaClient) ListLoadBalancerDestinations(ctx context.Context, lbID string) ([]LoadBalancerDestination, error) {
	return Paginate[LoadBalancerDestination](ctx, c, fmt.Sprintf("/load-balancers/%s/destinations", lbID))
}

func (c *SevallaClient) CreateLoadBalancerDestination(ctx context.Context, lbID string, input *CreateLoadBalancerDestinationRequest) (*LoadBalancerDestination, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create load balancer destination request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/load-balancers/%s/destinations", lbID), bytes.NewReader(body))
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

	var dest LoadBalancerDestination
	if err := json.NewDecoder(resp.Body).Decode(&dest); err != nil {
		return nil, fmt.Errorf("decoding load balancer destination response: %w", err)
	}

	return &dest, nil
}

func (c *SevallaClient) DeleteLoadBalancerDestination(ctx context.Context, lbID string, destID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/load-balancers/%s/destinations/%s", lbID, destID), nil)
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

func (c *SevallaClient) ToggleLoadBalancerDestination(ctx context.Context, lbID string, destID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/load-balancers/%s/destinations/%s/toggle", lbID, destID), nil)
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
