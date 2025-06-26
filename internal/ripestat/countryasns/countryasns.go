// Package countryasns provides access to the RIPEstat country-asns API.
package countryasns

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for country ASNs information.
	EndpointPath = "/data/country-asns/data.json"
)

// Client provides methods to interact with the RIPEstat country-asns API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat country-asns API.
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

// GetOptions represents optional parameters for the country-asns API.
type GetOptions struct {
	LOD int // Level of detail: 0 (default) or 1 (includes routed/non-routed ASN lists)
}

// Get fetches country ASN information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string, opts *GetOptions) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	if opts != nil && opts.LOD != 0 {
		if opts.LOD < 0 || opts.LOD > 1 {
			return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("lod parameter must be 0 or 1"))
		}
		params.Set("lod", strconv.Itoa(opts.LOD))
	}

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get country ASNs: %w", err))
	}

	return &response, nil
}

// GetCountryASNs is a convenience function that uses the default client to get country ASN information.
func GetCountryASNs(ctx context.Context, resource string, opts *GetOptions) (*Response, error) {
	return DefaultClient().Get(ctx, resource, opts)
}
