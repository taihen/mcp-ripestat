// Package cache provides TTL-aware caching for RIPEstat API responses.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"
	"time"
)

// Cache provides TTL-aware caching with endpoint-specific durations.
type Cache struct {
	data sync.Map
	ttls map[string]time.Duration
	mu   sync.RWMutex
}

// entry represents a cached item with expiration.
type entry struct {
	data      interface{}
	expiresAt time.Time
}

// DefaultTTLs provides default cache durations for different endpoints.
var DefaultTTLs = map[string]time.Duration{
	"whois":                24 * time.Hour,   // Highly static
	"network-info":         4 * time.Hour,    // Semi-static
	"as-overview":          4 * time.Hour,    // Semi-static
	"announced-prefixes":   2 * time.Hour,    // Moderately dynamic
	"routing-status":       30 * time.Minute, // Dynamic
	"routing-history":      1 * time.Hour,    // Moderately dynamic
	"rpki-validation":      1 * time.Hour,    // Moderately dynamic
	"rpki-history":         2 * time.Hour,    // Moderately dynamic
	"asn-neighbours":       1 * time.Hour,    // Moderately dynamic
	"country-asns":         4 * time.Hour,    // Semi-static
	"abuse-contact-finder": 24 * time.Hour,   // Highly static
	"bgplay":               2 * time.Minute,  // Very dynamic
	"looking-glass":        1 * time.Minute,  // Very dynamic
	"whats-my-ip":          5 * time.Minute,  // Dynamic but can be cached briefly
}

// New creates a new Cache with default TTLs.
func New() *Cache {
	return NewWithTTLs(DefaultTTLs)
}

// NewWithTTLs creates a new Cache with custom TTL configuration.
func NewWithTTLs(ttls map[string]time.Duration) *Cache {
	return &Cache{
		ttls: ttls,
	}
}

// generateKey creates a cache key from endpoint and parameters.
func generateKey(endpoint string, params url.Values) string {
	// Include params in the key to ensure different parameter sets are cached separately
	key := endpoint
	if len(params) > 0 {
		key += "?" + params.Encode()
	}

	// Hash the key to ensure consistent length and avoid special characters
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// getEndpointType extracts the endpoint type from the full endpoint path.
func getEndpointType(endpoint string) string {
	// Extract the main endpoint type from paths like "/data/network-info"
	if len(endpoint) > 6 && endpoint[:6] == "/data/" {
		return endpoint[6:]
	}

	// For other patterns, use the full endpoint
	return endpoint
}

// Get retrieves a cached value if it exists and hasn't expired.
func (c *Cache) Get(_ context.Context, endpoint string, params url.Values) (interface{}, bool) {
	key := generateKey(endpoint, params)

	if value, ok := c.data.Load(key); ok {
		if entry, ok := value.(entry); ok {
			if time.Now().Before(entry.expiresAt) {
				return entry.data, true
			}
			// Expired entry, remove it
			c.data.Delete(key)
		}
	}

	return nil, false
}

// Set stores a value in the cache with TTL based on endpoint type.
func (c *Cache) Set(_ context.Context, endpoint string, params url.Values, data interface{}) {
	key := generateKey(endpoint, params)
	endpointType := getEndpointType(endpoint)

	c.mu.RLock()
	ttl, exists := c.ttls[endpointType]
	c.mu.RUnlock()

	if !exists {
		// Default TTL for unknown endpoints
		ttl = 5 * time.Minute
	}

	entry := entry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}

	c.data.Store(key, entry)
}

// Delete removes a specific cache entry.
func (c *Cache) Delete(endpoint string, params url.Values) {
	key := generateKey(endpoint, params)
	c.data.Delete(key)
}

// Clear removes all cached entries.
func (c *Cache) Clear() {
	c.data.Range(func(key, _ interface{}) bool {
		c.data.Delete(key)
		return true
	})
}

// Stats returns cache statistics.
func (c *Cache) Stats() Stats {
	var total, expired int
	now := time.Now()

	c.data.Range(func(_, value interface{}) bool {
		total++
		if entry, ok := value.(entry); ok {
			if now.After(entry.expiresAt) {
				expired++
			}
		}
		return true
	})

	return Stats{
		TotalEntries:   total,
		ExpiredEntries: expired,
		ActiveEntries:  total - expired,
	}
}

// Stats provides cache statistics.
type Stats struct {
	TotalEntries   int `json:"total_entries"`
	ExpiredEntries int `json:"expired_entries"`
	ActiveEntries  int `json:"active_entries"`
}

// CleanupExpired removes all expired entries from the cache.
func (c *Cache) CleanupExpired() int {
	var removed int
	now := time.Now()

	c.data.Range(func(key, value interface{}) bool {
		if entry, ok := value.(entry); ok {
			if now.After(entry.expiresAt) {
				c.data.Delete(key)
				removed++
			}
		}
		return true
	})

	return removed
}

// SetTTL updates the TTL for a specific endpoint type.
func (c *Cache) SetTTL(endpointType string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttls[endpointType] = ttl
}

// GetTTL returns the TTL for a specific endpoint type.
func (c *Cache) GetTTL(endpointType string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ttl, exists := c.ttls[endpointType]
	return ttl, exists
}

// String returns a string representation of the cache.
func (c *Cache) String() string {
	stats := c.Stats()
	return fmt.Sprintf("Cache{active: %d, expired: %d, total: %d}",
		stats.ActiveEntries, stats.ExpiredEntries, stats.TotalEntries)
}
