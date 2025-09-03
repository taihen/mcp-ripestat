// Package bgpstate provides access to the RIPEstat bgp-state API.
package bgpstate

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for BGP state information.
	EndpointPath = "/data/bgp-state/data.json"
)

// Client provides methods to interact with the RIPEstat bgp-state API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat bgp-state API.
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

// Options represents parameters for BGP State API requests.
type Options struct {
	Resource       string `json:"resource"`
	Timestamp      string `json:"timestamp,omitempty"`
	RRCs           string `json:"rrcs,omitempty"`
	UnixTimestamps bool   `json:"unix_timestamps,omitempty"`
}

// Get fetches BGP state information for the specified resource.
func (c *Client) Get(ctx context.Context, opts Options) (*Response, error) {
	if opts.Resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", opts.Resource)

	if opts.Timestamp != "" {
		params.Set("timestamp", opts.Timestamp)
	}

	if opts.RRCs != "" {
		params.Set("rrcs", opts.RRCs)
	}

	if opts.UnixTimestamps {
		params.Set("unix_timestamps", "true")
	}

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get BGP state: %w", err))
	}

	return &response, nil
}
