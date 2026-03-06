package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetDatabase_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/databases/db-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Database{
			ID:          "db-123",
			Name:        "my-database",
			DisplayName: "My Database",
			Type:        "postgresql",
			ClusterID:   "cluster-1",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	db, err := c.GetDatabase(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-123" {
		t.Errorf("expected ID %q, got %q", "db-123", db.ID)
	}
	if db.DisplayName != "My Database" {
		t.Errorf("expected display name %q, got %q", "My Database", db.DisplayName)
	}
	if db.Type != "postgresql" {
		t.Errorf("expected type %q, got %q", "postgresql", db.Type)
	}
}

func TestGetDatabase_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetDatabase(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateDatabase_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateDatabaseRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New DB" {
			t.Errorf("expected display name %q, got %q", "New DB", req.DisplayName)
		}
		if req.Type != "postgresql" {
			t.Errorf("expected type %q, got %q", "postgresql", req.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Database{
			ID:             "db-new",
			Name:           "new-db",
			DisplayName:    req.DisplayName,
			Type:           req.Type,
			ClusterID:      req.ClusterID,
			ResourceTypeID: &req.ResourceTypeID,
			CreatedAt:      "2025-01-01T00:00:00.000Z",
			UpdatedAt:      "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	db, err := c.CreateDatabase(context.Background(), &CreateDatabaseRequest{
		DisplayName:    "New DB",
		Type:           "postgresql",
		ClusterID:      "cluster-1",
		ResourceTypeID: "rt-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-new" {
		t.Errorf("expected ID %q, got %q", "db-new", db.ID)
	}
	if db.DisplayName != "New DB" {
		t.Errorf("expected display name %q, got %q", "New DB", db.DisplayName)
	}
}

func TestUpdateDatabase_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateDatabaseRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Database{
			ID:          "db-123",
			Name:        "my-database",
			DisplayName: "Updated DB",
			Type:        "postgresql",
			ClusterID:   "cluster-1",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-02T00:00:00.000Z",
		})
	})
	defer server.Close()

	newName := "Updated DB"
	db, err := c.UpdateDatabase(context.Background(), "db-123", &UpdateDatabaseRequest{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.DisplayName != "Updated DB" {
		t.Errorf("expected display name %q, got %q", "Updated DB", db.DisplayName)
	}
}

func TestDeleteDatabase_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteDatabase(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSuspendDatabase(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123/suspend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.SuspendDatabase(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivateDatabase(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123/activate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.ActivateDatabase(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleExternalConnection(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123/external-connection/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := c.ToggleExternalConnection(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDatabaseIPRestriction(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123/ip-restriction" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(IPRestriction{
			Type:      "allow",
			IsEnabled: true,
			IPList:    []string{"192.168.1.0/24"},
		})
	})
	defer server.Close()

	restriction, err := c.GetDatabaseIPRestriction(context.Background(), "db-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !restriction.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}
	if len(restriction.IPList) != 1 {
		t.Fatalf("expected 1 IP, got %d", len(restriction.IPList))
	}
	if restriction.IPList[0] != "192.168.1.0/24" {
		t.Errorf("expected IP %q, got %q", "192.168.1.0/24", restriction.IPList[0])
	}
}

func TestUpdateDatabaseIPRestriction(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/databases/db-123/ip-restriction" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req IPRestriction
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(req)
	})
	defer server.Close()

	restriction, err := c.UpdateDatabaseIPRestriction(context.Background(), "db-123", &IPRestriction{
		Type:      "allow",
		IsEnabled: true,
		IPList:    []string{"10.0.0.0/8"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !restriction.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}
	if len(restriction.IPList) != 1 {
		t.Fatalf("expected 1 IP, got %d", len(restriction.IPList))
	}
	if restriction.IPList[0] != "10.0.0.0/8" {
		t.Errorf("expected IP %q, got %q", "10.0.0.0/8", restriction.IPList[0])
	}
}
