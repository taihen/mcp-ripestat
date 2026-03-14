package metrics

import (
	"expvar"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	RequestsTotal    *expvar.Map
	RequestsInFlight *expvar.Int
	RequestDuration  *expvar.Map

	CacheHits           *expvar.Int
	CacheMisses         *expvar.Int
	CacheTotalEntries   *expvar.Int
	CacheExpiredEntries *expvar.Int

	RateLimitWaits    *expvar.Int
	RateLimitTimeouts *expvar.Int

	DailyRequestCount *expvar.Int
	RequestCounter    *expvar.Map

	inFlightCount  int64
	dailyResetTime time.Time
	dailyMu        sync.Mutex
}

var globalMetrics *Metrics

func init() {
	globalMetrics = NewMetrics()
}

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

func RecordRequest(endpoint, status string) {
	key := endpoint + "_" + status
	globalMetrics.RequestsTotal.Add(key, 1)

	recordDailyRequest()
}

func recordDailyRequest() {
	now := time.Now()

	globalMetrics.dailyMu.Lock()
	defer globalMetrics.dailyMu.Unlock()

	if now.After(globalMetrics.dailyResetTime) {
		globalMetrics.DailyRequestCount.Set(0)
		globalMetrics.dailyResetTime = now.Add(24 * time.Hour)
	}

	globalMetrics.DailyRequestCount.Add(1)

	dateKey := now.Format("2006-01-02")
	globalMetrics.RequestCounter.Add(dateKey, 1)
}

func StartRequest() {
	atomic.AddInt64(&globalMetrics.inFlightCount, 1)
	globalMetrics.RequestsInFlight.Set(atomic.LoadInt64(&globalMetrics.inFlightCount))
}

func EndRequest(endpoint string, duration time.Duration) {
	atomic.AddInt64(&globalMetrics.inFlightCount, -1)
	globalMetrics.RequestsInFlight.Set(atomic.LoadInt64(&globalMetrics.inFlightCount))

	durationMs := float64(duration.Nanoseconds()) / 1e6
	globalMetrics.RequestDuration.Add(endpoint, int64(durationMs))
}

func RecordCacheHit() {
	globalMetrics.CacheHits.Add(1)
}

func RecordCacheMiss() {
	globalMetrics.CacheMisses.Add(1)
}

func UpdateCacheStats(total, expired int) {
	globalMetrics.CacheTotalEntries.Set(int64(total))
	globalMetrics.CacheExpiredEntries.Set(int64(expired))
}

func RecordRateLimitWait() {
	globalMetrics.RateLimitWaits.Add(1)
}

func RecordRateLimitTimeout() {
	globalMetrics.RateLimitTimeouts.Add(1)
}

func GetMetrics() *Metrics {
	return globalMetrics
}

func GetInFlightCount() int64 {
	return atomic.LoadInt64(&globalMetrics.inFlightCount)
}

func GetDailyRequestCount() int64 {

	now := time.Now()

	globalMetrics.dailyMu.Lock()
	defer globalMetrics.dailyMu.Unlock()

	if now.After(globalMetrics.dailyResetTime) {
		globalMetrics.DailyRequestCount.Set(0)
		globalMetrics.dailyResetTime = now.Add(24 * time.Hour)
	}

	return globalMetrics.DailyRequestCount.Value()
}

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
