package rpkihistory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	EndpointPath = "/data/rpki-history/data.json"
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
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get RPKI history: %w", err))
	}

	return &response, nil
}

func GetRPKIHistory(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
