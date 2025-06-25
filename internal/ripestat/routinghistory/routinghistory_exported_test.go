package routinghistory_test

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routinghistory"
)

func TestRoutingHistoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := client.DefaultClient()
	routingHistoryClient := routinghistory.New(c)

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid ASN",
			resource: "AS3333",
			wantErr:  false,
		},
		{
			name:     "valid IP prefix",
			resource: "193.0.0.0/21",
			wantErr:  false,
		},
		{
			name:     "valid IP address",
			resource: "8.8.8.8",
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
		})
	}
}

func TestGetRoutingHistoryFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid ASN",
			resource: "AS3333",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := routinghistory.GetRoutingHistory(context.Background(), tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetRoutingHistory() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRoutingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("GetRoutingHistory() result is nil, want non-nil")
			}
		})
	}
}
