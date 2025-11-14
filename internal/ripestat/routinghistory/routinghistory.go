package routinghistory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

type Client struct {
	client *client.Client
}

func New(c *client.Client) *Client {
	return &Client{client: c}
}

func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	endpoint := "/data/routing-history/data.json"

	var response Response
	if err := c.client.GetJSON(ctx, endpoint, params, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) GetWithOptions(ctx context.Context, resource, startTime, endTime string, maxResults int) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	if startTime != "" {
		params.Set("starttime", startTime)
	}

	if endTime != "" {
		params.Set("endtime", endTime)
	}

	if maxResults > 0 {
		params.Set("max_results", fmt.Sprintf("%d", maxResults))
	}

	endpoint := "/data/routing-history/data.json"

	var response Response
	if err := c.client.GetJSON(ctx, endpoint, params, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func DefaultClient() *Client {
	return New(client.DefaultClient())
}

func GetRoutingHistory(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}

func GetRoutingHistoryWithOptions(ctx context.Context, resource, startTime, endTime string, maxResults int) (*Response, error) {
	return DefaultClient().GetWithOptions(ctx, resource, startTime, endTime, maxResults)
}
