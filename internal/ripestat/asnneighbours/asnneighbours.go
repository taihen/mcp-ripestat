// Package asnneighbours provides access to the RIPEstat asn-neighbours API.
package asnneighbours

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
)

const (
	// EndpointPath is the path to the RIPEstat data API for ASN neighbours.
	EndpointPath = "/data/asn-neighbours/data.json"
	// CacheTTL is the time-to-live for cached responses (15 minutes).
	CacheTTL = 15 * time.Minute
)

// cacheEntry represents a cached API response.
type cacheEntry struct {
	response  *APIResponse
	timestamp time.Time
}

// cacheKey represents the key for caching responses.
type cacheKey struct {
	resource  string
	queryTime string
	lod       int
}

// Client provides methods to interact with the RIPEstat asn-neighbours API.
type Client struct {
	client *client.Client
	cache  map[cacheKey]*cacheEntry
	mutex  sync.RWMutex
}

// NewClient creates a new Client for the RIPEstat asn-neighbours API.
func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{
		client: c,
		cache:  make(map[cacheKey]*cacheEntry),
	}
}

// DefaultClient returns a new Client with default settings.
func DefaultClient() *Client {
	return NewClient(nil)
}

// Get fetches ASN neighbour information for the specified resource.
func (c *Client) Get(ctx context.Context, resource string, lod int, queryTime string) (*APIResponse, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	if lod < 0 || lod > 1 {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("lod parameter must be 0 or 1"))
	}

	// Check cache first
	key := cacheKey{
		resource:  resource,
		queryTime: queryTime,
		lod:       lod,
	}

	if cached := c.getCached(key); cached != nil {
		return cached, nil
	}

	params := url.Values{}
	params.Set("resource", resource)
	params.Set("lod", strconv.Itoa(lod))

	if queryTime != "" {
		params.Set("query_time", queryTime)
	}

	var response Response
	if err := c.client.GetJSON(ctx, EndpointPath, params, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get ASN neighbours: %w", err))
	}

	// Transform the response to the expected API format
	apiResponse := &APIResponse{
		Resource:        response.Data.Resource,
		QueryTime:       response.Data.QueryStartTime,
		NeighbourCounts: response.Data.NeighbourCounts,
		Neighbours:      response.Data.Neighbours,
		FetchedAt:       response.Time,
	}

	// Ensure neighbours is never nil, use empty slice instead
	if apiResponse.Neighbours == nil {
		apiResponse.Neighbours = []Neighbour{}
	}

	// Cache the response
	c.setCached(key, apiResponse)

	return apiResponse, nil
}

// getCached retrieves a cached response if it exists and is still valid.
func (c *Client) getCached(key cacheKey) *APIResponse {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(entry.timestamp) > CacheTTL {
		// Cache expired, remove it
		delete(c.cache, key)
		return nil
	}

	return entry.response
}

// setCached stores a response in the cache.
func (c *Client) setCached(key cacheKey, response *APIResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &cacheEntry{
		response:  response,
		timestamp: time.Now(),
	}
}

// clearExpiredCache removes expired entries from the cache.
func (c *Client) clearExpiredCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.Sub(entry.timestamp) > CacheTTL {
			delete(c.cache, key)
		}
	}
}

// GetASNNeighbours is a convenience function that uses the default client to get ASN neighbours.
func GetASNNeighbours(ctx context.Context, resource string, lod int, queryTime string) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, lod, queryTime)
}
