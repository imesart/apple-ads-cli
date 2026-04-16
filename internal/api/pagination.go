package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/imesart/apple-ads-cli/internal/types"
)

const (
	// defaultPageSize is the maximum number of records per page (API limit).
	defaultPageSize = 1000
)

// paginatedRequest wraps a Request with offset/limit query parameters.
type paginatedRequest struct {
	inner  Request
	offset int
	limit  int
}

func (p *paginatedRequest) Method() string { return p.inner.Method() }
func (p *paginatedRequest) Path() string   { return p.inner.Path() }
func (p *paginatedRequest) Body() any      { return p.inner.Body() }

func (p *paginatedRequest) Query() url.Values {
	q := url.Values{}
	// Copy existing query params
	if inner := p.inner.Query(); inner != nil {
		for k, vs := range inner {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
	}
	q.Set("offset", fmt.Sprintf("%d", p.offset))
	q.Set("limit", fmt.Sprintf("%d", p.limit))
	return q
}

// FetchAll fetches all pages for a GET list endpoint.
// It uses offset/limit query parameters and follows pagination until all
// records are retrieved. The default page size is 1000 (API maximum).
func FetchAll[T any](ctx context.Context, client *Client, req Request) ([]T, error) {
	var all []T
	offset := 0

	for {
		pageReq := &paginatedRequest{
			inner:  req,
			offset: offset,
			limit:  defaultPageSize,
		}

		var resp types.ListResponse[T]
		if err := client.Do(ctx, pageReq, &resp); err != nil {
			return nil, fmt.Errorf("fetching page at offset %d: %w", offset, err)
		}

		all = append(all, resp.Data...)

		// Check if we've received all results
		if resp.Pagination == nil {
			break
		}
		fetched := resp.Pagination.StartIndex + resp.Pagination.ItemsPerPage
		if fetched >= resp.Pagination.TotalResults {
			break
		}

		offset = fetched
	}

	return all, nil
}

// FetchAllRaw fetches all pages for a GET list endpoint and returns a
// response-shaped JSON object containing the merged data array.
func FetchAllRaw(ctx context.Context, client *Client, req Request) (json.RawMessage, error) {
	rows, err := FetchAll[json.RawMessage](ctx, client, req)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(map[string]any{"data": rows})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
