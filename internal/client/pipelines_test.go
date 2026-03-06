package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetPipeline_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pipelines/pipe-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Pipeline{
			ID:          "pipe-123",
			DisplayName: "My Pipeline",
			Type:        "trunk",
			Stages: []PipelineStage{
				{ID: "stage-1", DisplayName: "Dev", Type: "standard", Order: 1},
			},
			CreatedAt: "2025-01-01T00:00:00.000Z",
			UpdatedAt: "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	pipeline, err := c.GetPipeline(context.Background(), "pipe-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pipeline.ID != "pipe-123" {
		t.Errorf("expected ID %q, got %q", "pipe-123", pipeline.ID)
	}
	if pipeline.DisplayName != "My Pipeline" {
		t.Errorf("expected display name %q, got %q", "My Pipeline", pipeline.DisplayName)
	}
	if pipeline.Type != "trunk" {
		t.Errorf("expected type %q, got %q", "trunk", pipeline.Type)
	}
	if len(pipeline.Stages) != 1 {
		t.Fatalf("expected 1 stage, got %d", len(pipeline.Stages))
	}
	if pipeline.Stages[0].DisplayName != "Dev" {
		t.Errorf("expected stage display name %q, got %q", "Dev", pipeline.Stages[0].DisplayName)
	}
}

func TestGetPipeline_NotFound(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Not found",
			"status":  404,
		})
	})
	defer server.Close()

	_, err := c.GetPipeline(context.Background(), "missing-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestCreatePipeline_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/pipelines" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreatePipelineRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "New Pipeline" {
			t.Errorf("expected display name %q, got %q", "New Pipeline", req.DisplayName)
		}
		if req.Type != "trunk" {
			t.Errorf("expected type %q, got %q", "trunk", req.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Pipeline{
			ID:          "pipe-new",
			DisplayName: req.DisplayName,
			Type:        req.Type,
			Stages:      []PipelineStage{},
			CreatedAt:   "2025-01-01T00:00:00.000Z",
			UpdatedAt:   "2025-01-01T00:00:00.000Z",
		})
	})
	defer server.Close()

	pipeline, err := c.CreatePipeline(context.Background(), &CreatePipelineRequest{
		DisplayName: "New Pipeline",
		Type:        "trunk",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pipeline.ID != "pipe-new" {
		t.Errorf("expected ID %q, got %q", "pipe-new", pipeline.ID)
	}
	if pipeline.DisplayName != "New Pipeline" {
		t.Errorf("expected display name %q, got %q", "New Pipeline", pipeline.DisplayName)
	}
}

func TestDeletePipeline_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/pipelines/pipe-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeletePipeline(context.Background(), "pipe-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePipelineStage_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/pipelines/pipe-123/stages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreatePipelineStageRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "production" {
			t.Errorf("expected display name %q, got %q", "production", req.DisplayName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(PipelineStage{
			ID:          "stage-new",
			DisplayName: req.DisplayName,
			Type:        "standard",
			Order:       3,
		})
	})
	defer server.Close()

	stage, err := c.CreatePipelineStage(context.Background(), "pipe-123", &CreatePipelineStageRequest{
		DisplayName:  "production",
		InsertBefore: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stage.ID != "stage-new" {
		t.Errorf("expected ID %q, got %q", "stage-new", stage.ID)
	}
	if stage.DisplayName != "production" {
		t.Errorf("expected display name %q, got %q", "production", stage.DisplayName)
	}
}

func TestDeletePipelineStage_Success(t *testing.T) {
	c, server := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/pipelines/pipe-123/stages/stage-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := c.DeletePipelineStage(context.Background(), "pipe-123", "stage-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
