package consolidated

import (
	"context"
	"fmt"
	"testing"
	"time"
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

func TestNewDirectExecutor(t *testing.T) {
	executor := NewDirectExecutor()
	if executor == nil {
		t.Error("NewDirectExecutor() returned nil")
	}
}

func TestDirectExecutor_ExecuteEndpoint_UnknownEndpoint(t *testing.T) {
	executor := NewDirectExecutor()
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
	executor := NewDirectExecutor()
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
	result, err := NewDirectExecutor().ExecuteEndpoint(
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
	executor := NewDirectExecutor()
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
	executor := NewDirectExecutor()
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
	executor := NewDirectExecutor()
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
	executor := NewDirectExecutor()
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
	executor := NewDirectExecutor()
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
