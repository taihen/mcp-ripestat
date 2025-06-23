// Package rpkivalidation provides access to the RIPEstat rpki-validation API.
package rpkivalidation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for RPKI validation.
	EndpointPath = "/data/rpki-validation/data.json"
)

// Client provides methods to interact with the RIPEstat rpki-validation API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat rpki-validation API.
func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{client: c}
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return NewClient(nil)
}

// Get fetches RPKI validation status for the specified resource (ASN) and prefix combination.
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

	// Transform the response to the expected API format
	apiResponse := &APIResponse{
		Status:         response.Data.Status,
		Validator:      response.Data.Validator,
		Resource:       response.Data.Resource,
		Prefix:         response.Data.Prefix,
		ValidatingROAs: response.Data.ValidatingROAs,
		FetchedAt:      response.Time,
	}

	// Ensure ValidatingROAs is never nil, use empty slice instead
	if apiResponse.ValidatingROAs == nil {
		apiResponse.ValidatingROAs = []ValidatingROA{}
	}

	return apiResponse, nil
}

// GetRPKIValidation is a convenience function that uses the default client to get RPKI validation status.
func GetRPKIValidation(ctx context.Context, resource, prefix string) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, prefix)
}
