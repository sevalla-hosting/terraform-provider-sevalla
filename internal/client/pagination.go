package client

import (
	"context"
	"fmt"
	"net/http"
)

// pageLimit is the largest page size the API accepts.
const pageLimit = 100

// Paginate fetches every page of a paginated list endpoint.
func Paginate[T any](ctx context.Context, c *SevallaClient, path string) ([]T, error) {
	var all []T
	for {
		var page PaginatedResponse[T]
		if err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s?limit=%d&offset=%d", path, pageLimit, len(all)), nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if len(page.Data) == 0 || len(all) >= page.Total {
			return all, nil
		}
	}
}
