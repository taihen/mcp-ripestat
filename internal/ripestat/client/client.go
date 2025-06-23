// Package client provides a common HTTP client for interacting with the RIPEstat API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/config"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
	"github.com/taihen/mcp-ripestat/internal/ripestat/logging"
)

// Using constants from config package

// HTTPDoer is an interface for making HTTP requests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client provides methods to interact with the RIPEstat API.
type Client struct {
	BaseURL     string
	HTTPClient  HTTPDoer
	UserAgent   string
	RetryConfig *RetryConfig
	Logger      *logging.Logger
}

// RetryConfig holds retry-related configuration.
type RetryConfig struct {
	RetryCount       int
	RetryWaitTime    time.Duration
	MaxRetryWaitTime time.Duration
}

// New creates a new Client with the specified base URL and HTTP client.
// If baseURL is empty, config.DefaultBaseURL is used.
// If httpClient is nil, a new http.Client with config.DefaultTimeout is used.
func New(baseURL string, httpClient HTTPDoer) *Client {
	if baseURL == "" {
		baseURL = config.DefaultBaseURL
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.DefaultTimeout}
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		UserAgent:  config.DefaultUserAgent,
		RetryConfig: &RetryConfig{
			RetryCount:       config.DefaultRetryCount,
			RetryWaitTime:    config.DefaultRetryWaitTime,
			MaxRetryWaitTime: config.DefaultMaxRetryWaitTime,
		},
		Logger: logging.DefaultLogger,
	}
}

// NewWithConfig creates a new Client with the specified configuration.
func NewWithConfig(cfg *config.Config, httpClient HTTPDoer) *Client {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}

	return &Client{
		BaseURL:    cfg.BaseURL,
		HTTPClient: httpClient,
		UserAgent:  cfg.UserAgent,
		RetryConfig: &RetryConfig{
			RetryCount:       cfg.RetryCount,
			RetryWaitTime:    cfg.RetryWaitTime,
			MaxRetryWaitTime: cfg.MaxRetryWaitTime,
		},
		Logger: logging.DefaultLogger,
	}
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return NewWithConfig(config.DefaultConfig(), nil)
}

// Get performs a GET request to the specified endpoint with the given parameters.
func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		c.Logger.Error("Failed to parse URL: %v", err)
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("failed to parse URL: %w", err))
	}

	if params != nil {
		u.RawQuery = params.Encode()
	}

	c.Logger.Debug("Making request to %s", u.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.Logger.Error("Failed to create request: %v", err)
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("failed to create request: %w", err))
	}

	// Set User-Agent header
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("Request failed: %v", err)
		return nil, errors.ErrServerError.WithError(fmt.Errorf("request failed: %w", err))
	}

	c.Logger.Debug("Received response with status code: %d", resp.StatusCode)

	return resp, nil
}

// GetJSON performs a GET request and decodes the JSON response into the provided target.
func (c *Client) GetJSON(ctx context.Context, endpoint string, params url.Values, target interface{}) error {
	resp, err := c.Get(ctx, endpoint, params)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		c.Logger.Warning("Received non-OK status code: %d", resp.StatusCode)
		return errors.FromHTTPResponse(resp, "request failed")
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		c.Logger.Error("Failed to decode response: %v", err)
		return errors.ErrServerError.WithError(fmt.Errorf("failed to decode response: %w", err))
	}

	c.Logger.Debug("Successfully decoded response")

	return nil
}
