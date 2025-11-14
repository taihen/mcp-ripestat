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
	EndpointPath = "/data/looking-glass/data.json"

	MaxLookBackLimit = 48 * 60 * 60
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

	apiResponse := &APIResponse{
		RRCs:      response.Data.RRCs,
		FetchedAt: response.Time,
	}

	if apiResponse.RRCs == nil {
		apiResponse.RRCs = []RRC{}
	}

	return apiResponse, nil
}

func GetLookingGlass(ctx context.Context, resource string, lookBackLimit int) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, lookBackLimit)
}
