package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/config"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
	"github.com/taihen/mcp-ripestat/internal/ripestat/logging"
	"github.com/taihen/mcp-ripestat/internal/ripestat/metrics"
)

var ripeLimiter = make(chan struct{}, 7)

// MaxResponseBodyBytes bounds the decompressed upstream response held in
// memory. RIPEstat responses are read before decoding so compressed payloads
// cannot bypass the limit.
const MaxResponseBodyBytes int64 = 16 << 20

func createOptimizedHTTPClient(cfg *config.Config) *http.Client {

	transport := &http.Transport{

		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,

		DisableCompression: false,
		DisableKeepAlives:  false,

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: cfg.Timeout,
		ExpectContinueTimeout: 1 * time.Second,

		ForceAttemptHTTP2: cfg.ForceHTTP2,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL     string
	HTTPClient  HTTPDoer
	UserAgent   string
	SourceApp   string
	RetryConfig *RetryConfig
	Logger      *logging.Logger
	Cache       *cache.Cache
}

type RetryConfig struct {
	RetryCount       int
	RetryWaitTime    time.Duration
	MaxRetryWaitTime time.Duration
}

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

func DefaultClient() *Client {
	return NewWithConfig(config.DefaultConfig(), nil)
}

func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		c.Logger.Error("Failed to parse URL: %v", err)
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("failed to parse URL: %w", err))
	}

	if params == nil {
		params = url.Values{}
	}

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

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	// Execute request with retry logic
	return c.doWithRetry(ctx, req)
}

// doWithRetry executes the HTTP request with exponential backoff retry logic.
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	retryCount := 0
	if c.RetryConfig != nil {
		retryCount = c.RetryConfig.RetryCount
	}

	waitTime := c.RetryConfig.RetryWaitTime
	maxWait := c.RetryConfig.MaxRetryWaitTime

	for attempt := 0; attempt <= retryCount; attempt++ {
		// Check context cancellation before each attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		start := time.Now()
		resp, err := c.HTTPClient.Do(req)
		duration := time.Since(start)

		if err == nil {
			if duration > 10*time.Second {
				c.Logger.Warning("Slow request to %s took %v", req.URL.String(), duration)
			}

			// Check if we should retry based on status code
			if !isRetryableStatusCode(resp.StatusCode) {
				c.Logger.Debug("Request to %s completed in %v with status: %d", req.URL.String(), duration, resp.StatusCode)
				return resp, nil
			}

			// Retryable status code - close body and retry
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("received retryable status code: %d", resp.StatusCode)
			c.Logger.Warning("Retryable status code %d, attempt %d/%d", resp.StatusCode, attempt+1, retryCount+1)
		} else {
			lastErr = err
			c.Logger.Warning("Request failed (attempt %d/%d): %v", attempt+1, retryCount+1, err)
		}

		// Don't sleep after the last attempt
		if attempt < retryCount {
			// Wait before retrying with exponential backoff
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}

			// Exponential backoff: double the wait time, up to max
			waitTime *= 2
			if waitTime > maxWait {
				waitTime = maxWait
			}
		}
	}

	c.Logger.Error("Request failed after %d attempts: %v", retryCount+1, lastErr)
	return nil, errors.ErrServerError.WithError(fmt.Errorf("request failed after %d attempts: %w", retryCount+1, lastErr))
}

// isRetryableStatusCode returns true if the status code indicates a transient error
// that may succeed on retry.
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *Client) GetJSON(ctx context.Context, endpoint string, params url.Values, target interface{}) error {
	start := time.Now()
	endpointType := extractEndpointType(endpoint)

	if c.Cache != nil {
		if cached, found := c.Cache.Get(ctx, endpoint, params); found {
			c.Logger.Debug("Cache hit for endpoint %s", endpoint)
			metrics.RecordCacheHit()

			if err := copyInterface(cached, target); err != nil {
				c.Logger.Warning("Failed to copy cached data: %v", err)

			} else {
				metrics.EndRequest(endpointType, time.Since(start))
				return nil
			}
		}
	}

	metrics.RecordCacheMiss()

	select {
	case ripeLimiter <- struct{}{}:
		metrics.RecordRateLimitWait()
		defer func() { <-ripeLimiter }()
	case <-ctx.Done():
		metrics.RecordRateLimitTimeout()
		return ctx.Err()
	}

	c.Logger.Debug("Cache miss for endpoint %s, making API request", endpoint)

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

	if resp.ContentLength > MaxResponseBodyBytes {
		return errors.ErrServerError.WithError(fmt.Errorf("response body exceeds %d bytes", MaxResponseBodyBytes))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBodyBytes+1))
	if err != nil {
		c.Logger.Error("Failed to read response: %v", err)
		return errors.ErrServerError.WithError(fmt.Errorf("failed to read response: %w", err))
	}
	if int64(len(body)) > MaxResponseBodyBytes {
		return errors.ErrServerError.WithError(fmt.Errorf("response body exceeds %d bytes", MaxResponseBodyBytes))
	}

	if err := json.Unmarshal(body, target); err != nil {
		c.Logger.Error("Failed to decode response: %v", err)
		return errors.ErrServerError.WithError(fmt.Errorf("failed to decode response: %w", err))
	}

	if c.Cache != nil {
		c.Cache.Set(ctx, endpoint, params, target)
		c.Logger.Debug("Cached response for endpoint %s", endpoint)
	}

	c.Logger.Debug("Successfully decoded response")

	return nil
}

func extractEndpointType(endpoint string) string {

	if len(endpoint) > 6 && endpoint[:6] == "/data/" {
		return endpoint[6:]
	}

	return endpoint
}

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
