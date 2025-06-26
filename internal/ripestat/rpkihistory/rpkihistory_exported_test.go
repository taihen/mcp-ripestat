package rpkihistory_test

import (
	"context"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkihistory"
)

func TestGetRPKIHistory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid IPv4 prefix",
			resource: "193.0.22.0/23",
			wantErr:  false,
		},
		{
			name:     "valid IPv6 prefix",
			resource: "2001:7fb:ff00::/48",
			wantErr:  false,
		},
		{
			name:     "empty resource",
			resource: "",
			wantErr:  true,
		},
		{
			name:     "invalid resource",
			resource: "invalid-resource",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := rpkihistory.GetRPKIHistory(ctx, tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetRPKIHistory() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRPKIHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("GetRPKIHistory() result is nil")
				return
			}

			if result.Data.Timeseries == nil {
				t.Error("GetRPKIHistory() result.Data.Timeseries is nil")
				return
			}

			// Basic validation of response structure
			if result.DataCallName != "rpki-history" {
				t.Errorf("GetRPKIHistory() DataCallName = %s, want rpki-history", result.DataCallName)
			}

			// If timeseries has data, validate structure
			if len(result.Data.Timeseries) > 0 {
				entry := result.Data.Timeseries[0]
				if entry.Prefix == "" {
					t.Error("GetRPKIHistory() first entry has empty prefix")
				}
				if entry.Time == "" {
					t.Error("GetRPKIHistory() first entry has empty time")
				}
				if entry.Family == 0 {
					t.Error("GetRPKIHistory() first entry has zero family")
				}
			}
		})
	}
}

func TestGetRPKIHistory_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err := rpkihistory.GetRPKIHistory(ctx, "193.0.22.0/23")
	if err == nil {
		t.Error("GetRPKIHistory() expected timeout error, got nil")
	}
}
