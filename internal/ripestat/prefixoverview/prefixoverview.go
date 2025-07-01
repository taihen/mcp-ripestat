// Package prefixoverview provides access to the RIPEstat prefix-overview API.
package prefixoverview

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for prefix overview information.
	EndpointPath = "/data/prefix-overview/data.json"
)

// Client provides methods to interact with the RIPEstat prefix-overview API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat prefix-overview API.
func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{client: c}
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return NewClient(nil)
}

// Get fetches prefix overview information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get prefix overview: %w", err))
	}

	return &response, nil
}

// GetPrefixOverview is a convenience function that uses the default client to get prefix overview information.
func GetPrefixOverview(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
