package bgpstate

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	EndpointPath = "/data/bgp-state/data.json"
)

type Client struct {
	client *client.Client
}

func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{client: c}
}

func DefaultClient() *Client {
	return NewClient(nil)
}

type Options struct {
	Resource       string `json:"resource"`
	Timestamp      string `json:"timestamp,omitempty"`
	RRCs           string `json:"rrcs,omitempty"`
	UnixTimestamps bool   `json:"unix_timestamps,omitempty"`
}

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
