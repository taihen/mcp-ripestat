package metrics

import (
	"testing"
	"time"
)

func TestGlobalMetrics(t *testing.T) {
	m := GetMetrics()
	if m == nil {
		t.Fatal("Expected global metrics to be non-nil")
	}

	if m.RequestsTotal == nil {
		t.Error("Expected RequestsTotal to be non-nil")
	}
	if m.RequestsInFlight == nil {
		t.Error("Expected RequestsInFlight to be non-nil")
	}
	if m.CacheHits == nil {
		t.Error("Expected CacheHits to be non-nil")
	}
	if m.CacheMisses == nil {
		t.Error("Expected CacheMisses to be non-nil")
	}
}

func TestRecordRequest(t *testing.T) {
	// Use existing global metrics

	endpoint := "whois"
	status := "200"

	// Record a request
	RecordRequest(endpoint, status)

	// Check that it was recorded
	key := endpoint + "_" + status
	value := globalMetrics.RequestsTotal.Get(key)
	if value == nil {
		t.Fatal("Expected request to be recorded")
	}

	// Should be 1
	if value.String() != "1" {
		t.Errorf("Expected recorded value to be 1, got %s", value.String())
	}

	// Record another request
	RecordRequest(endpoint, status)

	// Should be 2
	value = globalMetrics.RequestsTotal.Get(key)
	if value.String() != "2" {
		t.Errorf("Expected recorded value to be 2, got %s", value.String())
	}
}

func TestStartEndRequest(t *testing.T) {
	// Use existing global metrics
	initialCount := GetInFlightCount()

	// Note initial count
	t.Logf("Initial in-flight count: %d", initialCount)

	// Start a request
	StartRequest()

	// Should be initial + 1
	if GetInFlightCount() != initialCount+1 {
		t.Errorf("Expected in-flight count to be %d, got %d", initialCount+1, GetInFlightCount())
	}

	// End the request
	endpoint := "whois"
	duration := 100 * time.Millisecond
	EndRequest(endpoint, duration)

	// Should be back to initial
	if GetInFlightCount() != initialCount {
		t.Errorf("Expected in-flight count to be %d after end, got %d", initialCount, GetInFlightCount())
	}

	// Check that duration was recorded
	durationValue := globalMetrics.RequestDuration.Get(endpoint)
	if durationValue == nil {
		t.Fatal("Expected duration to be recorded")
	}

	// Should be around 100ms (converted to ms)
	if durationValue.String() != "100" {
		t.Errorf("Expected duration to be 100ms, got %s", durationValue.String())
	}
}

func TestMultipleInFlightRequests(t *testing.T) {
	// Use existing global metrics
	initialCount := GetInFlightCount()

	// Start multiple requests
	StartRequest()
	StartRequest()
	StartRequest()

	// Should be initial + 3
	expectedCount := initialCount + 3
	if GetInFlightCount() != expectedCount {
		t.Errorf("Expected in-flight count to be %d, got %d", expectedCount, GetInFlightCount())
	}

	// End one request
	EndRequest("test", 50*time.Millisecond)

	// Should be initial + 2
	expectedCount = initialCount + 2
	if GetInFlightCount() != expectedCount {
		t.Errorf("Expected in-flight count to be %d, got %d", expectedCount, GetInFlightCount())
	}

	// End remaining requests
	EndRequest("test", 50*time.Millisecond)
	EndRequest("test", 50*time.Millisecond)

	// Should be back to initial
	if GetInFlightCount() != initialCount {
		t.Errorf("Expected in-flight count to be %d, got %d", initialCount, GetInFlightCount())
	}
}

func TestCacheMetrics(t *testing.T) {
	// Use existing global metrics
	initialHits := globalMetrics.CacheHits.Value()
	initialMisses := globalMetrics.CacheMisses.Value()

	// Record cache hits and misses
	RecordCacheHit()
	RecordCacheHit()
	RecordCacheMiss()

	// Check values
	if globalMetrics.CacheHits.Value() != initialHits+2 {
		t.Errorf("Expected cache hits to be %d, got %d", initialHits+2, globalMetrics.CacheHits.Value())
	}
	if globalMetrics.CacheMisses.Value() != initialMisses+1 {
		t.Errorf("Expected cache misses to be %d, got %d", initialMisses+1, globalMetrics.CacheMisses.Value())
	}
}

func TestUpdateCacheStats(t *testing.T) {
	// Use existing global metrics

	// Update cache stats
	total := 100
	expired := 10
	UpdateCacheStats(total, expired)

	// Check values
	if globalMetrics.CacheTotalEntries.Value() != int64(total) {
		t.Errorf("Expected cache total entries to be %d, got %d", total, globalMetrics.CacheTotalEntries.Value())
	}
	if globalMetrics.CacheExpiredEntries.Value() != int64(expired) {
		t.Errorf("Expected cache expired entries to be %d, got %d", expired, globalMetrics.CacheExpiredEntries.Value())
	}
}

func TestRateLimitMetrics(t *testing.T) {
	// Use existing global metrics

	// Initially should be 0
	if globalMetrics.RateLimitWaits.Value() != 0 {
		t.Errorf("Expected initial rate limit waits to be 0, got %d", globalMetrics.RateLimitWaits.Value())
	}
	if globalMetrics.RateLimitTimeouts.Value() != 0 {
		t.Errorf("Expected initial rate limit timeouts to be 0, got %d", globalMetrics.RateLimitTimeouts.Value())
	}

	// Record rate limit events
	RecordRateLimitWait()
	RecordRateLimitWait()
	RecordRateLimitTimeout()

	// Check values
	if globalMetrics.RateLimitWaits.Value() != 2 {
		t.Errorf("Expected rate limit waits to be 2, got %d", globalMetrics.RateLimitWaits.Value())
	}
	if globalMetrics.RateLimitTimeouts.Value() != 1 {
		t.Errorf("Expected rate limit timeouts to be 1, got %d", globalMetrics.RateLimitTimeouts.Value())
	}
}

func TestGetMetrics(t *testing.T) {
	m := GetMetrics()
	if m == nil {
		t.Fatal("Expected metrics to be non-nil")
	}

	// Should be the same instance as globalMetrics
	if m != globalMetrics {
		t.Error("Expected GetMetrics to return global metrics instance")
	}
}

func TestSummary(t *testing.T) {
	// Use existing global metrics
	initialInFlight := GetInFlightCount()
	initialCacheHits := globalMetrics.CacheHits.Value()
	initialCacheMisses := globalMetrics.CacheMisses.Value()
	initialRateLimitWaits := globalMetrics.RateLimitWaits.Value()

	// Add some test data
	RecordCacheHit()
	RecordCacheMiss()
	RecordRateLimitWait()
	StartRequest()

	summary := Summary()
	if summary == nil {
		t.Fatal("Expected summary to be non-nil")
	}

	// Check that all expected keys are present
	expectedKeys := []string{
		"requests_in_flight",
		"cache_hits",
		"cache_misses",
		"cache_total_entries",
		"cache_expired_entries",
		"rate_limit_waits",
		"rate_limit_timeouts",
	}

	for _, key := range expectedKeys {
		if _, exists := summary[key]; !exists {
			t.Errorf("Expected summary to contain key %s", key)
		}
	}

	// Check some specific values (accounting for previous test runs)
	if summary["requests_in_flight"] != initialInFlight+1 {
		t.Errorf("Expected requests_in_flight to be %d, got %v", initialInFlight+1, summary["requests_in_flight"])
	}
	if summary["cache_hits"] != initialCacheHits+1 {
		t.Errorf("Expected cache_hits to be %d, got %v", initialCacheHits+1, summary["cache_hits"])
	}
	if summary["cache_misses"] != initialCacheMisses+1 {
		t.Errorf("Expected cache_misses to be %d, got %v", initialCacheMisses+1, summary["cache_misses"])
	}
	if summary["rate_limit_waits"] != initialRateLimitWaits+1 {
		t.Errorf("Expected rate_limit_waits to be %d, got %v", initialRateLimitWaits+1, summary["rate_limit_waits"])
	}

	// Clean up the in-flight request we added
	EndRequest("test-summary", 1*time.Millisecond)
}

func TestDurationRecording(t *testing.T) {
	// Use existing global metrics

	endpoint := "test-endpoint"

	// Test different durations
	durations := []time.Duration{
		1 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		1000 * time.Millisecond,
	}

	for _, duration := range durations {
		StartRequest()
		EndRequest(endpoint, duration)

		// Check that duration was recorded (it accumulates)
		durationValue := globalMetrics.RequestDuration.Get(endpoint)
		if durationValue == nil {
			t.Fatalf("Expected duration to be recorded for %v", duration)
		}
	}
}

func TestConcurrentMetrics(t *testing.T) {
	// Use existing global metrics
	initialInFlight := GetInFlightCount()
	initialCacheHits := globalMetrics.CacheHits.Value()

	// Simulate concurrent requests
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			StartRequest()
			time.Sleep(1 * time.Millisecond) // Simulate some work
			EndRequest("concurrent-test", 1*time.Millisecond)
			RecordCacheHit()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check final state - should be back to initial
	if GetInFlightCount() != initialInFlight {
		t.Errorf("Expected in-flight count to be %d after concurrent test, got %d", initialInFlight, GetInFlightCount())
	}

	// Cache hits should increase by the number of goroutines
	if globalMetrics.CacheHits.Value() != initialCacheHits+int64(numGoroutines) {
		t.Errorf("Expected cache hits to be %d, got %d", initialCacheHits+int64(numGoroutines), globalMetrics.CacheHits.Value())
	}
}
