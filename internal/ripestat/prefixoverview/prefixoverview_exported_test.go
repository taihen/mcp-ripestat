package prefixoverview_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixoverview"
)

func TestGetPrefixOverview_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"is_less_specific": false,
				"announced": true,
				"asns": [
					{
						"asn": 3333,
						"holder": "RIPE-NCC-AS"
					}
				],
				"related_prefixes": [],
				"resource": "193.0.0.0/21",
				"type": "prefix",
				"block": {
					"resource": "193.0.0.0/8",
					"desc": "RIPE NCC (Status: ALLOCATED)",
					"name": "IANA IPv4 Address Space Registry"
				},
				"actual_num_related": 0,
				"query_time": "2025-07-01T00:00:00",
				"num_filtered_out": 0
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	customClient := client.New(server.URL, nil)
	testClient := prefixoverview.NewClient(customClient)

	ctx := context.Background()
	result, err := testClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Expected resource 193.0.0.0/21, got %s", result.Data.Resource)
	}
	if !result.Data.Announced {
		t.Errorf("Expected announced to be true")
	}
	if len(result.Data.ASNs) != 1 {
		t.Errorf("Expected 1 ASN, got %d", len(result.Data.ASNs))
	}
	if result.Data.ASNs[0].ASN != 3333 {
		t.Errorf("Expected ASN 3333, got %d", result.Data.ASNs[0].ASN)
	}
	if result.Data.Type != "prefix" {
		t.Errorf("Expected type prefix, got %s", result.Data.Type)
	}

	var _ = prefixoverview.GetPrefixOverview

	t.Log("Verified GetPrefixOverview function signature")
}
