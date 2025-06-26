// Package rpkihistory provides methods to interact with the RIPEStat RPKI
// History endpoint.
package rpkihistory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the API path for the RPKI History endpoint.
	EndpointPath = "/data/rpki-history/data.json"
)

// Client provides methods to interact with the RPKI History endpoint.
type Client struct {
	client *client.Client
}

// NewClient creates a new RPKI History client with the given HTTP client.
func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}
	return &Client{client: c}
}

// DefaultClient returns a RPKI History client with the default HTTP client.
func DefaultClient() *Client {
	return NewClient(nil)
}

// Get retrieves RPKI History information for the given resource.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get RPKI history: %w", err))
	}

	return &response, nil
}

// GetRPKIHistory retrieves RPKI History information using the default client.
func GetRPKIHistory(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
