// Package lookingglass provides access to the RIPEstat looking-glass API.
package lookingglass

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for looking glass.
	EndpointPath = "/data/looking-glass/data.json"

	// MaxLookBackLimit is the maximum allowed look_back_limit in seconds (48 hours).
	MaxLookBackLimit = 48 * 60 * 60 // 172800 seconds
)

// Client provides methods to interact with the RIPEstat looking-glass API.
type Client struct {
	client *client.Client
}

// NewClient creates a new Client for the RIPEstat looking-glass API.
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

// Get fetches looking glass information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string, lookBackLimit int) (*APIResponse, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	if lookBackLimit < 0 {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("look_back_limit must be non-negative"))
	}

	if lookBackLimit > MaxLookBackLimit {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("look_back_limit cannot exceed %d seconds (48 hours)", MaxLookBackLimit))
	}

	params := url.Values{}
	params.Set("resource", resource)
	if lookBackLimit > 0 {
		params.Set("look_back_limit", strconv.Itoa(lookBackLimit))
	}

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get looking glass information: %w", err))
	}

	// Transform the response to the expected API format
	apiResponse := &APIResponse{
		RRCs:      response.Data.RRCs,
		FetchedAt: response.Time,
	}

	// Ensure RRCs is never nil, use empty slice instead
	if apiResponse.RRCs == nil {
		apiResponse.RRCs = []RRC{}
	}

	return apiResponse, nil
}

// GetLookingGlass is a convenience function that uses the default client to get looking glass information.
func GetLookingGlass(ctx context.Context, resource string, lookBackLimit int) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, lookBackLimit)
}
