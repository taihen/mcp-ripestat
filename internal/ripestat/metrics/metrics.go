// Package metrics provides operational metrics for the RIPEstat API client.
package metrics

import (
	"expvar"
	"sync/atomic"
	"time"
)

// Metrics holds all operational metrics for the RIPEstat client.
type Metrics struct {
	// Request metrics
	RequestsTotal    *expvar.Map
	RequestsInFlight *expvar.Int
	RequestDuration  *expvar.Map

	// Cache metrics
	CacheHits           *expvar.Int
	CacheMisses         *expvar.Int
	CacheTotalEntries   *expvar.Int
	CacheExpiredEntries *expvar.Int

	// Rate limiting metrics
	RateLimitWaits    *expvar.Int
	RateLimitTimeouts *expvar.Int

	// Compliance metrics
	DailyRequestCount *expvar.Int
	RequestCounter    *expvar.Map

	// Internal counters for calculations
	inFlightCount  int64
	dailyResetTime time.Time
}

// Global metrics instance.
var globalMetrics *Metrics

func init() {
	globalMetrics = NewMetrics()
}

// NewMetrics creates a new Metrics instance with all expvar variables.
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal:       expvar.NewMap("ripe_client_requests_total"),
		RequestsInFlight:    expvar.NewInt("ripe_client_in_flight_requests"),
		RequestDuration:     expvar.NewMap("ripe_client_request_duration_seconds"),
		CacheHits:           expvar.NewInt("ripe_cache_hits_total"),
		CacheMisses:         expvar.NewInt("ripe_cache_misses_total"),
		CacheTotalEntries:   expvar.NewInt("ripe_cache_total_entries"),
		CacheExpiredEntries: expvar.NewInt("ripe_cache_expired_entries"),
		RateLimitWaits:      expvar.NewInt("ripe_rate_limit_waits_total"),
		RateLimitTimeouts:   expvar.NewInt("ripe_rate_limit_timeouts_total"),
		DailyRequestCount:   expvar.NewInt("ripe_daily_request_count"),
		RequestCounter:      expvar.NewMap("ripe_request_counter"),
		dailyResetTime:      time.Now().Add(24 * time.Hour),
	}

	return m
}

// Global functions that operate on the global metrics instance

// RecordRequest increments the request counter for a specific endpoint and status.
func RecordRequest(endpoint, status string) {
	key := endpoint + "_" + status
	globalMetrics.RequestsTotal.Add(key, 1)

	// Also increment daily request count for compliance monitoring
	recordDailyRequest()
}

// recordDailyRequest tracks daily request count for RIPE compliance.
func recordDailyRequest() {
	now := time.Now()

	// Check if we need to reset the daily counter
	if now.After(globalMetrics.dailyResetTime) {
		globalMetrics.DailyRequestCount.Set(0)
		globalMetrics.dailyResetTime = now.Add(24 * time.Hour)
	}

	globalMetrics.DailyRequestCount.Add(1)

	// Also track by date for historical purposes
	dateKey := now.Format("2006-01-02")
	globalMetrics.RequestCounter.Add(dateKey, 1)
}

// StartRequest increments the in-flight request counter.
func StartRequest() {
	atomic.AddInt64(&globalMetrics.inFlightCount, 1)
	globalMetrics.RequestsInFlight.Set(atomic.LoadInt64(&globalMetrics.inFlightCount))
}

// EndRequest decrements the in-flight request counter and records duration.
func EndRequest(endpoint string, duration time.Duration) {
	atomic.AddInt64(&globalMetrics.inFlightCount, -1)
	globalMetrics.RequestsInFlight.Set(atomic.LoadInt64(&globalMetrics.inFlightCount))

	// Record duration in milliseconds
	durationMs := float64(duration.Nanoseconds()) / 1e6
	globalMetrics.RequestDuration.Add(endpoint, int64(durationMs))
}

// RecordCacheHit increments the cache hit counter.
func RecordCacheHit() {
	globalMetrics.CacheHits.Add(1)
}

// RecordCacheMiss increments the cache miss counter.
func RecordCacheMiss() {
	globalMetrics.CacheMisses.Add(1)
}

// UpdateCacheStats updates cache entry counts.
func UpdateCacheStats(total, expired int) {
	globalMetrics.CacheTotalEntries.Set(int64(total))
	globalMetrics.CacheExpiredEntries.Set(int64(expired))
}

// RecordRateLimitWait increments the rate limit wait counter.
func RecordRateLimitWait() {
	globalMetrics.RateLimitWaits.Add(1)
}

// RecordRateLimitTimeout increments the rate limit timeout counter.
func RecordRateLimitTimeout() {
	globalMetrics.RateLimitTimeouts.Add(1)
}

// GetMetrics returns the global metrics instance.
func GetMetrics() *Metrics {
	return globalMetrics
}

// GetInFlightCount returns the current number of in-flight requests.
func GetInFlightCount() int64 {
	return atomic.LoadInt64(&globalMetrics.inFlightCount)
}

// GetDailyRequestCount returns the current daily request count.
func GetDailyRequestCount() int64 {
	// Check if we need to reset the counter first
	now := time.Now()
	if now.After(globalMetrics.dailyResetTime) {
		globalMetrics.DailyRequestCount.Set(0)
		globalMetrics.dailyResetTime = now.Add(24 * time.Hour)
	}

	return globalMetrics.DailyRequestCount.Value()
}

// Summary returns a summary of key metrics.
func Summary() map[string]interface{} {
	return map[string]interface{}{
		"requests_in_flight":    GetInFlightCount(),
		"daily_request_count":   GetDailyRequestCount(),
		"cache_hits":            globalMetrics.CacheHits.Value(),
		"cache_misses":          globalMetrics.CacheMisses.Value(),
		"cache_total_entries":   globalMetrics.CacheTotalEntries.Value(),
		"cache_expired_entries": globalMetrics.CacheExpiredEntries.Value(),
		"rate_limit_waits":      globalMetrics.RateLimitWaits.Value(),
		"rate_limit_timeouts":   globalMetrics.RateLimitTimeouts.Value(),
	}
}
