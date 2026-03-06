package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetProject_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Project{
			ID:          "proj-123",
			Name:        "my-project",
			DisplayName: "My Project",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	project, err := c.GetProject(context.Background(), "proj-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != "proj-123" {
		t.Errorf("expected ID %q, got %q", "proj-123", project.ID)
	}
	if project.DisplayName != "My Project" {
		t.Errorf("expected display name %q, got %q", "My Project", project.DisplayName)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetProject(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreateProject_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateProjectRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New Project" {
			t.Errorf("expected display name %q, got %q", "New Project", req.DisplayName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Project{
			ID:          "proj-new",
			Name:        "new-project",
			DisplayName: req.DisplayName,
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	project, err := c.CreateProject(context.Background(), &CreateProjectRequest{
		DisplayName: "New Project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != "proj-new" {
		t.Errorf("expected ID %q, got %q", "proj-new", project.ID)
	}
	if project.DisplayName != "New Project" {
		t.Errorf("expected display name %q, got %q", "New Project", project.DisplayName)
	}
}

func TestUpdateProject_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/projects/proj-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Project{
			ID:          "proj-123",
			Name:        "my-project",
			DisplayName: "Updated Project",
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-02T00:00:00.000Z",
		})
	})
	defer server.Close()

	newName := "Updated Project"
	project, err := c.UpdateProject(context.Background(), "proj-123", &UpdateProjectRequest{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.DisplayName != "Updated Project" {
		t.Errorf("expected display name %q, got %q", "Updated Project", project.DisplayName)
	}
}

func TestDeleteProject_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/projects/proj-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeleteProject(context.Background(), "proj-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
