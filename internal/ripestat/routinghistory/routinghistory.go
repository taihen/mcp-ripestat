package routinghistory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

// Client provides access to the RIPEstat routing-history API.
type Client struct {
	client *client.Client
}

// New creates a new routing-history client.
func New(c *client.Client) *Client {
	return &Client{client: c}
}

// Get retrieves routing history information for the specified resource.
// The resource can be an IP address, IP prefix, or ASN.
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

// GetWithOptions retrieves routing history information with optional time range and limit parameters.
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

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return New(client.DefaultClient())
}

// GetRoutingHistory is a convenience function that uses the default client to get routing history information.
func GetRoutingHistory(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}

// GetRoutingHistoryWithOptions is a convenience function that uses the default client to get routing history information with options.
func GetRoutingHistoryWithOptions(ctx context.Context, resource, startTime, endTime string, maxResults int) (*Response, error) {
	return DefaultClient().GetWithOptions(ctx, resource, startTime, endTime, maxResults)
}
