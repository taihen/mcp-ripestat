package cache

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cache := New()
	if cache == nil {
		t.Fatal("Expected cache to be non-nil")
	}

	// Check that default TTLs are set
	ttl, exists := cache.GetTTL("whois")
	if !exists {
		t.Error("Expected whois TTL to exist in default configuration")
	}
	if ttl != 24*time.Hour {
		t.Errorf("Expected whois TTL to be 24h, got %v", ttl)
	}
}

func TestNewWithTTLs(t *testing.T) {
	customTTLs := map[string]time.Duration{
		"test-endpoint": 30 * time.Second,
	}

	cache := NewWithTTLs(customTTLs)
	if cache == nil {
		t.Fatal("Expected cache to be non-nil")
	}

	ttl, exists := cache.GetTTL("test-endpoint")
	if !exists {
		t.Error("Expected custom TTL to exist")
	}
	if ttl != 30*time.Second {
		t.Errorf("Expected custom TTL to be 30s, got %v", ttl)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	cache := New()
	ctx := context.Background()

	endpoint := "/data/whois"
	params := url.Values{}
	params.Set("resource", "192.168.1.1")

	testData := map[string]interface{}{
		"result": "test-data",
	}

	// Set data in cache
	cache.Set(ctx, endpoint, params, testData)

	// Get data from cache
	result, found := cache.Get(ctx, endpoint, params)
	if !found {
		t.Fatal("Expected to find cached data")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if resultMap["result"] != "test-data" {
		t.Errorf("Expected result to be 'test-data', got %v", resultMap["result"])
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := New()
	ctx := context.Background()

	endpoint := "/data/non-existent"
	params := url.Values{}
	params.Set("resource", "test")

	result, found := cache.Get(ctx, endpoint, params)
	if found {
		t.Errorf("Expected not to find data, but got %v", result)
	}
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
}

func TestCache_Expiration(t *testing.T) {
	customTTLs := map[string]time.Duration{
		"test-endpoint": 50 * time.Millisecond,
	}
	cache := NewWithTTLs(customTTLs)
	ctx := context.Background()

	endpoint := "/data/test-endpoint"
	params := url.Values{}
	params.Set("resource", "test")

	testData := "test-data"

	// Set data in cache
	cache.Set(ctx, endpoint, params, testData)

	// Should be found immediately
	result, found := cache.Get(ctx, endpoint, params)
	if !found {
		t.Fatal("Expected to find cached data immediately")
	}
	if result != testData {
		t.Errorf("Expected result to be %v, got %v", testData, result)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should not be found after expiration
	result, found = cache.Get(ctx, endpoint, params)
	if found {
		t.Errorf("Expected not to find expired data, but got %v", result)
	}
}

func TestCache_DifferentParams(t *testing.T) {
	cache := New()
	ctx := context.Background()

	endpoint := "/data/whois"

	params1 := url.Values{}
	params1.Set("resource", "192.168.1.1")

	params2 := url.Values{}
	params2.Set("resource", "192.168.1.2")

	testData1 := "data-for-ip1"
	testData2 := "data-for-ip2"

	// Set different data for different params
	cache.Set(ctx, endpoint, params1, testData1)
	cache.Set(ctx, endpoint, params2, testData2)

	// Get data for first params
	result1, found1 := cache.Get(ctx, endpoint, params1)
	if !found1 {
		t.Fatal("Expected to find cached data for params1")
	}
	if result1 != testData1 {
		t.Errorf("Expected result1 to be %v, got %v", testData1, result1)
	}

	// Get data for second params
	result2, found2 := cache.Get(ctx, endpoint, params2)
	if !found2 {
		t.Fatal("Expected to find cached data for params2")
	}
	if result2 != testData2 {
		t.Errorf("Expected result2 to be %v, got %v", testData2, result2)
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New()
	ctx := context.Background()

	endpoint := "/data/whois"
	params := url.Values{}
	params.Set("resource", "test")

	testData := "test-data"

	// Set data in cache
	cache.Set(ctx, endpoint, params, testData)

	// Verify data exists
	_, found := cache.Get(ctx, endpoint, params)
	if !found {
		t.Fatal("Expected to find cached data before deletion")
	}

	// Delete data
	cache.Delete(endpoint, params)

	// Verify data is gone
	result, found := cache.Get(ctx, endpoint, params)
	if found {
		t.Errorf("Expected not to find deleted data, but got %v", result)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := New()
	ctx := context.Background()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		endpoint := "/data/whois"
		params := url.Values{}
		params.Set("resource", string(rune('a'+i)))
		cache.Set(ctx, endpoint, params, i)
	}

	// Verify entries exist
	stats := cache.Stats()
	if stats.TotalEntries != 5 {
		t.Errorf("Expected 5 entries, got %d", stats.TotalEntries)
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	stats = cache.Stats()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.TotalEntries)
	}
}

func TestCache_Stats(t *testing.T) {
	customTTLs := map[string]time.Duration{
		"short-ttl": 10 * time.Millisecond,
		"long-ttl":  1 * time.Hour,
	}
	cache := NewWithTTLs(customTTLs)
	ctx := context.Background()

	// Add entries with different TTLs
	params1 := url.Values{}
	params1.Set("resource", "test1")
	cache.Set(ctx, "/data/short-ttl", params1, "data1")

	params2 := url.Values{}
	params2.Set("resource", "test2")
	cache.Set(ctx, "/data/long-ttl", params2, "data2")

	// Check stats immediately
	stats := cache.Stats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 total entries, got %d", stats.TotalEntries)
	}
	if stats.ActiveEntries != 2 {
		t.Errorf("Expected 2 active entries, got %d", stats.ActiveEntries)
	}
	if stats.ExpiredEntries != 0 {
		t.Errorf("Expected 0 expired entries, got %d", stats.ExpiredEntries)
	}

	// Wait for short TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Check stats after expiration
	stats = cache.Stats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 total entries, got %d", stats.TotalEntries)
	}
	if stats.ActiveEntries != 1 {
		t.Errorf("Expected 1 active entry, got %d", stats.ActiveEntries)
	}
	if stats.ExpiredEntries != 1 {
		t.Errorf("Expected 1 expired entry, got %d", stats.ExpiredEntries)
	}
}

func TestCache_CleanupExpired(t *testing.T) {
	customTTLs := map[string]time.Duration{
		"short-ttl": 10 * time.Millisecond,
		"long-ttl":  1 * time.Hour,
	}
	cache := NewWithTTLs(customTTLs)
	ctx := context.Background()

	// Add entries with different TTLs
	params1 := url.Values{}
	params1.Set("resource", "test1")
	cache.Set(ctx, "/data/short-ttl", params1, "data1")

	params2 := url.Values{}
	params2.Set("resource", "test2")
	cache.Set(ctx, "/data/long-ttl", params2, "data2")

	// Wait for short TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Cleanup expired entries
	removed := cache.CleanupExpired()
	if removed != 1 {
		t.Errorf("Expected to remove 1 expired entry, removed %d", removed)
	}

	// Check stats after cleanup
	stats := cache.Stats()
	if stats.TotalEntries != 1 {
		t.Errorf("Expected 1 total entry after cleanup, got %d", stats.TotalEntries)
	}
	if stats.ActiveEntries != 1 {
		t.Errorf("Expected 1 active entry after cleanup, got %d", stats.ActiveEntries)
	}
	if stats.ExpiredEntries != 0 {
		t.Errorf("Expected 0 expired entries after cleanup, got %d", stats.ExpiredEntries)
	}
}

func TestCache_SetTTL(t *testing.T) {
	cache := New()

	// Set custom TTL
	newTTL := 45 * time.Minute
	cache.SetTTL("custom-endpoint", newTTL)

	// Verify TTL was set
	ttl, exists := cache.GetTTL("custom-endpoint")
	if !exists {
		t.Error("Expected custom TTL to exist")
	}
	if ttl != newTTL {
		t.Errorf("Expected TTL to be %v, got %v", newTTL, ttl)
	}
}

func TestGenerateKey(t *testing.T) {
	endpoint := "/data/whois"

	params1 := url.Values{}
	params1.Set("resource", "192.168.1.1")

	params2 := url.Values{}
	params2.Set("resource", "192.168.1.2")

	// Different params should generate different keys
	key1 := generateKey(endpoint, params1)
	key2 := generateKey(endpoint, params2)

	if key1 == key2 {
		t.Error("Expected different keys for different parameters")
	}

	// Same params should generate same key
	key1Again := generateKey(endpoint, params1)
	if key1 != key1Again {
		t.Error("Expected same key for same parameters")
	}

	// Key should be hex encoded (64 chars for SHA256)
	if len(key1) != 64 {
		t.Errorf("Expected key length to be 64, got %d", len(key1))
	}
}

func TestGetEndpointType(t *testing.T) {
	tests := []struct {
		endpoint string
		expected string
	}{
		{"/data/whois", "whois"},
		{"/data/network-info", "network-info"},
		{"/other/endpoint", "/other/endpoint"},
		{"simple", "simple"},
	}

	for _, test := range tests {
		result := getEndpointType(test.endpoint)
		if result != test.expected {
			t.Errorf("For endpoint %s, expected %s, got %s", test.endpoint, test.expected, result)
		}
	}
}

func TestCache_String(t *testing.T) {
	cache := New()
	ctx := context.Background()

	// Add some data
	params := url.Values{}
	params.Set("resource", "test")
	cache.Set(ctx, "/data/whois", params, "data")

	str := cache.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Should contain relevant information
	if !contains(str, "Cache{") {
		t.Error("Expected string to contain 'Cache{'")
	}
	if !contains(str, "active:") {
		t.Error("Expected string to contain 'active:'")
	}
}

func TestDefaultTTLs(t *testing.T) {
	// Verify that default TTLs are reasonable
	expectedEndpoints := []string{
		"whois", "network-info", "as-overview", "announced-prefixes",
		"routing-status", "routing-history", "rpki-validation", "rpki-history",
		"asn-neighbours", "country-asns", "abuse-contact-finder",
		"bgplay", "looking-glass", "whats-my-ip",
	}

	for _, endpoint := range expectedEndpoints {
		ttl, exists := DefaultTTLs[endpoint]
		if !exists {
			t.Errorf("Expected default TTL for endpoint %s", endpoint)
		}
		if ttl <= 0 {
			t.Errorf("Expected positive TTL for endpoint %s, got %v", endpoint, ttl)
		}
	}

	// Verify ordering makes sense (more static data has longer TTL)
	if DefaultTTLs["whois"] <= DefaultTTLs["bgplay"] {
		t.Error("Expected whois TTL to be longer than bgplay TTL")
	}

	if DefaultTTLs["network-info"] <= DefaultTTLs["looking-glass"] {
		t.Error("Expected network-info TTL to be longer than looking-glass TTL")
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
