package e2e

import (
	"context"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whois"
)

func TestWhoisE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	c := client.DefaultClient()
	whoisClient := whois.New(c)

	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "Google DNS IP",
			resource: "8.8.8.8",
			wantErr:  false,
		},
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
			name:     "Cloudflare IP",
			resource: "1.1.1.1",
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

			// For ASNs, the API may return the number without the "AS" prefix
			expectedResource := tt.resource
			if tt.resource == "AS3333" && result.Data.Resource == "3333" {
				expectedResource = "3333"
			}
			if result.Data.Resource != expectedResource {
				t.Errorf("Get() result.Data.Resource = %v, want %v", result.Data.Resource, expectedResource)
			}

			// Verify we have some data
			if len(result.Data.Records) == 0 && len(result.Data.IRRRecords) == 0 && len(result.Data.Authorities) == 0 {
				t.Error("Get() result has no records, IRR records, or authorities")
			}

			// Log some details for debugging
			t.Logf("Resource: %s", result.Data.Resource)
			t.Logf("Query time: %s", result.Data.QueryTime)
			t.Logf("Number of record groups: %d", len(result.Data.Records))
			t.Logf("Number of IRR records: %d", len(result.Data.IRRRecords))
			t.Logf("Number of authorities: %d", len(result.Data.Authorities))

			// Verify record structure if we have records
			if len(result.Data.Records) > 0 {
				for i, recordGroup := range result.Data.Records {
					if len(recordGroup) == 0 {
						t.Errorf("Record group %d is empty", i)
						continue
					}

					for j, record := range recordGroup {
						if record.Key == "" {
							t.Errorf("Record group %d, record %d has empty key", i, j)
						}
						if record.Value == "" {
							t.Errorf("Record group %d, record %d has empty value", i, j)
						}
					}
				}
			}

			t.Logf("Successfully retrieved whois data for %s", tt.resource)
		})
	}
}
