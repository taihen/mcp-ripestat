package e2e

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routinghistory"
)

func TestRoutingHistoryE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	c := client.DefaultClient()
	routingHistoryClient := routinghistory.New(c)

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "RIPE NCC ASN",
			resource: "AS3333",
			wantErr:  false,
		},
		{
			name:     "RIPE NCC prefix",
			resource: "193.0.0.0/21",
			wantErr:  false,
		},
		{
			name:     "Google DNS IP",
			resource: "8.8.8.8",
			wantErr:  false,
		},
		{
			name:     "Cloudflare ASN",
			resource: "AS13335",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := routingHistoryClient.Get(context.Background(), tt.resource)

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

			if result.Data.Resource == "" {
				t.Error("Get() result.Data.Resource is empty, want non-empty")
			}

			if len(result.Data.ByOrigin) == 0 {
				t.Log("Get() result.Data.ByOrigin is empty - this may be normal for some resources")
			}
		})
	}
}
