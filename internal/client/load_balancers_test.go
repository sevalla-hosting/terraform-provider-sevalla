package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetLoadBalancer_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/load-balancers/lb-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		lbType := "DEFAULT"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoadBalancer{
			ID:          "lb-123",
			Name:        "my-lb",
			DisplayName: "My Load Balancer",
			Type:        &lbType,
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	lb, err := c.GetLoadBalancer(context.Background(), "lb-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lb.ID != "lb-123" {
		t.Errorf("expected ID %q, got %q", "lb-123", lb.ID)
	}
	if lb.DisplayName != "My Load Balancer" {
		t.Errorf("expected display name %q, got %q", "My Load Balancer", lb.DisplayName)
	}
}

func TestGetLoadBalancer_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetLoadBalancer(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateLoadBalancer_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/load-balancers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateLoadBalancerRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New LB" {
			t.Errorf("expected display name %q, got %q", "New LB", req.DisplayName)
		}

		lbType := "DEFAULT"
		if req.Type != nil {
			lbType = *req.Type
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LoadBalancer{
			ID:          "lb-new",
			Name:        "new-lb",
			DisplayName: req.DisplayName,
			Type:        &lbType,
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	lb, err := c.CreateLoadBalancer(context.Background(), &CreateLoadBalancerRequest{
		DisplayName: "New LB",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lb.ID != "lb-new" {
		t.Errorf("expected ID %q, got %q", "lb-new", lb.ID)
	}
}

func TestDeleteLoadBalancer_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/load-balancers/lb-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteLoadBalancer(context.Background(), "lb-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateLoadBalancerDestination_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/load-balancers/lb-123/destinations" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateLoadBalancerDestinationRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.ServiceID != "app-456" {
			t.Errorf("expected service ID %q, got %q", "app-456", req.ServiceID)
		}
		if req.ServiceType != "APP" {
			t.Errorf("expected service type %q, got %q", "APP", req.ServiceType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LoadBalancerDestination{
			ID:        "dest-789",
			ServiceID: &req.ServiceID,
			IsEnabled: true,
		})
	})
	defer server.Close()

	dest, err := c.CreateLoadBalancerDestination(context.Background(), "lb-123", &CreateLoadBalancerDestinationRequest{
		ServiceType: "APP",
		ServiceID:   "app-456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.ID != "dest-789" {
		t.Errorf("expected ID %q, got %q", "dest-789", dest.ID)
	}
	if dest.ServiceID == nil || *dest.ServiceID != "app-456" {
		t.Errorf("expected service ID %q, got %v", "app-456", dest.ServiceID)
	}
	if !dest.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}
}

func TestDeleteLoadBalancerDestination_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/load-balancers/lb-123/destinations/dest-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteLoadBalancerDestination(context.Background(), "lb-123", "dest-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
