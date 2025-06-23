package whois_test

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whois"
)

func TestWhoisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := client.DefaultClient()
	whoisClient := whois.New(c)

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "valid IP address",
			resource: "8.8.8.8",
			wantErr:  false,
		},
		{
			name:     "valid ASN",
			resource: "AS3333",
			wantErr:  false,
		},
		{
			name:     "valid prefix",
			resource: "193.0.0.0/21",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := whoisClient.Get(context.Background(), tt.resource)

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
				t.Error("Get() result = nil, want non-nil")
				return
			}

			// Verify basic response structure
			if result.Status != "ok" {
				t.Errorf("Get() result.Status = %v, want ok", result.Status)
			}

			if result.Data.Resource != tt.resource {
				t.Errorf("Get() result.Data.Resource = %v, want %v", result.Data.Resource, tt.resource)
			}

			// Verify we have some records or authorities
			if len(result.Data.Records) == 0 && len(result.Data.Authorities) == 0 {
				t.Error("Get() result has no records or authorities")
			}

			t.Logf("Successfully retrieved whois data for %s", tt.resource)
		})
	}
}
