
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

	EndpointPath = "/data/asn-neighbours/data.json"

	CacheTTL = 15 * time.Minute
)


type cacheEntry struct {
	response  *APIResponse
	timestamp time.Time
}


type cacheKey struct {
	resource  string
	queryTime string
	lod       int
}


type Client struct {
	client *client.Client
	cache  map[cacheKey]*cacheEntry
	mutex  sync.RWMutex
}


func NewClient(c *client.Client) *Client {
	if c == nil {
		c = client.DefaultClient()
	}

	return &Client{
		client: c,
		cache:  make(map[cacheKey]*cacheEntry),
	}
}


func DefaultClient() *Client {
	return NewClient(nil)
}


func (c *Client) Get(ctx context.Context, resource string, lod int, queryTime string) (*APIResponse, error) {
	if resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	if lod < 0 || lod > 1 {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("lod parameter must be 0 or 1"))
	}


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


	apiResponse := &APIResponse{
		Resource:        response.Data.Resource,
		QueryTime:       response.Data.QueryStartTime,
		NeighbourCounts: response.Data.NeighbourCounts,
		Neighbours:      response.Data.Neighbours,
		FetchedAt:       response.Time,
	}


	if apiResponse.Neighbours == nil {
		apiResponse.Neighbours = []Neighbour{}
	}


	c.setCached(key, apiResponse)

	return apiResponse, nil
}


func (c *Client) getCached(key cacheKey) *APIResponse {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}


	if time.Since(entry.timestamp) > CacheTTL {

		delete(c.cache, key)
		return nil
	}

	return entry.response
}


func (c *Client) setCached(key cacheKey, response *APIResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &cacheEntry{
		response:  response,
		timestamp: time.Now(),
	}
}


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


func GetASNNeighbours(ctx context.Context, resource string, lod int, queryTime string) (*APIResponse, error) {
	return DefaultClient().Get(ctx, resource, lod, queryTime)
}
