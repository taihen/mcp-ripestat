
package config

import (
	"os"
	"time"
)


const (

	DefaultBaseURL = "https:


	DefaultTimeout = 30 * time.Second


	DefaultRetryCount = 3


	DefaultRetryWaitTime = 1 * time.Second


	DefaultMaxRetryWaitTime = 30 * time.Second


	DefaultUserAgent = "mcp-ripestat/1.0"


	DefaultSourceApp = "mcp-ripestat"


	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultMaxConnsPerHost     = 100
	DefaultIdleConnTimeout     = 90 * time.Second


	DefaultHTTP2ReadIdleTimeout = 30 * time.Second
	DefaultHTTP2PingTimeout     = 15 * time.Second
)


type Config struct {

	BaseURL string


	Timeout time.Duration


	RetryCount int


	RetryWaitTime time.Duration


	MaxRetryWaitTime time.Duration


	UserAgent string


	SourceApp string


	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration


	ForceHTTP2           bool
	HTTP2ReadIdleTimeout time.Duration
	HTTP2PingTimeout     time.Duration
}


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


		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		MaxConnsPerHost:     DefaultMaxConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,


		ForceHTTP2:           true,
		HTTP2ReadIdleTimeout: DefaultHTTP2ReadIdleTimeout,
		HTTP2PingTimeout:     DefaultHTTP2PingTimeout,
	}
}


func (c *Config) WithBaseURL(baseURL string) *Config {
	if baseURL == "" {
		return c
	}

	newConfig := *c
	newConfig.BaseURL = baseURL

	return &newConfig
}


func (c *Config) WithTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.Timeout = timeout

	return &newConfig
}


func (c *Config) WithRetryCount(retryCount int) *Config {
	if retryCount < 0 {
		return c
	}

	newConfig := *c
	newConfig.RetryCount = retryCount

	return &newConfig
}


func (c *Config) WithRetryWaitTime(retryWaitTime time.Duration) *Config {
	if retryWaitTime <= 0 {
		return c
	}

	newConfig := *c
	newConfig.RetryWaitTime = retryWaitTime

	return &newConfig
}


func (c *Config) WithMaxRetryWaitTime(maxRetryWaitTime time.Duration) *Config {
	if maxRetryWaitTime <= 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxRetryWaitTime = maxRetryWaitTime

	return &newConfig
}


func (c *Config) WithUserAgent(userAgent string) *Config {
	if userAgent == "" {
		return c
	}

	newConfig := *c
	newConfig.UserAgent = userAgent

	return &newConfig
}


func (c *Config) WithSourceApp(sourceApp string) *Config {
	if sourceApp == "" {
		return c
	}

	newConfig := *c
	newConfig.SourceApp = sourceApp

	return &newConfig
}


func (c *Config) WithMaxIdleConns(maxIdleConns int) *Config {
	if maxIdleConns < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxIdleConns = maxIdleConns

	return &newConfig
}


func (c *Config) WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) *Config {
	if maxIdleConnsPerHost < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxIdleConnsPerHost = maxIdleConnsPerHost

	return &newConfig
}


func (c *Config) WithMaxConnsPerHost(maxConnsPerHost int) *Config {
	if maxConnsPerHost < 0 {
		return c
	}

	newConfig := *c
	newConfig.MaxConnsPerHost = maxConnsPerHost

	return &newConfig
}


func (c *Config) WithIdleConnTimeout(idleConnTimeout time.Duration) *Config {
	if idleConnTimeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.IdleConnTimeout = idleConnTimeout

	return &newConfig
}


func (c *Config) WithForceHTTP2(forceHTTP2 bool) *Config {
	newConfig := *c
	newConfig.ForceHTTP2 = forceHTTP2

	return &newConfig
}


func (c *Config) WithHTTP2ReadIdleTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.HTTP2ReadIdleTimeout = timeout

	return &newConfig
}


func (c *Config) WithHTTP2PingTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		return c
	}

	newConfig := *c
	newConfig.HTTP2PingTimeout = timeout

	return &newConfig
}
