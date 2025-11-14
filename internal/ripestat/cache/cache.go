
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


type Cache struct {
	data sync.Map
	ttls map[string]time.Duration
	mu   sync.RWMutex
}


type entry struct {
	data      interface{}
	expiresAt time.Time
}


var DefaultTTLs = map[string]time.Duration{
	"whois":                24 * time.Hour,
	"network-info":         4 * time.Hour,
	"as-overview":          4 * time.Hour,
	"announced-prefixes":   2 * time.Hour,
	"routing-status":       30 * time.Minute,
	"routing-history":      1 * time.Hour,
	"rpki-validation":      1 * time.Hour,
	"rpki-history":         2 * time.Hour,
	"asn-neighbours":       1 * time.Hour,
	"country-asns":         4 * time.Hour,
	"abuse-contact-finder": 24 * time.Hour,
	"bgplay":               2 * time.Minute,
	"looking-glass":        1 * time.Minute,
	"whats-my-ip":          5 * time.Minute,
}


func New() *Cache {
	return NewWithTTLs(DefaultTTLs)
}


func NewWithTTLs(ttls map[string]time.Duration) *Cache {
	return &Cache{
		ttls: ttls,
	}
}


func generateKey(endpoint string, params url.Values) string {

	key := endpoint
	if len(params) > 0 {
		key += "?" + params.Encode()
	}


	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}


func getEndpointType(endpoint string) string {

	if len(endpoint) > 6 && endpoint[:6] == "/data/" {
		return endpoint[6:]
	}


	return endpoint
}


func (c *Cache) Get(_ context.Context, endpoint string, params url.Values) (interface{}, bool) {
	key := generateKey(endpoint, params)

	if value, ok := c.data.Load(key); ok {
		if entry, ok := value.(entry); ok {
			if time.Now().Before(entry.expiresAt) {
				return entry.data, true
			}

			c.data.Delete(key)
		}
	}

	return nil, false
}


func (c *Cache) Set(_ context.Context, endpoint string, params url.Values, data interface{}) {
	key := generateKey(endpoint, params)
	endpointType := getEndpointType(endpoint)

	c.mu.RLock()
	ttl, exists := c.ttls[endpointType]
	c.mu.RUnlock()

	if !exists {

		ttl = 5 * time.Minute
	}

	entry := entry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}

	c.data.Store(key, entry)
}


func (c *Cache) Delete(endpoint string, params url.Values) {
	key := generateKey(endpoint, params)
	c.data.Delete(key)
}


func (c *Cache) Clear() {
	c.data.Range(func(key, _ interface{}) bool {
		c.data.Delete(key)
		return true
	})
}


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


type Stats struct {
	TotalEntries   int `json:"total_entries"`
	ExpiredEntries int `json:"expired_entries"`
	ActiveEntries  int `json:"active_entries"`
}


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


func (c *Cache) SetTTL(endpointType string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttls[endpointType] = ttl
}


func (c *Cache) GetTTL(endpointType string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ttl, exists := c.ttls[endpointType]
	return ttl, exists
}


func (c *Cache) String() string {
	stats := c.Stats()
	return fmt.Sprintf("Cache{active: %d, expired: %d, total: %d}",
		stats.ActiveEntries, stats.ExpiredEntries, stats.TotalEntries)
}
