package consolidated

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/config"
)

func TestTimeframeBounds(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		timeframe string
		wantStart time.Time
		wantErr   bool
	}{
		{timeframe: TimeframeCurrent, wantStart: now.Add(-time.Hour)},
		{timeframe: Timeframe1Day, wantErr: true},
		{timeframe: Timeframe1Week, wantErr: true},
		{timeframe: Timeframe1Month, wantErr: true},
		{timeframe: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.timeframe, func(t *testing.T) {
			start, end, err := timeframeBounds(tt.timeframe, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("timeframeBounds() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !start.Equal(tt.wantStart) || !end.Equal(now) {
				t.Errorf("timeframeBounds() = (%v, %v), want (%v, %v)", start, end, tt.wantStart, now)
			}
		})
	}
}

func TestBGPUpdateOptions(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)

	opts, err := bgpUpdateOptions(nil, now)
	if err != nil || opts != nil {
		t.Fatalf("bgpUpdateOptions(nil) = (%v, %v), want (nil, nil)", opts, err)
	}

	opts, err = bgpUpdateOptions(map[string]interface{}{"timeframe": TimeframeCurrent}, now)
	if err != nil {
		t.Fatalf("bgpUpdateOptions(current) error = %v", err)
	}
	if opts == nil || opts.StartTime != "2026-08-21T09:00:00Z" || opts.EndTime != "2026-08-21T10:00:00Z" {
		t.Errorf("bgpUpdateOptions(current) = %#v", opts)
	}
}

func testExecutor(t *testing.T) *DirectExecutor {
	t.Helper()
	return NewDirectExecutor(client.NewWithConfig(config.DefaultConfig(), nil))
}

func TestNewDirectExecutor_NilPanics(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("NewDirectExecutor(nil) did not panic")
		}
		msg, ok := recovered.(string)
		if !ok || msg != "mcp: ripe client is nil" {
			t.Fatalf("panic = %v", recovered)
		}
	}()
	NewDirectExecutor(nil)
}

func TestNewDirectExecutor_KeepsProvidedClient(t *testing.T) {
	ripe := client.New("https://stat.ripe.net", nil)
	executor := NewDirectExecutor(ripe)
	if executor.RIPEClient() != ripe {
		t.Fatal("executor stored a different client")
	}
}

func TestNewDirectExecutor(t *testing.T) {
	executor := testExecutor(t)
	if executor == nil {
		t.Fatal("NewDirectExecutor() returned nil")
	}
	if executor.RIPEClient() == nil {
		t.Fatal("executor client is nil")
	}
}

func TestDirectExecutor_ExecuteEndpoint_UnknownEndpoint(t *testing.T) {
	executor := testExecutor(t)
	ctx := context.Background()

	result, err := executor.ExecuteEndpoint(ctx, "unknownEndpoint", "8.8.8.8", nil)
	if err == nil {
		t.Error("ExecuteEndpoint() expected error for unknown endpoint")
	}
	if result != nil {
		t.Error("ExecuteEndpoint() expected nil result for unknown endpoint")
	}
}

func TestDirectExecutor_ExecuteEndpoint_EmptyResourceValidation(t *testing.T) {
	executor := testExecutor(t)
	tests := []struct {
		endpoint string
		params   map[string]interface{}
	}{
		{endpoint: "getNetworkInfo"},
		{endpoint: "getASOverview"},
		{endpoint: "getAnnouncedPrefixes"},
		{endpoint: "getRelatedPrefixes"},
		{endpoint: "getRoutingStatus"},
		{endpoint: "getWhois"},
		{endpoint: "getAbuseContactFinder"},
		{endpoint: "getRPKIValidation"},
		{endpoint: "getRPKIHistory"},
		{endpoint: "getASNNeighbours"},
		{endpoint: "getLookingGlass"},
		{endpoint: "getCountryASNs"},
		{endpoint: "getBGPlay"},
		{endpoint: "getBGPUpdates"},
		{endpoint: "getBGPState"},
		{endpoint: "getPrefixRoutingConsistency"},
		{endpoint: "getPrefixOverview"},
		{endpoint: "getAddressSpaceHierarchy"},
		{endpoint: "getAllocationHistory"},
		{endpoint: "getASPathLength"},
		{endpoint: "getASRoutingConsistency"},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			result, err := executor.ExecuteEndpoint(context.Background(), tt.endpoint, "", tt.params)
			if err == nil {
				t.Fatalf("ExecuteEndpoint(%s) expected validation error, result=%v", tt.endpoint, result)
			}
		})
	}
}

func TestDirectExecutor_HandleBGPUpdates_InvalidTimeframe(t *testing.T) {
	result, err := testExecutor(t).ExecuteEndpoint(
		context.Background(),
		"getBGPUpdates",
		"AS15169",
		map[string]interface{}{"timeframe": "invalid"},
	)
	if err == nil || result != nil {
		t.Fatalf("ExecuteEndpoint(getBGPUpdates) = (%v, %v), want nil result and error", result, err)
	}
}

func TestDirectExecutor_HandleRoutingHistory(t *testing.T) {
	executor := testExecutor(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "no optional params",
			params: map[string]interface{}{},
		},
		{
			name: "with start_time",
			params: map[string]interface{}{
				"start_time": "2023-01-01T00:00:00",
			},
		},
		{
			name: "with end_time",
			params: map[string]interface{}{
				"end_time": "2023-12-31T23:59:59",
			},
		},
		{
			name: "with max_results as int",
			params: map[string]interface{}{
				"max_results": 10,
			},
		},
		{
			name: "with max_results as string",
			params: map[string]interface{}{
				"max_results": "10",
			},
		},
		{
			name: "with max_results as float64",
			params: map[string]interface{}{
				"max_results": 10.0,
			},
		},
		{
			name: "with max_results as int64",
			params: map[string]interface{}{
				"max_results": int64(10),
			},
		},
		{
			name: "with all optional params",
			params: map[string]interface{}{
				"start_time":  "2023-01-01T00:00:00",
				"end_time":    "2023-12-31T23:59:59",
				"max_results": 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := executor.ExecuteEndpoint(ctx, "getRoutingHistory", "8.8.8.0/24", tt.params)

			if err != nil && err.Error() == "unknown endpoint: getRoutingHistory" {
				t.Errorf("ExecuteEndpoint() should handle getRoutingHistory")
			}
		})
	}
}

func TestDirectExecutor_HandleRPKIValidation(t *testing.T) {
	executor := testExecutor(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name:        "missing prefix parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "with prefix parameter",
			params: map[string]interface{}{
				"prefix": "8.8.8.0/24",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.ExecuteEndpoint(ctx, "getRPKIValidation", "AS15169", tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("ExecuteEndpoint() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestDirectExecutor_HandleASNNeighbours(t *testing.T) {
	executor := testExecutor(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "no optional params",
			params: map[string]interface{}{},
		},
		{
			name: "with lod as int",
			params: map[string]interface{}{
				"lod": 1,
			},
		},
		{
			name: "with lod as string",
			params: map[string]interface{}{
				"lod": "1",
			},
		},
		{
			name: "with lod as float64",
			params: map[string]interface{}{
				"lod": 1.0,
			},
		},
		{
			name: "with lod as int64",
			params: map[string]interface{}{
				"lod": int64(1),
			},
		},
		{
			name: "with query_time",
			params: map[string]interface{}{
				"query_time": "2023-01-01T00:00:00",
			},
		},
		{
			name: "with both lod and query_time",
			params: map[string]interface{}{
				"lod":        1,
				"query_time": "2023-01-01T00:00:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.ExecuteEndpoint(ctx, "getASNNeighbours", "AS15169", tt.params)

			if err != nil && err.Error() == "unknown endpoint: getASNNeighbours" {
				t.Errorf("ExecuteEndpoint() should handle getASNNeighbours")
			}
		})
	}
}

func testEndpointWithIntParam(t *testing.T, endpointName, resource, paramName string, paramValue int) {
	executor := testExecutor(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "no optional params",
			params: map[string]interface{}{},
		},
		{
			name: "with " + paramName + " as int",
			params: map[string]interface{}{
				paramName: paramValue,
			},
		},
		{
			name: "with " + paramName + " as string",
			params: map[string]interface{}{
				paramName: fmt.Sprintf("%d", paramValue),
			},
		},
		{
			name: "with " + paramName + " as float64",
			params: map[string]interface{}{
				paramName: float64(paramValue),
			},
		},
		{
			name: "with " + paramName + " as int64",
			params: map[string]interface{}{
				paramName: int64(paramValue),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.ExecuteEndpoint(ctx, endpointName, resource, tt.params)

			if err != nil && err.Error() == "unknown endpoint: "+endpointName {
				t.Errorf("ExecuteEndpoint() should handle %s", endpointName)
			}
		})
	}
}

func TestDirectExecutor_HandleLookingGlass(t *testing.T) {
	testEndpointWithIntParam(t, "getLookingGlass", "8.8.8.8", "look_back_limit", 10)
}

func TestDirectExecutor_HandleCountryASNs(t *testing.T) {
	testEndpointWithIntParam(t, "getCountryASNs", "US", "lod", 1)
}

func TestDirectExecutor_HandleBGPState(t *testing.T) {
	executor := testExecutor(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "no optional params",
			params: map[string]interface{}{},
		},
		{
			name: "with timestamp",
			params: map[string]interface{}{
				"timestamp": "2023-01-01T00:00:00",
			},
		},
		{
			name: "with rrcs",
			params: map[string]interface{}{
				"rrcs": "rrc00,rrc01",
			},
		},
		{
			name: "with unix_timestamps as bool",
			params: map[string]interface{}{
				"unix_timestamps": true,
			},
		},
		{
			name: "with all optional params",
			params: map[string]interface{}{
				"timestamp":       "2023-01-01T00:00:00",
				"rrcs":            "rrc00,rrc01",
				"unix_timestamps": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.ExecuteEndpoint(ctx, "getBGPState", "8.8.8.0/24", tt.params)

			if err != nil && err.Error() == "unknown endpoint: getBGPState" {
				t.Errorf("ExecuteEndpoint() should handle getBGPState")
			}
		})
	}
}

func TestGetOptionalStringParam(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		key    string
		want   string
	}{
		{
			name:   "existing string parameter",
			params: map[string]interface{}{"key": "value"},
			key:    "key",
			want:   "value",
		},
		{
			name:   "missing parameter",
			params: map[string]interface{}{},
			key:    "key",
			want:   "",
		},
		{
			name:   "non-string parameter",
			params: map[string]interface{}{"key": 123},
			key:    "key",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOptionalStringParam(tt.params, tt.key)
			if got != tt.want {
				t.Errorf("getOptionalStringParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectExecutor_SharedCacheSkipsRepeatUpstream(t *testing.T) {
	var networkHits, whoisHits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/data/network-info/data.json":
			networkHits++
			_, _ = w.Write([]byte(`{
				"status":"ok","status_code":200,
				"data":{"asns":["1205"],"prefix":"140.78.0.0/16"}
			}`))
		case "/data/whois/data.json":
			whoisHits++
			_, _ = w.Write([]byte(`{
				"status":"ok","status_code":200,
				"data":{
					"records":[[{"key":"NetName","value":"LEVEL3"}]],
					"irr_records":[],
					"authorities":["whois.arin.net"],
					"resource":"8.8.8.8"
				}
			}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	ripe := client.New(ts.URL, ts.Client())
	ripe.RetryConfig.RetryCount = 0
	executor := NewDirectExecutor(ripe)
	ctx := context.Background()

	if _, err := executor.ExecuteEndpoint(ctx, "getNetworkInfo", "140.78.90.50", nil); err != nil {
		t.Fatalf("getNetworkInfo: %v", err)
	}
	if _, err := executor.ExecuteEndpoint(ctx, "getWhois", "8.8.8.8", nil); err != nil {
		t.Fatalf("getWhois: %v", err)
	}
	if got := ripe.Cache.Stats().TotalEntries; got != 2 {
		t.Fatalf("shared cache entries = %d, want 2 (network-info and whois on one client)", got)
	}
	if _, err := executor.ExecuteEndpoint(ctx, "getNetworkInfo", "140.78.90.50", nil); err != nil {
		t.Fatalf("repeated getNetworkInfo: %v", err)
	}
	if networkHits != 1 {
		t.Fatalf("network-info upstream hits = %d, want 1", networkHits)
	}
	if whoisHits != 1 {
		t.Fatalf("whois upstream hits = %d, want 1", whoisHits)
	}
}

func TestExecutorSourceAvoidsPackageHelpers(t *testing.T) {
	src, err := os.ReadFile("executor.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	forbidden := []string{
		"DefaultClient(",
		"networkinfo.GetNetworkInfo",
		"asoverview.GetASOverview",
		"announcedprefixes.GetAnnouncedPrefixes",
		"relatedprefixes.GetRelatedPrefixes",
		"routingstatus.GetRoutingStatus",
		"routinghistory.GetRoutingHistory",
		"routinghistory.GetRoutingHistoryWithOptions",
		"whois.GetWhois",
		"abusecontactfinder.GetAbuseContactFinder",
		"rpkivalidation.GetRPKIValidation",
		"rpkihistory.GetRPKIHistory",
		"asnneighbours.GetASNNeighbours",
		"lookingglass.GetLookingGlass",
		"countryasns.GetCountryASNs",
		"bgplay.GetBGPlay",
		"bgpupdates.GetBGPUpdates",
		"bgpstate.DefaultClient",
		"prefixroutingconsistency.GetPrefixRoutingConsistency",
		"prefixoverview.GetPrefixOverview",
		"addressspacehierarchy.GetAddressSpaceHierarchy",
		"allocationhistory.GetAllocationHistory",
		"aspathlength.GetASPathLength",
		"asroutingconsistency.GetASRoutingConsistency",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Errorf("executor.go still contains %q", needle)
		}
	}
	required := []string{
		"NewClient(de.ripe)",
		"networkinfo.NewClient(de.ripe)",
		"whois.NewClient(de.ripe)",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Errorf("executor.go missing %q", needle)
		}
	}
}

func TestDirectExecutor_ZeroValuePanics(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("zero DirectExecutor did not panic")
		}
		msg, ok := recovered.(string)
		if !ok || msg != "mcp: ripe client is nil" {
			t.Fatalf("panic = %v", recovered)
		}
	}()
	var executor DirectExecutor
	_, _ = executor.ExecuteEndpoint(context.Background(), "no-such-endpoint", "8.8.8.8", nil)
}

func TestDirectExecutor_DistinctEndpointsKeepSeparateCacheKeys(t *testing.T) {
	var networkHits, asHits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/data/network-info/data.json":
			networkHits++
			_, _ = w.Write([]byte(`{"status":"ok","status_code":200,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"}}`))
		case "/data/as-overview/data.json":
			asHits++
			_, _ = w.Write([]byte(`{"status":"ok","status_code":200,"data":{"asn":1205,"holder":"example"}}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	ripe := client.New(ts.URL, ts.Client())
	ripe.RetryConfig.RetryCount = 0
	executor := NewDirectExecutor(ripe)
	ctx := context.Background()
	resource := "AS1205"

	if _, err := executor.ExecuteEndpoint(ctx, "getNetworkInfo", resource, nil); err != nil {
		t.Fatalf("getNetworkInfo: %v", err)
	}
	if _, err := executor.ExecuteEndpoint(ctx, "getASOverview", resource, nil); err != nil {
		t.Fatalf("getASOverview: %v", err)
	}
	if networkHits != 1 || asHits != 1 {
		t.Fatalf("hits network=%d as=%d, want 1 each (distinct cache keys)", networkHits, asHits)
	}
	if got := ripe.Cache.Stats().TotalEntries; got != 2 {
		t.Fatalf("cache entries = %d, want 2", got)
	}
}

func TestDirectExecutor_ConcurrentCallsShareClient(t *testing.T) {
	var networkHits int
	var mu sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/network-info/data.json" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		mu.Lock()
		networkHits++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","status_code":200,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"}}`))
	}))
	defer ts.Close()

	ripe := client.New(ts.URL, ts.Client())
	ripe.RetryConfig.RetryCount = 0
	executor := NewDirectExecutor(ripe)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := executor.ExecuteEndpoint(ctx, "getNetworkInfo", "140.78.90.50", nil); err != nil {
				t.Errorf("getNetworkInfo: %v", err)
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	hits := networkHits
	mu.Unlock()
	if hits < 1 {
		t.Fatal("expected at least one upstream hit")
	}
	if got := ripe.Cache.Stats().TotalEntries; got != 1 {
		t.Fatalf("shared cache entries = %d, want 1", got)
	}
}

func TestGetOptionalIntParam(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		key    string
		want   int
	}{
		{
			name:   "existing int parameter",
			params: map[string]interface{}{"key": 10},
			key:    "key",
			want:   10,
		},
		{
			name:   "existing int64 parameter",
			params: map[string]interface{}{"key": int64(20)},
			key:    "key",
			want:   20,
		},
		{
			name:   "existing float64 parameter",
			params: map[string]interface{}{"key": 30.0},
			key:    "key",
			want:   30,
		},
		{
			name:   "existing string parameter (numeric)",
			params: map[string]interface{}{"key": "40"},
			key:    "key",
			want:   40,
		},
		{
			name:   "existing string parameter (non-numeric)",
			params: map[string]interface{}{"key": "abc"},
			key:    "key",
			want:   0,
		},
		{
			name:   "missing parameter",
			params: map[string]interface{}{},
			key:    "key",
			want:   0,
		},
		{
			name:   "unsupported type",
			params: map[string]interface{}{"key": []string{"a", "b"}},
			key:    "key",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getOptionalIntParam(tt.params, tt.key)
			if got != tt.want {
				t.Errorf("getOptionalIntParam() = %v, want %v", got, tt.want)
			}
		})
	}
}
