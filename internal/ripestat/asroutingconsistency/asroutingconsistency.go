// Package asroutingconsistency provides access to the RIPEstat as-routing-consistency API.
package asroutingconsistency

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for AS routing consistency information.
	EndpointPath = "/data/as-routing-consistency/data.json"
)

// Client provides methods to interact with the RIPEstat as-routing-consistency API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat as-routing-consistency API.
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

// Get fetches AS routing consistency information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get AS routing consistency: %w", err))
	}

	return &response, nil
}

// GetASRoutingConsistency is a convenience function that uses the default client to get AS routing consistency information.
func GetASRoutingConsistency(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
