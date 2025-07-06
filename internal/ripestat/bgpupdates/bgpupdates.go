package bgpupdates

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const EndpointPath = "/data/bgp-updates/data.json"

type Client struct {
	client *client.Client
}

func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get BGP updates: %w", err))
	}

	return &response, nil
}

func DefaultClient() *Client {
	return NewClient(client.DefaultClient())
}

func GetBGPUpdates(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
