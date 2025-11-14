
package whois

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

	endpoint := "/data/whois/data.json"

	var response Response
	if err := c.client.GetJSON(ctx, endpoint, params, &response); err != nil {
		return nil, err
	}

	return &response, nil
}


func DefaultClient() *Client {
	return New(client.DefaultClient())
}


func GetWhois(ctx context.Context, resource string) (*Response, error) {
	return DefaultClient().Get(ctx, resource)
}
