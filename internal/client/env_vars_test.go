package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListEnvironmentVariables_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/app-123/env-vars" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]EnvironmentVariable{
			{ID: "env-1", Key: "DATABASE_URL", Value: "postgres://localhost/mydb"},
			{ID: "env-2", Key: "API_KEY", Value: "secret-key-123"},
		})
	})
	defer server.Close()

	envVars, err := c.ListEnvironmentVariables(context.Background(), "/applications", "app-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envVars) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(envVars))
	}
	if envVars[0].Key != "DATABASE_URL" {
		t.Errorf("expected key %q, got %q", "DATABASE_URL", envVars[0].Key)
	}
	if envVars[1].Value != "secret-key-123" {
		t.Errorf("expected value %q, got %q", "secret-key-123", envVars[1].Value)
	}
}

func TestListEnvironmentVariables_Error(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Bad request",
			"status":  400,
		})
	})
	defer server.Close()

	_, err := c.ListEnvironmentVariables(context.Background(), "/applications", "bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateEnvironmentVariable_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/env-vars" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateEnvironmentVariableRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Key != "NEW_VAR" {
			t.Errorf("expected key %q, got %q", "NEW_VAR", req.Key)
		}
		if req.Value != "new-value" {
			t.Errorf("expected value %q, got %q", "new-value", req.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(EnvironmentVariable{
			ID:    "env-new",
			Key:   req.Key,
			Value: req.Value,
		})
	})
	defer server.Close()

	envVar, err := c.CreateEnvironmentVariable(context.Background(), "/applications", "app-123", &CreateEnvironmentVariableRequest{
		Key:   "NEW_VAR",
		Value: "new-value",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envVar.ID != "env-new" {
		t.Errorf("expected ID %q, got %q", "env-new", envVar.ID)
	}
	if envVar.Key != "NEW_VAR" {
		t.Errorf("expected key %q, got %q", "NEW_VAR", envVar.Key)
	}
}

func TestUpdateEnvironmentVariable_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/env-vars/env-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(EnvironmentVariable{
			ID:    "env-456",
			Key:   "UPDATED_VAR",
			Value: "updated-value",
		})
	})
	defer server.Close()

	newValue := "updated-value"
	envVar, err := c.UpdateEnvironmentVariable(context.Background(), "/applications", "app-123", "env-456", &UpdateEnvironmentVariableRequest{
		Value: &newValue,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if envVar.Value != "updated-value" {
		t.Errorf("expected value %q, got %q", "updated-value", envVar.Value)
	}
}

func TestDeleteEnvironmentVariable_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/applications/app-123/env-vars/env-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteEnvironmentVariable(context.Background(), "/applications", "app-123", "env-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
