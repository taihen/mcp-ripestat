package cache

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// DefaultMaxEntries is the default maximum number of cache entries.
const DefaultMaxEntries = 1000

// Cache implements a thread-safe LRU cache with TTL support.
type Cache struct {
	mu         sync.RWMutex
	data       map[string]*list.Element
	lruList    *list.List
	ttls       map[string]time.Duration
	maxEntries int
}

// entry represents a cache entry with data, expiration, and key.
type entry struct {
	key       string
	data      interface{}
	expiresAt time.Time
}

// Stats represents cache statistics.
type Stats struct {
	TotalEntries   int `json:"total_entries"`
	ExpiredEntries int `json:"expired_entries"`
	ActiveEntries  int `json:"active_entries"`
	MaxEntries     int `json:"max_entries"`
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

// New creates a new cache with default settings.
// The maximum number of entries can be configured via CACHE_MAX_ENTRIES environment variable.
func New() *Cache {
	return NewWithOptions(DefaultTTLs, getMaxEntriesFromEnv())
}

// NewWithTTLs creates a new cache with custom TTLs and default max entries.
func NewWithTTLs(ttls map[string]time.Duration) *Cache {
	return NewWithOptions(ttls, getMaxEntriesFromEnv())
}

// NewWithOptions creates a new cache with custom TTLs and max entries.
func NewWithOptions(ttls map[string]time.Duration, maxEntries int) *Cache {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxEntries
	}
	return &Cache{
		data:       make(map[string]*list.Element),
		lruList:    list.New(),
		ttls:       ttls,
		maxEntries: maxEntries,
	}
}

// getMaxEntriesFromEnv reads the max entries from environment variable.
func getMaxEntriesFromEnv() int {
	if maxStr := os.Getenv("CACHE_MAX_ENTRIES"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil && max > 0 {
			return max
		}
	}
	return DefaultMaxEntries
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

// Get retrieves a value from the cache.
// It moves the accessed entry to the front of the LRU list.
func (c *Cache) Get(_ context.Context, endpoint string, params url.Values) (interface{}, bool) {
	key := generateKey(endpoint, params)

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		e := elem.Value.(*entry)

		// Check if entry has expired
		if time.Now().After(e.expiresAt) {
			c.removeElementLocked(elem)
			return nil, false
		}

		// Move to front (most recently used)
		c.lruList.MoveToFront(elem)
		return e.data, true
	}

	return nil, false
}

// Set stores a value in the cache.
// If the cache is full, the least recently used entry is evicted.
func (c *Cache) Set(_ context.Context, endpoint string, params url.Values, data interface{}) {
	key := generateKey(endpoint, params)
	endpointType := getEndpointType(endpoint)

	c.mu.Lock()
	defer c.mu.Unlock()

	ttl, exists := c.ttls[endpointType]
	if !exists {
		ttl = 5 * time.Minute
	}

	// If key already exists, update it and move to front
	if elem, ok := c.data[key]; ok {
		c.lruList.MoveToFront(elem)
		e := elem.Value.(*entry)
		e.data = data
		e.expiresAt = time.Now().Add(ttl)
		return
	}

	// Evict oldest entries if at capacity
	for c.lruList.Len() >= c.maxEntries {
		c.evictOldestLocked()
	}

	// Add new entry at front
	e := &entry{
		key:       key,
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
	elem := c.lruList.PushFront(e)
	c.data[key] = elem
}

// Delete removes an entry from the cache.
func (c *Cache) Delete(endpoint string, params url.Values) {
	key := generateKey(endpoint, params)

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		c.removeElementLocked(elem)
	}
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*list.Element)
	c.lruList.Init()
}

// Stats returns cache statistics.
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var expired int
	now := time.Now()

	for elem := c.lruList.Front(); elem != nil; elem = elem.Next() {
		if e, ok := elem.Value.(*entry); ok {
			if now.After(e.expiresAt) {
				expired++
			}
		}
	}

	total := c.lruList.Len()
	return Stats{
		TotalEntries:   total,
		ExpiredEntries: expired,
		ActiveEntries:  total - expired,
		MaxEntries:     c.maxEntries,
	}
}

// CleanupExpired removes all expired entries from the cache.
func (c *Cache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var removed int
	now := time.Now()

	// Collect elements to remove
	var toRemove []*list.Element
	for elem := c.lruList.Front(); elem != nil; elem = elem.Next() {
		if e, ok := elem.Value.(*entry); ok {
			if now.After(e.expiresAt) {
				toRemove = append(toRemove, elem)
			}
		}
	}

	// Remove collected elements
	for _, elem := range toRemove {
		c.removeElementLocked(elem)
		removed++
	}

	return removed
}

// SetTTL sets the TTL for a specific endpoint type.
func (c *Cache) SetTTL(endpointType string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttls[endpointType] = ttl
}

// GetTTL gets the TTL for a specific endpoint type.
func (c *Cache) GetTTL(endpointType string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ttl, exists := c.ttls[endpointType]
	return ttl, exists
}

// SetMaxEntries updates the maximum number of cache entries.
// Entries will be evicted if the new limit is smaller than current size.
func (c *Cache) SetMaxEntries(max int) {
	if max <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxEntries = max

	// Evict entries if over the new limit
	for c.lruList.Len() > c.maxEntries {
		c.evictOldestLocked()
	}
}

// GetMaxEntries returns the maximum number of cache entries.
func (c *Cache) GetMaxEntries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxEntries
}

// String returns a string representation of the cache.
func (c *Cache) String() string {
	stats := c.Stats()
	return fmt.Sprintf("Cache{active: %d, expired: %d, total: %d, max: %d}",
		stats.ActiveEntries, stats.ExpiredEntries, stats.TotalEntries, stats.MaxEntries)
}

// evictOldestLocked removes the oldest (least recently used) entry.
// Must be called with lock held.
func (c *Cache) evictOldestLocked() {
	elem := c.lruList.Back()
	if elem != nil {
		c.removeElementLocked(elem)
	}
}

// removeElementLocked removes an element from the cache.
// Must be called with lock held.
func (c *Cache) removeElementLocked(elem *list.Element) {
	c.lruList.Remove(elem)
	if e, ok := elem.Value.(*entry); ok {
		delete(c.data, e.key)
	}
}

