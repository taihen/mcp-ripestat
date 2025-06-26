package e2e

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkihistory"
)

func TestRPKIHistoryE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	c := client.DefaultClient()
	rpkiHistoryClient := rpkihistory.NewClient(c)

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "RIPE NCC prefix",
			resource: "193.0.22.0/23",
			wantErr:  false,
		},
		{
			name:     "IPv6 prefix",
			resource: "2001:7fb:ff00::/48",
			wantErr:  false,
		},
		{
			name:     "Smaller prefix",
			resource: "8.8.8.0/24",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rpkiHistoryClient.Get(context.Background(), tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Get() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("Get() result is nil, want non-nil")
				return
			}

			if result.DataCallName != "rpki-history" {
				t.Errorf("Get() result.DataCallName = %s, want rpki-history", result.DataCallName)
			}

			if result.Data.Timeseries == nil {
				t.Error("Get() result.Data.Timeseries is nil, want non-nil")
				return
			}

			// If timeseries has data, validate structure
			if len(result.Data.Timeseries) > 0 {
				entry := result.Data.Timeseries[0]
				if entry.Prefix == "" {
					t.Error("Get() first entry has empty prefix")
				}
				if entry.Time.IsZero() {
					t.Error("Get() first entry has zero time")
				}
				if entry.Family == 0 {
					t.Error("Get() first entry has zero family")
				}
				if entry.MaxLength == 0 {
					t.Error("Get() first entry has zero max_length")
				}
			} else {
				t.Log("Get() result.Data.Timeseries is empty - this may be normal for some prefixes")
			}
		})
	}
}
