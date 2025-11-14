
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

	EndpointPath = "/data/country-asns/data.json"
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


type GetOptions struct {
	LOD int
}


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


func GetCountryASNs(ctx context.Context, resource string, opts *GetOptions) (*Response, error) {
	return DefaultClient().Get(ctx, resource, opts)
}
