package client

import (
	"context"
	"net/http"
)

type LoadBalancer struct {
	ID          string  `json:"id"`
	CompanyID   *string `json:"company_id"`
	ProjectID   *string `json:"project_id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateLoadBalancerRequest struct {
	DisplayName string  `json:"display_name"`
	Type        *string `json:"type,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
}

type UpdateLoadBalancerRequest struct {
	DisplayName *string          `json:"display_name,omitempty"`
	Type        *string          `json:"type,omitempty"`
	ProjectID   Nullable[string] `json:"project_id,omitzero"`
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

func (c *SevallaClient) ListLoadBalancers(ctx context.Context) ([]LoadBalancer, error) {
	return Paginate[LoadBalancer](ctx, c, "/load-balancers")
}

func (c *SevallaClient) GetLoadBalancer(ctx context.Context, id string) (*LoadBalancer, error) {
	var out LoadBalancer
	if err := c.do(ctx, http.MethodGet, path("/load-balancers/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateLoadBalancer(ctx context.Context, input *CreateLoadBalancerRequest) (*LoadBalancer, error) {
	var out LoadBalancer
	if err := c.do(ctx, http.MethodPost, "/load-balancers", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateLoadBalancer(ctx context.Context, id string, input *UpdateLoadBalancerRequest) (*LoadBalancer, error) {
	var out LoadBalancer
	if err := c.do(ctx, http.MethodPatch, path("/load-balancers/%s", id), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteLoadBalancer(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, path("/load-balancers/%s", id), nil, nil)
}

func (c *SevallaClient) ListLoadBalancerDestinations(ctx context.Context, lbID string) ([]LoadBalancerDestination, error) {
	return Paginate[LoadBalancerDestination](ctx, c, path("/load-balancers/%s/destinations", lbID))
}

func (c *SevallaClient) CreateLoadBalancerDestination(ctx context.Context, lbID string, input *CreateLoadBalancerDestinationRequest) (*LoadBalancerDestination, error) {
	var out LoadBalancerDestination
	if err := c.do(ctx, http.MethodPost, path("/load-balancers/%s/destinations", lbID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteLoadBalancerDestination(ctx context.Context, lbID string, destID string) error {
	return c.do(ctx, http.MethodDelete, path("/load-balancers/%s/destinations/%s", lbID, destID), nil, nil)
}

func (c *SevallaClient) ToggleLoadBalancerDestination(ctx context.Context, lbID string, destID string) error {
	return c.do(ctx, http.MethodPost, path("/load-balancers/%s/destinations/%s/toggle", lbID, destID), nil, nil)
}
