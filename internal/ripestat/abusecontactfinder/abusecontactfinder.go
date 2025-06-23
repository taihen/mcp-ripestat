// Package abusecontactfinder provides access to the RIPEstat abuse-contact-finder API.
package abusecontactfinder

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for abuse contact finder.
	EndpointPath = "/data/abuse-contact-finder/data.json"
)

// Client provides methods to interact with the RIPEstat abuse-contact-finder API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat abuse-contact-finder API.
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

// Get fetches abuse contact information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string) (*APIResponse, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	params := url.Values{}
	params.Set("resource", resource)

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get abuse contact information: %w", err))
	}

	// Transform the response to the expected API format
	apiResponse := &APIResponse{
		Contacts:  response.Data.AbuseContacts,
		FetchedAt: response.Time,
	}

	// Ensure contacts is never nil, use empty slice instead
	if apiResponse.Contacts == nil {
		apiResponse.Contacts = []string{}
	}

	return apiResponse, nil
}

// GetAbuseContactFinder is a convenience function that uses the default client to get abuse contact information.
func GetAbuseContactFinder(ctx context.Context, resource string) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource)
}
