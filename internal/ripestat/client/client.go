// Package client provides a common HTTP client for interacting with the RIPEstat API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/config"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
	"github.com/taihen/mcp-ripestat/internal/ripestat/logging"
	"github.com/taihen/mcp-ripestat/internal/ripestat/metrics"
)

// ripeLimiter enforces RIPE's 8 concurrent request limit with a safety margin.
var ripeLimiter = make(chan struct{}, 7)

// createOptimizedHTTPClient creates an HTTP client with connection pooling and HTTP/2 support.
func createOptimizedHTTPClient(cfg *config.Config) *http.Client {
	// Create custom transport with connection pooling
	transport := &http.Transport{
		// Connection pool settings
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,

		// Performance optimizations
		DisableCompression: false, // Enable compression for bandwidth efficiency
		DisableKeepAlives:  false, // Enable keep-alives for connection reuse

		// Timeouts for connection establishment
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: cfg.Timeout,
		ExpectContinueTimeout: 1 * time.Second,

		// Enable HTTP/2 if requested (Go's standard library enables HTTP/2 by default for HTTPS)
		ForceAttemptHTTP2: cfg.ForceHTTP2,

		// Additional HTTP/2 optimizations
		// Note: HTTP/2 specific settings like ReadIdleTimeout and PingTimeout
		// are handled automatically by Go's HTTP/2 implementation
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		// Note: We don't set CheckRedirect to maintain backward compatibility
	}
}

// HTTPDoer is an interface for making HTTP requests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client provides methods to interact with the RIPEstat API.
type Client struct {
	BaseURL     string
	HTTPClient  HTTPDoer
	UserAgent   string
	SourceApp   string
	RetryConfig *RetryConfig
	Logger      *logging.Logger
	Cache       *cache.Cache
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
		httpClient = createOptimizedHTTPClient(config.DefaultConfig())
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		UserAgent:  config.DefaultUserAgent,
		SourceApp:  config.DefaultSourceApp,
		RetryConfig: &RetryConfig{
			RetryCount:       config.DefaultRetryCount,
			RetryWaitTime:    config.DefaultRetryWaitTime,
			MaxRetryWaitTime: config.DefaultMaxRetryWaitTime,
		},
		Logger: logging.DefaultLogger,
		Cache:  cache.New(),
	}
}

// NewWithConfig creates a new Client with the specified configuration.
func NewWithConfig(cfg *config.Config, httpClient HTTPDoer) *Client {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if httpClient == nil {
		httpClient = createOptimizedHTTPClient(cfg)
	}

	return &Client{
		BaseURL:    cfg.BaseURL,
		HTTPClient: httpClient,
		UserAgent:  cfg.UserAgent,
		SourceApp:  cfg.SourceApp,
		RetryConfig: &RetryConfig{
			RetryCount:       cfg.RetryCount,
			RetryWaitTime:    cfg.RetryWaitTime,
			MaxRetryWaitTime: cfg.MaxRetryWaitTime,
		},
		Logger: logging.DefaultLogger,
		Cache:  cache.New(),
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

	if params == nil {
		params = url.Values{}
	}

	// Add sourceapp parameter for compliance
	if c.SourceApp != "" {
		params.Set("sourceapp", c.SourceApp)
	}

	u.RawQuery = params.Encode()

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

	// Track request timing for slow request warnings.
	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.Logger.Error("Request failed after %v: %v", duration, err)
		return nil, errors.ErrServerError.WithError(fmt.Errorf("request failed: %w", err))
	}

	// Log warning for requests taking more than 10 seconds.
	if duration > 10*time.Second {
		c.Logger.Warning("Slow request to %s took %v", u.String(), duration)
	}

	c.Logger.Debug("Request to %s completed in %v with status: %d", u.String(), duration, resp.StatusCode)

	return resp, nil
}

// GetJSON performs a GET request and decodes the JSON response into the provided target.
func (c *Client) GetJSON(ctx context.Context, endpoint string, params url.Values, target interface{}) error {
	start := time.Now()
	endpointType := extractEndpointType(endpoint)

	// Check cache first
	if c.Cache != nil {
		if cached, found := c.Cache.Get(ctx, endpoint, params); found {
			c.Logger.Debug("Cache hit for endpoint %s", endpoint)
			metrics.RecordCacheHit()

			// Copy cached data to target
			if err := copyInterface(cached, target); err != nil {
				c.Logger.Warning("Failed to copy cached data: %v", err)
				// Continue with API request on cache error
			} else {
				metrics.EndRequest(endpointType, time.Since(start))
				return nil
			}
		}
	}

	metrics.RecordCacheMiss()

	// Acquire rate limiter semaphore
	select {
	case ripeLimiter <- struct{}{}:
		metrics.RecordRateLimitWait()
		defer func() { <-ripeLimiter }()
	case <-ctx.Done():
		metrics.RecordRateLimitTimeout()
		return ctx.Err()
	}

	c.Logger.Debug("Cache miss for endpoint %s, making API request", endpoint)

	// Start request tracking
	metrics.StartRequest()
	defer func() {
		metrics.EndRequest(endpointType, time.Since(start))
	}()

	resp, err := c.Get(ctx, endpoint, params)
	if err != nil {
		metrics.RecordRequest(endpointType, "error")
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	status := fmt.Sprintf("%d", resp.StatusCode)
	metrics.RecordRequest(endpointType, status)

	if resp.StatusCode != http.StatusOK {
		c.Logger.Warning("Received non-OK status code: %d", resp.StatusCode)
		return errors.FromHTTPResponse(resp, "request failed")
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		c.Logger.Error("Failed to decode response: %v", err)
		return errors.ErrServerError.WithError(fmt.Errorf("failed to decode response: %w", err))
	}

	// Cache the successful response
	if c.Cache != nil {
		c.Cache.Set(ctx, endpoint, params, target)
		c.Logger.Debug("Cached response for endpoint %s", endpoint)
	}

	c.Logger.Debug("Successfully decoded response")

	return nil
}

// extractEndpointType extracts the endpoint type from the full endpoint path.
func extractEndpointType(endpoint string) string {
	// Extract the main endpoint type from paths like "/data/network-info"
	if len(endpoint) > 6 && endpoint[:6] == "/data/" {
		return endpoint[6:]
	}

	// For other patterns, use the full endpoint
	return endpoint
}

// copyInterface copies data from source to target using JSON marshaling/unmarshaling.
func copyInterface(source, target interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return nil
}
