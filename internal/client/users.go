package client

import (
	"context"
	"net/http"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Image    string `json:"image"`
}

func (c *SevallaClient) ListUsers(ctx context.Context) ([]User, error) {
	var out struct {
		Data []User `json:"data"`
	}
	err := c.do(ctx, http.MethodGet, "/users", nil, &out)
	return out.Data, err
}
