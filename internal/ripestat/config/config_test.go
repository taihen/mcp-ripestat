package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL to be %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to be %v, got %v", DefaultTimeout, cfg.Timeout)
	}

	if cfg.RetryCount != DefaultRetryCount {
		t.Errorf("Expected RetryCount to be %d, got %d", DefaultRetryCount, cfg.RetryCount)
	}

	if cfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to be %v, got %v", DefaultRetryWaitTime, cfg.RetryWaitTime)
	}

	if cfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to be %v, got %v", DefaultMaxRetryWaitTime, cfg.MaxRetryWaitTime)
	}

	if cfg.UserAgent != DefaultUserAgent {
		t.Errorf("Expected UserAgent to be %q, got %q", DefaultUserAgent, cfg.UserAgent)
	}
}

func TestConfig_WithBaseURL(t *testing.T) {
	cfg := DefaultConfig()

	newCfg := cfg.WithBaseURL("https://example.com")
	if newCfg.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", newCfg.BaseURL)
	}

	newCfg = cfg.WithBaseURL("")
	if newCfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL to remain %q, got %q", DefaultBaseURL, newCfg.BaseURL)
	}

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected original BaseURL to remain %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}

func TestConfig_WithTimeout(t *testing.T) {
	cfg := DefaultConfig()

	newTimeout := 20 * time.Second
	newCfg := cfg.WithTimeout(newTimeout)

	if newCfg.Timeout != newTimeout {
		t.Errorf("Expected Timeout to be %v, got %v", newTimeout, newCfg.Timeout)
	}

	newCfg = cfg.WithTimeout(0)
	if newCfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to remain %v, got %v", DefaultTimeout, newCfg.Timeout)
	}

	newCfg = cfg.WithTimeout(-1 * time.Second)
	if newCfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to remain %v, got %v", DefaultTimeout, newCfg.Timeout)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Expected original Timeout to remain %v, got %v", DefaultTimeout, cfg.Timeout)
	}
}

func TestConfig_WithRetryCount(t *testing.T) {
	cfg := DefaultConfig()

	newRetryCount := 5
	newCfg := cfg.WithRetryCount(newRetryCount)

	if newCfg.RetryCount != newRetryCount {
		t.Errorf("Expected RetryCount to be %d, got %d", newRetryCount, newCfg.RetryCount)
	}

	newCfg = cfg.WithRetryCount(0)
	if newCfg.RetryCount != 0 {
		t.Errorf("Expected RetryCount to be 0, got %d", newCfg.RetryCount)
	}

	newCfg = cfg.WithRetryCount(-1)
	if newCfg.RetryCount != DefaultRetryCount {
		t.Errorf("Expected RetryCount to remain %d, got %d", DefaultRetryCount, newCfg.RetryCount)
	}

	if cfg.RetryCount != DefaultRetryCount {
		t.Errorf("Expected original RetryCount to remain %d, got %d", DefaultRetryCount, cfg.RetryCount)
	}
}

func TestConfig_WithRetryWaitTime(t *testing.T) {
	cfg := DefaultConfig()

	newRetryWaitTime := 2 * time.Second
	newCfg := cfg.WithRetryWaitTime(newRetryWaitTime)

	if newCfg.RetryWaitTime != newRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to be %v, got %v", newRetryWaitTime, newCfg.RetryWaitTime)
	}

	newCfg = cfg.WithRetryWaitTime(0)
	if newCfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, newCfg.RetryWaitTime)
	}

	newCfg = cfg.WithRetryWaitTime(-1 * time.Second)
	if newCfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, newCfg.RetryWaitTime)
	}

	if cfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected original RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, cfg.RetryWaitTime)
	}
}

func TestConfig_WithMaxRetryWaitTime(t *testing.T) {
	cfg := DefaultConfig()

	newMaxRetryWaitTime := 60 * time.Second
	newCfg := cfg.WithMaxRetryWaitTime(newMaxRetryWaitTime)

	if newCfg.MaxRetryWaitTime != newMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to be %v, got %v", newMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	newCfg = cfg.WithMaxRetryWaitTime(0)
	if newCfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	newCfg = cfg.WithMaxRetryWaitTime(-1 * time.Second)
	if newCfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	if cfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected original MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, cfg.MaxRetryWaitTime)
	}
}

func TestConfig_WithUserAgent(t *testing.T) {
	cfg := DefaultConfig()

	newUserAgent := "test-agent/1.0"
	newCfg := cfg.WithUserAgent(newUserAgent)

	if newCfg.UserAgent != newUserAgent {
		t.Errorf("Expected UserAgent to be %q, got %q", newUserAgent, newCfg.UserAgent)
	}

	newCfg = cfg.WithUserAgent("")
	if newCfg.UserAgent != DefaultUserAgent {
		t.Errorf("Expected UserAgent to remain %q, got %q", DefaultUserAgent, newCfg.UserAgent)
	}

	if cfg.UserAgent != DefaultUserAgent {
		t.Errorf("Expected original UserAgent to remain %q, got %q", DefaultUserAgent, cfg.UserAgent)
	}
}

func TestConfig_ChainedMethods(t *testing.T) {
	cfg := DefaultConfig()

	newCfg := cfg.
		WithBaseURL("https://example.com").
		WithTimeout(20 * time.Second).
		WithRetryCount(5).
		WithRetryWaitTime(2 * time.Second).
		WithMaxRetryWaitTime(60 * time.Second).
		WithUserAgent("test-agent/1.0")

	if newCfg.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", newCfg.BaseURL)
	}

	if newCfg.Timeout != 20*time.Second {
		t.Errorf("Expected Timeout to be %v, got %v", 20*time.Second, newCfg.Timeout)
	}

	if newCfg.RetryCount != 5 {
		t.Errorf("Expected RetryCount to be %d, got %d", 5, newCfg.RetryCount)
	}

	if newCfg.RetryWaitTime != 2*time.Second {
		t.Errorf("Expected RetryWaitTime to be %v, got %v", 2*time.Second, newCfg.RetryWaitTime)
	}

	if newCfg.MaxRetryWaitTime != 60*time.Second {
		t.Errorf("Expected MaxRetryWaitTime to be %v, got %v", 60*time.Second, newCfg.MaxRetryWaitTime)
	}

	if newCfg.UserAgent != "test-agent/1.0" {
		t.Errorf("Expected UserAgent to be %q, got %q", "test-agent/1.0", newCfg.UserAgent)
	}

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected original BaseURL to remain %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}

func TestConfig_WithConnectionPoolSettings(t *testing.T) {
	cfg := DefaultConfig().
		WithMaxIdleConns(50).
		WithMaxIdleConnsPerHost(5).
		WithMaxConnsPerHost(50).
		WithIdleConnTimeout(60 * time.Second)

	if cfg.MaxIdleConns != 50 {
		t.Errorf("Expected MaxIdleConns to be 50, got %d", cfg.MaxIdleConns)
	}

	if cfg.MaxIdleConnsPerHost != 5 {
		t.Errorf("Expected MaxIdleConnsPerHost to be 5, got %d", cfg.MaxIdleConnsPerHost)
	}

	if cfg.MaxConnsPerHost != 50 {
		t.Errorf("Expected MaxConnsPerHost to be 50, got %d", cfg.MaxConnsPerHost)
	}

	if cfg.IdleConnTimeout != 60*time.Second {
		t.Errorf("Expected IdleConnTimeout to be 60s, got %v", cfg.IdleConnTimeout)
	}
}

func TestConfig_WithHTTP2Settings(t *testing.T) {
	cfg := DefaultConfig().
		WithForceHTTP2(true).
		WithHTTP2ReadIdleTimeout(45 * time.Second).
		WithHTTP2PingTimeout(20 * time.Second)

	if !cfg.ForceHTTP2 {
		t.Error("Expected ForceHTTP2 to be true")
	}

	if cfg.HTTP2ReadIdleTimeout != 45*time.Second {
		t.Errorf("Expected HTTP2ReadIdleTimeout to be 45s, got %v", cfg.HTTP2ReadIdleTimeout)
	}

	if cfg.HTTP2PingTimeout != 20*time.Second {
		t.Errorf("Expected HTTP2PingTimeout to be 20s, got %v", cfg.HTTP2PingTimeout)
	}
}

func TestConfig_WithConnectionPoolSettings_InvalidValues(t *testing.T) {
	original := DefaultConfig()

	cfg := original.WithMaxIdleConns(-1)
	if cfg.MaxIdleConns != original.MaxIdleConns {
		t.Error("Expected negative MaxIdleConns to be ignored")
	}

	cfg = original.WithMaxIdleConnsPerHost(-1)
	if cfg.MaxIdleConnsPerHost != original.MaxIdleConnsPerHost {
		t.Error("Expected negative MaxIdleConnsPerHost to be ignored")
	}

	cfg = original.WithMaxConnsPerHost(-1)
	if cfg.MaxConnsPerHost != original.MaxConnsPerHost {
		t.Error("Expected negative MaxConnsPerHost to be ignored")
	}

	cfg = original.WithIdleConnTimeout(-1 * time.Second)
	if cfg.IdleConnTimeout != original.IdleConnTimeout {
		t.Error("Expected negative IdleConnTimeout to be ignored")
	}
}

func TestConfig_WithHTTP2Settings_InvalidValues(t *testing.T) {
	original := DefaultConfig()

	cfg := original.WithHTTP2ReadIdleTimeout(0)
	if cfg.HTTP2ReadIdleTimeout != original.HTTP2ReadIdleTimeout {
		t.Error("Expected zero HTTP2ReadIdleTimeout to be ignored")
	}

	cfg = original.WithHTTP2PingTimeout(-1 * time.Second)
	if cfg.HTTP2PingTimeout != original.HTTP2PingTimeout {
		t.Error("Expected negative HTTP2PingTimeout to be ignored")
	}
}

func TestConfig_PerformanceChainedMethods(t *testing.T) {
	cfg := DefaultConfig().
		WithMaxIdleConns(200).
		WithMaxIdleConnsPerHost(20).
		WithMaxConnsPerHost(200).
		WithIdleConnTimeout(120 * time.Second).
		WithForceHTTP2(true).
		WithHTTP2ReadIdleTimeout(60 * time.Second).
		WithHTTP2PingTimeout(30 * time.Second).
		WithTimeout(45 * time.Second)

	if cfg.MaxIdleConns != 200 {
		t.Errorf("Expected MaxIdleConns to be 200, got %d", cfg.MaxIdleConns)
	}

	if cfg.MaxIdleConnsPerHost != 20 {
		t.Errorf("Expected MaxIdleConnsPerHost to be 20, got %d", cfg.MaxIdleConnsPerHost)
	}

	if cfg.MaxConnsPerHost != 200 {
		t.Errorf("Expected MaxConnsPerHost to be 200, got %d", cfg.MaxConnsPerHost)
	}

	if cfg.IdleConnTimeout != 120*time.Second {
		t.Errorf("Expected IdleConnTimeout to be 120s, got %v", cfg.IdleConnTimeout)
	}

	if !cfg.ForceHTTP2 {
		t.Error("Expected ForceHTTP2 to be true")
	}

	if cfg.HTTP2ReadIdleTimeout != 60*time.Second {
		t.Errorf("Expected HTTP2ReadIdleTimeout to be 60s, got %v", cfg.HTTP2ReadIdleTimeout)
	}

	if cfg.HTTP2PingTimeout != 30*time.Second {
		t.Errorf("Expected HTTP2PingTimeout to be 30s, got %v", cfg.HTTP2PingTimeout)
	}

	if cfg.Timeout != 45*time.Second {
		t.Errorf("Expected Timeout to be 45s, got %v", cfg.Timeout)
	}
}

func TestConfig_WithSourceApp(t *testing.T) {
	cfg := DefaultConfig()

	newSourceApp := "test-app/1.0"
	newCfg := cfg.WithSourceApp(newSourceApp)
	if newCfg.SourceApp != newSourceApp {
		t.Errorf("Expected SourceApp to be %q, got %q", newSourceApp, newCfg.SourceApp)
	}

	newCfg = cfg.WithSourceApp("")
	if newCfg.SourceApp != DefaultSourceApp {
		t.Errorf("Expected SourceApp to remain %q, got %q", DefaultSourceApp, newCfg.SourceApp)
	}

	if cfg.SourceApp != DefaultSourceApp {
		t.Errorf("Expected original SourceApp to remain %q, got %q", DefaultSourceApp, cfg.SourceApp)
	}
}
