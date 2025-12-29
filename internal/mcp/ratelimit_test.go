package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	if config.RequestsPerSecond != 10 {
		t.Errorf("Expected RequestsPerSecond to be 10, got %f", config.RequestsPerSecond)
	}
	if config.BurstSize != 20 {
		t.Errorf("Expected BurstSize to be 20, got %d", config.BurstSize)
	}
	if !config.Enabled {
		t.Error("Expected Enabled to be true by default")
	}
}

func TestNewRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 5,
		BurstSize:         10,
		Enabled:           true,
	}
	limiter := NewRateLimiter(config)

	if limiter == nil {
		t.Fatal("Expected limiter to be created")
	}
	if limiter.config.RequestsPerSecond != 5 {
		t.Errorf("Expected RequestsPerSecond to be 5, got %f", limiter.config.RequestsPerSecond)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:         5,
			Enabled:           true,
		}
		limiter := NewRateLimiter(config)

		// Should allow burst size number of requests
		for i := 0; i < 5; i++ {
			if !limiter.Allow("192.168.1.1") {
				t.Errorf("Expected request %d to be allowed", i+1)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerSecond: 1,
			BurstSize:         2,
			Enabled:           true,
		}
		limiter := NewRateLimiter(config)

		// Use up the burst
		limiter.Allow("192.168.1.1")
		limiter.Allow("192.168.1.1")

		// Next request should be blocked
		if limiter.Allow("192.168.1.1") {
			t.Error("Expected request to be blocked after burst exhausted")
		}
	})

	t.Run("allows all requests when disabled", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerSecond: 1,
			BurstSize:         1,
			Enabled:           false,
		}
		limiter := NewRateLimiter(config)

		// Should allow unlimited requests when disabled
		for i := 0; i < 100; i++ {
			if !limiter.Allow("192.168.1.1") {
				t.Errorf("Expected request %d to be allowed when disabled", i+1)
			}
		}
	})

	t.Run("tracks clients separately", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerSecond: 1,
			BurstSize:         1,
			Enabled:           true,
		}
		limiter := NewRateLimiter(config)

		// Client 1 uses their token
		if !limiter.Allow("192.168.1.1") {
			t.Error("Expected first client request to be allowed")
		}

		// Client 2 should still have their token
		if !limiter.Allow("192.168.1.2") {
			t.Error("Expected second client request to be allowed")
		}

		// Client 1 should be blocked
		if limiter.Allow("192.168.1.1") {
			t.Error("Expected first client to be blocked")
		}
	})

	t.Run("replenishes tokens over time", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerSecond: 100, // Fast replenish for testing
			BurstSize:         1,
			Enabled:           true,
		}
		limiter := NewRateLimiter(config)

		// Use the token
		limiter.Allow("192.168.1.1")

		// Wait for replenishment
		time.Sleep(20 * time.Millisecond)

		// Should have replenished
		if !limiter.Allow("192.168.1.1") {
			t.Error("Expected token to be replenished")
		}
	})
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
		Enabled:           true,
	}
	limiter := NewRateLimiter(config)
	limiter.cleanupInterval = 1 * time.Millisecond

	// Add some clients
	limiter.Allow("192.168.1.1")
	limiter.Allow("192.168.1.2")

	// Verify clients exist
	limiter.mu.Lock()
	if len(limiter.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(limiter.clients))
	}
	limiter.mu.Unlock()
}

func TestRateLimiter_CleanupLocked(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
		Enabled:           true,
	}
	limiter := NewRateLimiter(config)

	// Add clients manually
	limiter.mu.Lock()
	now := time.Now()
	// Add a stale client (older than 10 minutes)
	limiter.clients["stale"] = &clientBucket{
		tokens:     5,
		lastUpdate: now.Add(-15 * time.Minute),
	}
	// Add a fresh client
	limiter.clients["fresh"] = &clientBucket{
		tokens:     5,
		lastUpdate: now,
	}

	// Run cleanup
	limiter.cleanupLocked(now)

	// Check stale was removed
	if _, exists := limiter.clients["stale"]; exists {
		t.Error("Expected stale client to be removed")
	}
	// Check fresh remains
	if _, exists := limiter.clients["fresh"]; !exists {
		t.Error("Expected fresh client to remain")
	}
	limiter.mu.Unlock()
}

func TestRateLimitMiddleware(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		Enabled:           true,
	}
	limiter := NewRateLimiter(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(limiter, handler)

	t.Run("allows first request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("blocks rate limited request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", w.Code)
		}
	})
}

func TestExtractClientIPForRateLimit(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
	}{
		{
			name:       "uses RemoteAddr when no headers",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
		},
		{
			name:       "handles RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			headers:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := extractClientIPForRateLimit(req)
			if ip == "" {
				t.Error("Expected non-empty IP")
			}
		})
	}
}

