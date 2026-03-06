package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func Paginate[T any](ctx context.Context, c *SevallaClient, path string) ([]T, error) {
	var all []T
	offset := 0
	limit := 100

	for {
		pageData, total, err := fetchPage[T](ctx, c, path, limit, offset)
		if err != nil {
			return nil, err
		}

		all = append(all, pageData...)

		if offset+limit >= total {
			break
		}
		offset += limit
	}

	return all, nil
}

func fetchPage[T any](ctx context.Context, c *SevallaClient, path string, limit, offset int) ([]T, int, error) {
	url := fmt.Sprintf("%s?limit=%d&offset=%d", c.url(path), limit, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, parseErrorResponse(resp)
	}

	var page PaginatedResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, 0, fmt.Errorf("decoding paginated response: %w", err)
	}

	return page.Data, page.Total, nil
}
