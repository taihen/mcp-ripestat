package aspathlength

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)


const EndpointPath = "/data/as-path-length/data.json"


type Client struct {
	client *client.Client
}


func NewClient(httpClient *client.Client) *Client {
	if httpClient == nil {
		httpClient = client.DefaultClient()
	}
	return &Client{client: httpClient}
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
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get AS path length data: %w", err))
	}

	return &response, nil
}


func GetASPathLength(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
