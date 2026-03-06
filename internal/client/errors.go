package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Status     int    `json:"status"`
	Data       *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"data,omitempty"`
}

func (e *APIError) Error() string {
	if e.Data != nil && e.Data.Code != "" {
		return fmt.Sprintf("sevalla API error (HTTP %d): %s [%s]", e.StatusCode, e.Message, e.Data.Code)
	}
	return fmt.Sprintf("sevalla API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func IsRateLimited(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

func parseErrorResponse(resp *http.Response) error {
	var apiErr APIError
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code %d", resp.StatusCode),
		}
	}
	apiErr.StatusCode = resp.StatusCode
	return &apiErr
}
