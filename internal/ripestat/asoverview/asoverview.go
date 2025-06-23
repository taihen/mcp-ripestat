// Package asoverview provides access to the RIPEstat as-overview API.
package asoverview

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for AS overview information.
	EndpointPath = "/data/as-overview/data.json"
)

// Client provides methods to interact with the RIPEstat as-overview API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat as-overview API.
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

// Get fetches AS overview information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get AS overview: %w", err))
	}

	return &response, nil
}

// GetASOverview is a convenience function that uses the default client to get AS overview information.
func GetASOverview(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
