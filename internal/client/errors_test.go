package client

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParseErrorResponse_WithData(t *testing.T) {
	body := `{"message":"Validation failed","status":422,"data":{"code":"INVALID_INPUT","message":"display_name is required"}}`
	resp := &http.Response{
		StatusCode: 422,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := parseErrorResponse(resp)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected status code 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Validation failed" {
		t.Errorf("expected message %q, got %q", "Validation failed", apiErr.Message)
	}
	if apiErr.Data == nil {
		t.Fatal("expected Data to be non-nil")
	}
	if apiErr.Data.Code != "INVALID_INPUT" {
		t.Errorf("expected data code %q, got %q", "INVALID_INPUT", apiErr.Data.Code)
	}
	if apiErr.Data.Message != "display_name is required" {
		t.Errorf("expected data message %q, got %q", "display_name is required", apiErr.Data.Message)
	}

	expected := "sevalla API error (HTTP 422): Validation failed [INVALID_INPUT]"
	if apiErr.Error() != expected {
		t.Errorf("expected error string %q, got %q", expected, apiErr.Error())
	}
}

func TestParseErrorResponse_MinimalBody(t *testing.T) {
	body := `{"message":"Not found","status":404}`
	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := parseErrorResponse(resp)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Not found" {
		t.Errorf("expected message %q, got %q", "Not found", apiErr.Message)
	}
	if apiErr.Data != nil {
		t.Errorf("expected Data to be nil, got %+v", apiErr.Data)
	}

	expected := "sevalla API error (HTTP 404): Not found"
	if apiErr.Error() != expected {
		t.Errorf("expected error string %q, got %q", expected, apiErr.Error())
	}
}

func TestParseErrorResponse_InvalidJSON(t *testing.T) {
	body := `<html>Internal Server Error</html>`
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := parseErrorResponse(resp)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "unexpected status code 500" {
		t.Errorf("expected fallback message, got %q", apiErr.Message)
	}
}

func TestIsRateLimited(t *testing.T) {
	if IsRateLimited(nil) {
		t.Error("expected false for nil error")
	}
	if !IsRateLimited(&APIError{StatusCode: 429}) {
		t.Error("expected true for 429")
	}
	if IsRateLimited(&APIError{StatusCode: 400}) {
		t.Error("expected false for 400")
	}
	if IsRateLimited(&APIError{StatusCode: 500}) {
		t.Error("expected false for 500")
	}
}
