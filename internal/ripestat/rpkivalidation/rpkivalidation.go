
package rpkivalidation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (

	EndpointPath = "/data/rpki-validation/data.json"
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


func (c *Client) Get(ctx context.Context, resource, prefix string) (*APIResponse, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	if prefix == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("prefix parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)
	params.Set("prefix", prefix)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get RPKI validation status: %w", err))
	}


	apiResponse := &APIResponse{
		Status:         response.Data.Status,
		Validator:      response.Data.Validator,
		Resource:       response.Data.Resource,
		Prefix:         response.Data.Prefix,
		ValidatingROAs: response.Data.ValidatingROAs,
		FetchedAt:      response.Time,
	}


	if apiResponse.ValidatingROAs == nil {
		apiResponse.ValidatingROAs = []ValidatingROA{}
	}

	return apiResponse, nil
}


func GetRPKIValidation(ctx context.Context, resource, prefix string) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, prefix)
}
