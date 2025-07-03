// Package relatedprefixes provides access to the RIPEstat related-prefixes API.
package relatedprefixes

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for related prefixes information.
	EndpointPath = "/data/related-prefixes/data.json"
)

// Client provides methods to interact with the RIPEstat related-prefixes API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat related-prefixes API.
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

// Get fetches related prefixes information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get related prefixes: %w", err))
	}

	return &response, nil
}

// GetRelatedPrefixes is a convenience function that uses the default client to get related prefixes information.
func GetRelatedPrefixes(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
