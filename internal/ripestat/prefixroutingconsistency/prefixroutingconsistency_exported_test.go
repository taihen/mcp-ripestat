package prefixroutingconsistency_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixroutingconsistency"
)

func TestGetPrefixRoutingConsistency_Integration(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"resource": "193.0.0.0/21",
				"routes": [
					{
						"in_bgp": true,
						"in_whois": true,
						"prefix": "193.0.0.0/21",
						"origin": 3333,
						"irr_sources": ["RIPE"],
						"asn_name": "RIPE-NCC-AS"
					}
				],
				"parameters": {
					"resource": "193.0.0.0/21",
					"data_overload_limit": ""
				},
				"query_starttime": "2025-06-30T16:00:00",
				"query_endtime": "2025-06-30T16:00:00"
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	// Create a custom client with our test server URL
	customClient := client.New(server.URL, nil)
	testClient := prefixroutingconsistency.NewClient(customClient)

	// Call the function directly on our test client instead of the global function
	ctx := context.Background()
	result, err := testClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Check the result
	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Expected resource 193.0.0.0/21, got %s", result.Data.Resource)
	}
	if len(result.Data.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(result.Data.Routes))
	}

	route := result.Data.Routes[0]
	if !route.InBGP {
		t.Error("Expected route in_bgp to be true")
	}
	if !route.InWHOIS {
		t.Error("Expected route in_whois to be true")
	}
	if route.Origin != 3333 {
		t.Errorf("Expected origin 3333, got %d", route.Origin)
	}

	// For completeness, let's verify the function exists and has the right signature
	var _ = prefixroutingconsistency.GetPrefixRoutingConsistency

	// This is just a compile-time check, not an actual test execution
	t.Log("Verified GetPrefixRoutingConsistency function signature")
}
