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

	// Test with valid URL
	newCfg := cfg.WithBaseURL("https://example.com")
	if newCfg.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", newCfg.BaseURL)
	}

	// Test with empty URL (should not change)
	newCfg = cfg.WithBaseURL("")
	if newCfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL to remain %q, got %q", DefaultBaseURL, newCfg.BaseURL)
	}

	// Verify original config is unchanged
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected original BaseURL to remain %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}

func TestConfig_WithTimeout(t *testing.T) {
	cfg := DefaultConfig()

	// Test with valid timeout
	newTimeout := 20 * time.Second
	newCfg := cfg.WithTimeout(newTimeout)

	if newCfg.Timeout != newTimeout {
		t.Errorf("Expected Timeout to be %v, got %v", newTimeout, newCfg.Timeout)
	}

	// Test with zero timeout (should not change)
	newCfg = cfg.WithTimeout(0)
	if newCfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to remain %v, got %v", DefaultTimeout, newCfg.Timeout)
	}

	// Test with negative timeout (should not change)
	newCfg = cfg.WithTimeout(-1 * time.Second)
	if newCfg.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout to remain %v, got %v", DefaultTimeout, newCfg.Timeout)
	}

	// Verify original config is unchanged
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Expected original Timeout to remain %v, got %v", DefaultTimeout, cfg.Timeout)
	}
}

func TestConfig_WithRetryCount(t *testing.T) {
	cfg := DefaultConfig()

	// Test with valid retry count
	newRetryCount := 5
	newCfg := cfg.WithRetryCount(newRetryCount)

	if newCfg.RetryCount != newRetryCount {
		t.Errorf("Expected RetryCount to be %d, got %d", newRetryCount, newCfg.RetryCount)
	}

	// Test with zero retry count
	newCfg = cfg.WithRetryCount(0)
	if newCfg.RetryCount != 0 {
		t.Errorf("Expected RetryCount to be 0, got %d", newCfg.RetryCount)
	}

	// Test with negative retry count (should not change)
	newCfg = cfg.WithRetryCount(-1)
	if newCfg.RetryCount != DefaultRetryCount {
		t.Errorf("Expected RetryCount to remain %d, got %d", DefaultRetryCount, newCfg.RetryCount)
	}

	// Verify original config is unchanged
	if cfg.RetryCount != DefaultRetryCount {
		t.Errorf("Expected original RetryCount to remain %d, got %d", DefaultRetryCount, cfg.RetryCount)
	}
}

func TestConfig_WithRetryWaitTime(t *testing.T) {
	cfg := DefaultConfig()

	// Test with valid retry wait time
	newRetryWaitTime := 2 * time.Second
	newCfg := cfg.WithRetryWaitTime(newRetryWaitTime)

	if newCfg.RetryWaitTime != newRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to be %v, got %v", newRetryWaitTime, newCfg.RetryWaitTime)
	}

	// Test with zero retry wait time (should not change)
	newCfg = cfg.WithRetryWaitTime(0)
	if newCfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, newCfg.RetryWaitTime)
	}

	// Test with negative retry wait time (should not change)
	newCfg = cfg.WithRetryWaitTime(-1 * time.Second)
	if newCfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, newCfg.RetryWaitTime)
	}

	// Verify original config is unchanged
	if cfg.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected original RetryWaitTime to remain %v, got %v", DefaultRetryWaitTime, cfg.RetryWaitTime)
	}
}

func TestConfig_WithMaxRetryWaitTime(t *testing.T) {
	cfg := DefaultConfig()

	// Test with valid max retry wait time
	newMaxRetryWaitTime := 60 * time.Second
	newCfg := cfg.WithMaxRetryWaitTime(newMaxRetryWaitTime)

	if newCfg.MaxRetryWaitTime != newMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to be %v, got %v", newMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	// Test with zero max retry wait time (should not change)
	newCfg = cfg.WithMaxRetryWaitTime(0)
	if newCfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	// Test with negative max retry wait time (should not change)
	newCfg = cfg.WithMaxRetryWaitTime(-1 * time.Second)
	if newCfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, newCfg.MaxRetryWaitTime)
	}

	// Verify original config is unchanged
	if cfg.MaxRetryWaitTime != DefaultMaxRetryWaitTime {
		t.Errorf("Expected original MaxRetryWaitTime to remain %v, got %v", DefaultMaxRetryWaitTime, cfg.MaxRetryWaitTime)
	}
}

func TestConfig_WithUserAgent(t *testing.T) {
	cfg := DefaultConfig()

	// Test with valid user agent
	newUserAgent := "test-agent/1.0"
	newCfg := cfg.WithUserAgent(newUserAgent)

	if newCfg.UserAgent != newUserAgent {
		t.Errorf("Expected UserAgent to be %q, got %q", newUserAgent, newCfg.UserAgent)
	}

	// Test with empty user agent (should not change)
	newCfg = cfg.WithUserAgent("")
	if newCfg.UserAgent != DefaultUserAgent {
		t.Errorf("Expected UserAgent to remain %q, got %q", DefaultUserAgent, newCfg.UserAgent)
	}

	// Verify original config is unchanged
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

	// Verify original config is unchanged
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("Expected original BaseURL to remain %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
}
