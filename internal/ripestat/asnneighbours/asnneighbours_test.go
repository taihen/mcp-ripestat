package asnneighbours

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		lod            int
		queryTime      string
		serverResponse string
		statusCode     int
		wantResource   string
		wantNeighbours int
		wantErr        bool
	}{
		{
			name:     "successful response with lod 0",
			resource: "AS1205",
			lod:      0,
			serverResponse: `{
				"messages": [["info","Query time has been set to the latest available time (2025-06-23 00:00 UTC)"]],
				"see_also": [],
				"version": "3.2",
				"data_call_name": "asn-neighbours",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"resource": "1205",
					"query_starttime": "2025-06-23T00:00:00",
					"query_endtime": "2025-06-23T00:00:00",
					"latest_time": "2025-06-23T00:00:00",
					"earliest_time": "2018-04-06T00:00:00",
					"neighbour_counts": {
						"left": 1,
						"right": 0,
						"unique": 1,
						"uncertain": 0
					},
					"neighbours": [
						{
							"asn": 1853,
							"type": "left"
						}
					]
				},
				"query_id": "test-query-id",
				"process_time": 25,
				"server_id": "app192",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T21:24:00.704269"
			}`,
			statusCode:     200,
			wantResource:   "1205",
			wantNeighbours: 1,
			wantErr:        false,
		},
		{
			name:     "successful response with lod 1",
			resource: "AS1205",
			lod:      1,
			serverResponse: `{
				"messages": [["info","Query time has been set to the latest available time (2025-06-23 00:00 UTC)"]],
				"see_also": [],
				"version": "3.2",
				"data_call_name": "asn-neighbours",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"resource": "1205",
					"query_starttime": "2025-06-23T00:00:00",
					"query_endtime": "2025-06-23T00:00:00",
					"latest_time": "2025-06-23T00:00:00",
					"earliest_time": "2018-05-26T00:00:00",
					"neighbour_counts": {
						"left": 1,
						"right": 0,
						"unique": 1,
						"uncertain": 0
					},
					"neighbours": [
						{
							"asn": 1853,
							"type": "left",
							"power": 372,
							"v4_peers": 1254,
							"v6_peers": 393
						}
					]
				},
				"query_id": "test-query-id",
				"process_time": 27,
				"server_id": "app176",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T21:24:08.525301"
			}`,
			statusCode:     200,
			wantResource:   "1205",
			wantNeighbours: 1,
			wantErr:        false,
		},
		{
			name:      "successful response with query_time",
			resource:  "AS1205",
			lod:       0,
			queryTime: "2024-01-01T00:00:00",
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "3.2",
				"data_call_name": "asn-neighbours",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"resource": "1205",
					"query_starttime": "2024-01-01T00:00:00",
					"query_endtime": "2024-01-01T00:00:00",
					"latest_time": "2025-06-23T00:00:00",
					"earliest_time": "2018-06-02T00:00:00",
					"neighbour_counts": {
						"left": 1,
						"right": 0,
						"unique": 1,
						"uncertain": 0
					},
					"neighbours": [
						{
							"asn": 1853,
							"type": "left"
						}
					]
				},
				"query_id": "test-query-id",
				"process_time": 69,
				"server_id": "app191",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T21:24:16.506445"
			}`,
			statusCode:     200,
			wantResource:   "1205",
			wantNeighbours: 1,
			wantErr:        false,
		},
		{
			name:     "successful response with empty neighbours",
			resource: "AS65000",
			lod:      0,
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "3.2",
				"data_call_name": "asn-neighbours",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"resource": "65000",
					"query_starttime": "2025-06-23T00:00:00",
					"query_endtime": "2025-06-23T00:00:00",
					"latest_time": "2025-06-23T00:00:00",
					"earliest_time": "2018-04-06T00:00:00",
					"neighbour_counts": {
						"left": 0,
						"right": 0,
						"unique": 0,
						"uncertain": 0
					},
					"neighbours": []
				},
				"query_id": "test-query-id",
				"process_time": 25,
				"server_id": "app192",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T21:24:00.704269"
			}`,
			statusCode:     200,
			wantResource:   "65000",
			wantNeighbours: 0,
			wantErr:        false,
		},
		{
			name:     "successful response with null neighbours",
			resource: "AS65001",
			lod:      0,
			serverResponse: `{
				"messages": [],
				"see_also": [],
				"version": "3.2",
				"data_call_name": "asn-neighbours",
				"data_call_status": "supported",
				"cached": false,
				"data": {
					"resource": "65001",
					"query_starttime": "2025-06-23T00:00:00",
					"query_endtime": "2025-06-23T00:00:00",
					"latest_time": "2025-06-23T00:00:00",
					"earliest_time": "2018-04-06T00:00:00",
					"neighbour_counts": {
						"left": 0,
						"right": 0,
						"unique": 0,
						"uncertain": 0
					},
					"neighbours": null
				},
				"query_id": "test-query-id",
				"process_time": 25,
				"server_id": "app192",
				"build_version": "main-2025.06.23",
				"status": "ok",
				"status_code": 200,
				"time": "2025-06-23T21:24:00.704269"
			}`,
			statusCode:     200,
			wantResource:   "65001",
			wantNeighbours: 0,
			wantErr:        false,
		},
		{
			name:         "empty resource parameter",
			resource:     "",
			lod:          0,
			statusCode:   200,
			wantResource: "",
			wantErr:      true,
		},
		{
			name:         "invalid lod parameter negative",
			resource:     "AS1205",
			lod:          -1,
			statusCode:   200,
			wantResource: "",
			wantErr:      true,
		},
		{
			name:         "invalid lod parameter too high",
			resource:     "AS1205",
			lod:          2,
			statusCode:   200,
			wantResource: "",
			wantErr:      true,
		},
		{
			name:           "server error",
			resource:       "AS1205",
			lod:            0,
			serverResponse: `{"error": "internal server error"}`,
			statusCode:     500,
			wantResource:   "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			httpClient := &http.Client{}
			c := NewClient(client.New(server.URL, httpClient))

			got, err := c.Get(context.Background(), tt.resource, tt.lod, tt.queryTime)

			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Errorf("Client.Get() returned nil response")
					return
				}

				if got.Resource != tt.wantResource {
					t.Errorf("Client.Get() resource = %v, want %v", got.Resource, tt.wantResource)
				}

				if len(got.Neighbours) != tt.wantNeighbours {
					t.Errorf("Client.Get() neighbours length = %v, want %v", len(got.Neighbours), tt.wantNeighbours)
				}

				if got.FetchedAt == "" {
					t.Errorf("Client.Get() FetchedAt is empty")
				}

				// Verify neighbours is never nil
				if got.Neighbours == nil {
					t.Errorf("Client.Get() neighbours should not be nil")
				}
			}
		})
	}
}

func TestClient_Caching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "3.2",
			"data_call_name": "asn-neighbours",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"resource": "1205",
				"query_starttime": "2025-06-23T00:00:00",
				"query_endtime": "2025-06-23T00:00:00",
				"latest_time": "2025-06-23T00:00:00",
				"earliest_time": "2018-04-06T00:00:00",
				"neighbour_counts": {
					"left": 1,
					"right": 0,
					"unique": 1,
					"uncertain": 0
				},
				"neighbours": [
					{
						"asn": 1853,
						"type": "left"
					}
				]
			},
			"query_id": "test-query-id",
			"process_time": 25,
			"server_id": "app192",
			"build_version": "main-2025.06.23",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-23T21:24:00.704269"
		}`))
	}))
	defer server.Close()

	httpClient := &http.Client{}
	c := NewClient(client.New(server.URL, httpClient))

	ctx := context.Background()

	// First call should hit the server
	_, err := c.Get(ctx, "AS1205", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call with same parameters should use cache
	_, err = c.Get(ctx, "AS1205", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 server call (cached), got %d", callCount)
	}

	// Call with different lod should hit server again
	_, err = c.Get(ctx, "AS1205", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 server calls (different lod), got %d", callCount)
	}

	// Call with different query_time should hit server again
	_, err = c.Get(ctx, "AS1205", 0, "2024-01-01T00:00:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 server calls (different query_time), got %d", callCount)
	}
}

func TestClient_CacheExpiration(t *testing.T) {
	// Create a client with a very short cache TTL for testing
	c := &Client{
		client: client.DefaultClient(),
		cache:  make(map[cacheKey]*cacheEntry),
	}

	// Manually add an expired cache entry
	key := cacheKey{resource: "AS1205", queryTime: "", lod: 0}
	expiredEntry := &cacheEntry{
		response: &APIResponse{
			Resource:  "1205",
			QueryTime: "2025-06-23T00:00:00",
			Neighbours: []Neighbour{
				{ASN: 1853, Type: "left"},
			},
			FetchedAt: "2025-06-23T21:24:00.704269",
		},
		timestamp: time.Now().Add(-20 * time.Minute), // Expired (older than 15 minutes)
	}

	c.setCached(key, expiredEntry.response)
	// Manually set the timestamp to simulate expiration
	c.cache[key].timestamp = expiredEntry.timestamp

	// getCached should return nil for expired entry
	cached := c.getCached(key)
	if cached != nil {
		t.Errorf("expected nil for expired cache entry, got %v", cached)
	}

	// The expired entry should be removed from cache
	c.mutex.RLock()
	_, exists := c.cache[key]
	c.mutex.RUnlock()

	if exists {
		t.Errorf("expired cache entry should be removed")
	}
}

func TestGetASNNeighbours(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "3.2",
			"data_call_name": "asn-neighbours",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"resource": "1205",
				"query_starttime": "2025-06-23T00:00:00",
				"query_endtime": "2025-06-23T00:00:00",
				"latest_time": "2025-06-23T00:00:00",
				"earliest_time": "2018-04-06T00:00:00",
				"neighbour_counts": {
					"left": 1,
					"right": 0,
					"unique": 1,
					"uncertain": 0
				},
				"neighbours": [
					{
						"asn": 1853,
						"type": "left"
					}
				]
			},
			"query_id": "test-query-id",
			"process_time": 25,
			"server_id": "app192",
			"build_version": "main-2025.06.23",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-23T21:24:00.704269"
		}`))
	}))
	defer ts.Close()

	// Create a custom client that uses our test server
	customClient := client.New(ts.URL, ts.Client())
	asnNeighboursClient := NewClient(customClient)

	// Test the client directly
	ctx := context.Background()
	resp, err := asnNeighboursClient.Get(ctx, "AS1205", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Resource != "1205" {
		t.Errorf("expected resource 1205, got %v", resp.Resource)
	}
	if len(resp.Neighbours) != 1 {
		t.Errorf("expected 1 neighbour, got %v", len(resp.Neighbours))
	}
	if resp.Neighbours[0].ASN != 1853 {
		t.Errorf("expected ASN 1853, got %v", resp.Neighbours[0].ASN)
	}
	if resp.Neighbours[0].Type != "left" {
		t.Errorf("expected type left, got %v", resp.Neighbours[0].Type)
	}
}

func TestGetASNNeighbours_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetASNNeighbours(ctx, "", 0, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}

	// Test with invalid lod
	_, err = GetASNNeighbours(ctx, "AS1205", 5, "")
	if err == nil {
		t.Fatal("expected error for invalid lod, got nil")
	}
	if !strings.Contains(err.Error(), "lod parameter must be 0 or 1") {
		t.Errorf("expected lod validation error, got %v", err)
	}
}

func TestNewClient(t *testing.T) {
	t.Run("with existing client", func(t *testing.T) {
		existingClient := client.DefaultClient()
		c := NewClient(existingClient)
		if c == nil {
			t.Errorf("NewClient() returned nil")
			return
		}
		if c.client != existingClient {
			t.Errorf("NewClient() did not use provided client")
		}
		if c.cache == nil {
			t.Errorf("NewClient() cache is nil")
		}
	})

	t.Run("with nil client", func(t *testing.T) {
		c := NewClient(nil)
		if c == nil {
			t.Errorf("NewClient() returned nil")
			return
		}
		if c.client == nil {
			t.Errorf("NewClient() client field is nil")
		}
		if c.cache == nil {
			t.Errorf("NewClient() cache is nil")
		}
	})
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c == nil {
		t.Errorf("DefaultClient() returned nil")
		return
	}
	if c.client == nil {
		t.Errorf("DefaultClient() client field is nil")
	}
	if c.cache == nil {
		t.Errorf("DefaultClient() cache is nil")
	}
}

func TestClearExpiredCache(t *testing.T) {
	c := &Client{
		client: client.DefaultClient(),
		cache:  make(map[cacheKey]*cacheEntry),
	}

	// Add some cache entries
	validKey := cacheKey{resource: "AS1205", queryTime: "", lod: 0}
	expiredKey := cacheKey{resource: "AS1853", queryTime: "", lod: 0}

	validEntry := &cacheEntry{
		response:  &APIResponse{Resource: "1205"},
		timestamp: time.Now(),
	}
	expiredEntry := &cacheEntry{
		response:  &APIResponse{Resource: "1853"},
		timestamp: time.Now().Add(-20 * time.Minute),
	}

	c.cache[validKey] = validEntry
	c.cache[expiredKey] = expiredEntry

	// Clear expired cache
	c.clearExpiredCache()

	// Valid entry should still exist
	if _, exists := c.cache[validKey]; !exists {
		t.Errorf("valid cache entry should still exist")
	}

	// Expired entry should be removed
	if _, exists := c.cache[expiredKey]; exists {
		t.Errorf("expired cache entry should be removed")
	}
}
