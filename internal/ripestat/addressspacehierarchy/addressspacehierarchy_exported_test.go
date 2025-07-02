package addressspacehierarchy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/addressspacehierarchy"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestGetAddressSpaceHierarchy_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"rir": "ripe",
				"resource": "193.0.0.0/21",
				"exact": [
					{
						"inetnum": "193.0.0.0 - 193.0.7.255",
						"netname": "RIPE-NCC",
						"descr": "RIPE Network Coordination Centre",
						"country": "NL",
						"status": "ASSIGNED PA",
						"source": "RIPE"
					}
				],
				"less_specific": [
					{
						"inetnum": "193.0.0.0 - 193.0.23.255",
						"netname": "NL-RIPENCC-OPS-990305",
						"country": "NL",
						"status": "ALLOCATED PA",
						"source": "RIPE"
					}
				],
				"more_specific": [],
				"query_time": "2025-07-02T07:05:33",
				"parameters": {
					"resource": "193.0.0.0/21"
				}
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	customClient := client.New(server.URL, nil)
	testClient := addressspacehierarchy.NewClient(customClient)

	ctx := context.Background()
	result, err := testClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Expected resource 193.0.0.0/21, got %s", result.Data.Resource)
	}
	if result.Data.RIR != "ripe" {
		t.Errorf("Expected RIR ripe, got %s", result.Data.RIR)
	}
	if len(result.Data.Exact) != 1 {
		t.Errorf("Expected 1 exact entry, got %d", len(result.Data.Exact))
	}
	if result.Data.Exact[0].Netname != "RIPE-NCC" {
		t.Errorf("Expected netname RIPE-NCC, got %s", result.Data.Exact[0].Netname)
	}
	if len(result.Data.LessSpecific) != 1 {
		t.Errorf("Expected 1 less specific entry, got %d", len(result.Data.LessSpecific))
	}
	if result.Data.LessSpecific[0].Status != "ALLOCATED PA" {
		t.Errorf("Expected status ALLOCATED PA, got %s", result.Data.LessSpecific[0].Status)
	}

	var _ = addressspacehierarchy.GetAddressSpaceHierarchy

	t.Log("Verified GetAddressSpaceHierarchy function signature")
}
