// Package config provides centralized configuration for the RIPEstat API client.
package config

import (
	"os"
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

	// DefaultSourceApp is the default sourceapp parameter for RIPE API compliance.
	DefaultSourceApp = "mcp-ripestat"

	// Connection pool defaults for optimal performance.
	DefaultMaxIdleConns        = 100              // Maximum idle connections across all hosts
	DefaultMaxIdleConnsPerHost = 10               // Maximum idle connections per host
	DefaultMaxConnsPerHost     = 100              // Maximum connections per host
	DefaultIdleConnTimeout     = 90 * time.Second // Idle connection timeout

	// HTTP/2 defaults.
	DefaultHTTP2ReadIdleTimeout = 30 * time.Second // HTTP/2 read idle timeout
	DefaultHTTP2PingTimeout     = 15 * time.Second // HTTP/2 ping timeout
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

	// SourceApp is the sourceapp parameter for RIPE API compliance.
	SourceApp string

	// Connection pool settings for HTTP client optimization
	MaxIdleConns        int           // Maximum number of idle connections across all hosts
	MaxIdleConnsPerHost int           // Maximum number of idle connections per host
	MaxConnsPerHost     int           // Maximum number of connections per host
	IdleConnTimeout     time.Duration // Maximum time an idle connection will remain idle

	// HTTP/2 settings
	ForceHTTP2           bool          // Force HTTP/2 usage (with fallback to HTTP/1.1)
	HTTP2ReadIdleTimeout time.Duration // HTTP/2 read idle timeout
	HTTP2PingTimeout     time.Duration // HTTP/2 ping timeout
}

// DefaultConfig returns a new Config with default settings.
func DefaultConfig() *Config {
	sourceApp := os.Getenv("RIPE_SOURCE_APP")
	if sourceApp == "" {
		sourceApp = DefaultSourceApp
	}

	return &Config{
		BaseURL:          DefaultBaseURL,
		Timeout:          DefaultTimeout,
		RetryCount:       DefaultRetryCount,
		RetryWaitTime:    DefaultRetryWaitTime,
		MaxRetryWaitTime: DefaultMaxRetryWaitTime,
		UserAgent:        DefaultUserAgent,
		SourceApp:        sourceApp,

		// Connection pool settings
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		MaxConnsPerHost:     DefaultMaxConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,

		// HTTP/2 settings
		ForceHTTP2:           true, // Enable HTTP/2 by default
		HTTP2ReadIdleTimeout: DefaultHTTP2ReadIdleTimeout,
		HTTP2PingTimeout:     DefaultHTTP2PingTimeout,
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

// WithSourceApp returns a new Config with the specified SourceApp.
func (c *Config) WithSourceApp(sourceApp string) *Config {
	if sourceApp == "" {
		return c
	}

	newConfig := *c
	newConfig.SourceApp = sourceApp

	return &newConfig
}

// WithMaxIdleConns returns a new Config with the specified maximum idle connections.
func (c *Config) WithMaxIdleConns(maxIdleConns int) *Config {
	if maxIdleConns < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxIdleConns = maxIdleConns

	return &newConfig
}

// WithMaxIdleConnsPerHost returns a new Config with the specified maximum idle connections per host.
func (c *Config) WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) *Config {
	if maxIdleConnsPerHost < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxIdleConnsPerHost = maxIdleConnsPerHost

	return &newConfig
}

// WithMaxConnsPerHost returns a new Config with the specified maximum connections per host.
func (c *Config) WithMaxConnsPerHost(maxConnsPerHost int) *Config {
	if maxConnsPerHost < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxConnsPerHost = maxConnsPerHost

	return &newConfig
}

// WithIdleConnTimeout returns a new Config with the specified idle connection timeout.
func (c *Config) WithIdleConnTimeout(idleConnTimeout time.Duration) *Config {
	if idleConnTimeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.IdleConnTimeout = idleConnTimeout

	return &newConfig
}

// WithForceHTTP2 returns a new Config with the specified HTTP/2 force setting.
func (c *Config) WithForceHTTP2(forceHTTP2 bool) *Config {
	newConfig := *c
	newConfig.ForceHTTP2 = forceHTTP2

	return &newConfig
}

// WithHTTP2ReadIdleTimeout returns a new Config with the specified HTTP/2 read idle timeout.
func (c *Config) WithHTTP2ReadIdleTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.HTTP2ReadIdleTimeout = timeout

	return &newConfig
}

// WithHTTP2PingTimeout returns a new Config with the specified HTTP/2 ping timeout.
func (c *Config) WithHTTP2PingTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.HTTP2PingTimeout = timeout

	return &newConfig
}
