package asroutingconsistency

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	EndpointPath = "/data/as-routing-consistency/data.json"
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

func GetASRoutingConsistency(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
