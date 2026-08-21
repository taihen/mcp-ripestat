package bgpupdates

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const EndpointPath = "/data/bgp-updates/data.json"

const MaxUpdates = 10000

type Client struct {
	client *client.Client
}

type GetOptions struct {
	StartTime string
	EndTime   string
}

func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

func (c *Client) Get(ctx context.Context, resource string) (*Response, error) {
	return c.GetWithOptions(ctx, resource, nil)
}

func (c *Client) GetWithOptions(ctx context.Context, resource string, opts *GetOptions) (*Response, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)
	if opts != nil {
		if opts.StartTime != "" {
			params.Set("starttime", opts.StartTime)
		}
		if opts.EndTime != "" {
			params.Set("endtime", opts.EndTime)
		}
	}

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get BGP updates: %w", err))
	}
	if len(response.Data.Updates) > MaxUpdates || response.Data.NumUpdates > MaxUpdates {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("BGP update result exceeds limit of %d records", MaxUpdates))
	}

	return &response, nil
}

func DefaultClient() *Client {
	return NewClient(client.DefaultClient())
}

func GetBGPUpdates(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
