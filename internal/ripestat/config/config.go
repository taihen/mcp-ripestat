// Package config provides centralized configuration for the RIPEstat API client.
package config

import (
	"time"
)

// Default configuration values.
const (
	// DefaultBaseURL is the base URL for the RIPEstat API.
	DefaultBaseURL = "https://stat.ripe.net"

	// DefaultTimeout is the default timeout for HTTP requests.
	DefaultTimeout = 30 * time.Second

	// DefaultRetryCount is the default number of retries for failed requests.
	DefaultRetryCount = 3

	// DefaultRetryWaitTime is the default wait time between retries.
	DefaultRetryWaitTime = 1 * time.Second

	// DefaultMaxRetryWaitTime is the default maximum wait time between retries.
	DefaultMaxRetryWaitTime = 30 * time.Second

	// DefaultUserAgent is the default User-Agent header value for HTTP requests.
	DefaultUserAgent = "mcp-ripestat/1.0"
)

// Config represents the configuration for the RIPEstat API client.
type Config struct {
	// BaseURL is the base URL for the RIPEstat API.
	BaseURL string

	// Timeout is the timeout for HTTP requests.
	Timeout time.Duration

	// RetryCount is the number of retries for failed requests.
	RetryCount int

	// RetryWaitTime is the wait time between retries.
	RetryWaitTime time.Duration

	// MaxRetryWaitTime is the maximum wait time between retries.
	MaxRetryWaitTime time.Duration

	// UserAgent is the User-Agent header value for HTTP requests.
	UserAgent string
}

// DefaultConfig returns a new Config with default settings.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:          DefaultBaseURL,
		Timeout:          DefaultTimeout,
		RetryCount:       DefaultRetryCount,
		RetryWaitTime:    DefaultRetryWaitTime,
		MaxRetryWaitTime: DefaultMaxRetryWaitTime,
		UserAgent:        DefaultUserAgent,
	}
}

// WithBaseURL returns a new Config with the specified base URL.
func (c *Config) WithBaseURL(baseURL string) *Config {
	if baseURL == "" {
		return c
	}

	newConfig := *c
	newConfig.BaseURL = baseURL

	return &newConfig
}

// WithTimeout returns a new Config with the specified timeout.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.Timeout = timeout

	return &newConfig
}

// WithRetryCount returns a new Config with the specified retry count.
func (c *Config) WithRetryCount(retryCount int) *Config {
	if retryCount < 0 {
		return c
	}

	newConfig := *c
	newConfig.RetryCount = retryCount

	return &newConfig
}

// WithRetryWaitTime returns a new Config with the specified retry wait time.
func (c *Config) WithRetryWaitTime(retryWaitTime time.Duration) *Config {
	if retryWaitTime <= 0 {
		return c
	}

	newConfig := *c
	newConfig.RetryWaitTime = retryWaitTime

	return &newConfig
}

// WithMaxRetryWaitTime returns a new Config with the specified maximum retry wait time.
func (c *Config) WithMaxRetryWaitTime(maxRetryWaitTime time.Duration) *Config {
	if maxRetryWaitTime <= 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxRetryWaitTime = maxRetryWaitTime

	return &newConfig
}

// WithUserAgent returns a new Config with the specified User-Agent.
func (c *Config) WithUserAgent(userAgent string) *Config {
	if userAgent == "" {
		return c
	}

	newConfig := *c
	newConfig.UserAgent = userAgent

	return &newConfig
}
