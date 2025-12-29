package mcp

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
)

// RateLimiter implements a per-client token bucket rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
	config  RateLimitConfig

	// cleanupInterval is how often to clean up expired buckets
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// RequestsPerSecond is the rate at which tokens are added to the bucket.
	RequestsPerSecond float64

	// BurstSize is the maximum number of tokens (requests) that can be accumulated.
	BurstSize int

	// Enabled controls whether rate limiting is active.
	Enabled bool
}

// clientBucket tracks rate limit state for a single client.
type clientBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// DefaultRateLimitConfig returns the default rate limit configuration.
// It can be overridden via environment variables:
// RATE_LIMIT_ENABLED, RATE_LIMIT_RPS, and RATE_LIMIT_BURST.
func DefaultRateLimitConfig() RateLimitConfig {
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
		Enabled:           true,
	}

	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true" || enabled == "1"
	}

	if rps := os.Getenv("RATE_LIMIT_RPS"); rps != "" {
		if val, err := strconv.ParseFloat(rps, 64); err == nil && val > 0 {
			config.RequestsPerSecond = val
		}
	}

	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if val, err := strconv.Atoi(burst); err == nil && val > 0 {
			config.BurstSize = val
		}
	}

	return config
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		clients:         make(map[string]*clientBucket),
		config:          config,
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
}

// Allow checks if a request from the given client IP should be allowed.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(clientIP string) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Periodic cleanup of stale entries
	if now.Sub(rl.lastCleanup) > rl.cleanupInterval {
		rl.cleanupLocked(now)
		rl.lastCleanup = now
	}

	bucket, exists := rl.clients[clientIP]
	if !exists {
		bucket = &clientBucket{
			tokens:     float64(rl.config.BurstSize),
			lastUpdate: now,
		}
		rl.clients[clientIP] = bucket
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.tokens += elapsed * rl.config.RequestsPerSecond
	bucket.lastUpdate = now

	// Cap tokens at burst size
	if bucket.tokens > float64(rl.config.BurstSize) {
		bucket.tokens = float64(rl.config.BurstSize)
	}

	// Check if we have at least one token
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanupLocked removes stale client entries. Must be called with mutex held.
func (rl *RateLimiter) cleanupLocked(now time.Time) {
	staleThreshold := 10 * time.Minute
	for ip, bucket := range rl.clients {
		if now.Sub(bucket.lastUpdate) > staleThreshold {
			delete(rl.clients, ip)
		}
	}
}

// RateLimitMiddleware wraps an http.Handler with rate limiting.
func RateLimitMiddleware(limiter *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP
		clientIP := extractClientIPForRateLimit(r)

		if !limiter.Allow(clientIP) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractClientIPForRateLimit extracts the client IP for rate limiting purposes.
// It securely extracts the client IP by trusting proxy headers only from configured trusted proxies.
func extractClientIPForRateLimit(r *http.Request) string {
	return whatsmyip.ExtractClientIP(r)
}
