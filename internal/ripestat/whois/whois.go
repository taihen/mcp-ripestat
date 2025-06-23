// Package whois provides access to the RIPEstat whois API.
package whois

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

// Client provides access to the RIPEstat whois API.
type Client struct {
	client *client.Client
}

// New creates a new whois client.
func New(c *client.Client) *Client {
	return &Client{client: c}
}

// Get retrieves whois information for the specified resource.
// The resource can be an IP address, IP prefix, or ASN.
func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	// Build the URL with query parameters
	params := url.Values{}
	params.Set("resource", resource)

	endpoint := "/data/whois/data.json"

	var response Response
	if err := c.client.GetJSON(ctx, endpoint, params, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return New(client.DefaultClient())
}

// GetWhois is a convenience function that uses the default client to get whois information.
func GetWhois(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
