package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetObjectStorage_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/object-storage/os-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		endpoint := "https://s3.example.com"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ObjectStorage{
			ID:          "os-123",
			Name:        "my-storage",
			DisplayName: "My Storage",
			Location:    "enam",
			Jurisdiction: "default",
			Endpoint:    &endpoint,
			BucketName:  "my-bucket",
			CompanyID:   "company-1",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	os, err := c.GetObjectStorage(context.Background(), "os-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if os.ID != "os-123" {
		t.Errorf("expected ID %q, got %q", "os-123", os.ID)
	}
	if os.DisplayName != "My Storage" {
		t.Errorf("expected display name %q, got %q", "My Storage", os.DisplayName)
	}
	if os.BucketName != "my-bucket" {
		t.Errorf("expected bucket name %q, got %q", "my-bucket", os.BucketName)
	}
}

func TestGetObjectStorage_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetObjectStorage(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateObjectStorage_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/object-storage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateObjectStorageRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New Storage" {
			t.Errorf("expected display name %q, got %q", "New Storage", req.DisplayName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ObjectStorage{
			ID:          "os-new",
			Name:        "new-storage",
			DisplayName: req.DisplayName,
			Location:    "enam",
			Jurisdiction: "default",
			BucketName:  "new-storage",
			CompanyID:   "company-1",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	os, err := c.CreateObjectStorage(context.Background(), &CreateObjectStorageRequest{
		DisplayName: "New Storage",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if os.ID != "os-new" {
		t.Errorf("expected ID %q, got %q", "os-new", os.ID)
	}
	if os.DisplayName != "New Storage" {
		t.Errorf("expected display name %q, got %q", "New Storage", os.DisplayName)
	}
}

func TestDeleteObjectStorage_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/object-storage/os-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteObjectStorage(context.Background(), "os-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCORSPolicy_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/object-storage/os-123/cors-policies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateCORSPolicyRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.AllowedOrigins) != 1 || req.AllowedOrigins[0] != "https://example.com" {
			t.Errorf("unexpected allowed origins: %v", req.AllowedOrigins)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "CORS policy created"})
	})
	defer server.Close()

	err := c.CreateCORSPolicy(context.Background(), "os-123", &CreateCORSPolicyRequest{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCORSPolicy_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/object-storage/os-123/cors-policies/cors-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteCORSPolicy(context.Background(), "os-123", "cors-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
