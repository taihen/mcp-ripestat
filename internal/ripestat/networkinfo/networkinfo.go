// Package networkinfo provides access to the RIPEstat network-info API.
package networkinfo

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for network information.
	EndpointPath = "/data/network-info/data.json"
)

// Client provides methods to interact with the RIPEstat network-info API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat network-info API.
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

// Get fetches network information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get network information: %w", err))
	}

	// Convert ASNs to strings if needed
	for i, asn := range response.Data.ASNs {
		if _, ok := asn.(string); !ok {
			// Convert to string if it's not already a string
			response.Data.ASNs[i] = fmt.Sprintf("%v", asn)
		}
	}

	return &response, nil
}

// GetNetworkInfo is a convenience function that uses the default client to get network information.
func GetNetworkInfo(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
