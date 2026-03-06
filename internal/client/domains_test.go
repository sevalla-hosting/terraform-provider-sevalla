package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListDomains_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/app-123/domains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Domain{
			{
				ID:        "dom-1",
				Name:      "example.com",
				Type:      "custom",
				IsPrimary: true,
				IsEnabled: true,
				CreatedAt: "2025-01-01T00:00:00.000Z",
				UpdatedAt: "2025-01-01T00:00:00.000Z",
			},
			{
				ID:        "dom-2",
				Name:      "www.example.com",
				Type:      "custom",
				IsPrimary: false,
				IsEnabled:   true,
				CreatedAt: "2025-01-01T00:00:00.000Z",
				UpdatedAt: "2025-01-01T00:00:00.000Z",
			},
		})
	})
	defer server.Close()

	domains, err := c.ListDomains(context.Background(), "/applications", "app-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}
	if domains[0].Name != "example.com" {
		t.Errorf("expected name %q, got %q", "example.com", domains[0].Name)
	}
	if !domains[0].IsPrimary {
		t.Error("expected first domain to be primary")
	}
}

func TestGetDomain_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/app-123/domains/dom-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Domain{
			ID:        "dom-456",
			Name:      "example.com",
			Type:      "custom",
			IsPrimary: true,
			IsEnabled:   true,
			CreatedAt: "2025-01-01T00:00:00.000Z",
			UpdatedAt: "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	domain, err := c.GetDomain(context.Background(), "/applications", "app-123", "dom-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.ID != "dom-456" {
		t.Errorf("expected ID %q, got %q", "dom-456", domain.ID)
	}
	if domain.Name != "example.com" {
		t.Errorf("expected name %q, got %q", "example.com", domain.Name)
	}
}

func TestGetDomain_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetDomain(context.Background(), "/applications", "app-123", "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateDomain_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/domains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateDomainRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DomainName != "new.example.com" {
			t.Errorf("expected domain_name %q, got %q", "new.example.com", req.DomainName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Domain{
			ID:        "dom-new",
			Name:      req.DomainName,
			Type:      "custom",
			IsPrimary: false,
			IsEnabled:   true,
			CreatedAt: "2025-01-01T00:00:00.000Z",
			UpdatedAt: "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	domain, err := c.CreateDomain(context.Background(), "/applications", "app-123", &CreateDomainRequest{
		DomainName: "new.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.ID != "dom-new" {
		t.Errorf("expected ID %q, got %q", "dom-new", domain.ID)
	}
	if domain.Name != "new.example.com" {
		t.Errorf("expected name %q, got %q", "new.example.com", domain.Name)
	}
}

func TestDeleteDomain_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/domains/dom-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteDomain(context.Background(), "/applications", "app-123", "dom-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetPrimaryDomain_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/domains/dom-456/set-primary" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.SetPrimaryDomain(context.Background(), "/applications", "app-123", "dom-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
