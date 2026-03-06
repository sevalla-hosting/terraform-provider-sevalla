package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Image    string `json:"image"`
}

type usersResponse struct {
	Data []User `json:"data"`
}

func (c *SevallaClient) ListUsers(ctx context.Context) ([]User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url("/users"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var result usersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding users response: %w", err)
	}

	return result.Data, nil
}
