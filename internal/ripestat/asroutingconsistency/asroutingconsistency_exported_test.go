package asroutingconsistency_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/asroutingconsistency"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestGetASRoutingConsistency_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"prefixes": [
					{
						"in_bgp": true,
						"in_whois": true,
						"irr_sources": ["RIPE"],
						"prefix": "193.0.0.0/21"
					}
				],
				"imports": [
					{
						"in_bgp": true,
						"in_whois": false,
						"peer": 1234
					}
				]
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	customClient := client.New(server.URL, nil)
	testClient := asroutingconsistency.NewClient(customClient)

	ctx := context.Background()
	result, err := testClient.Get(ctx, "AS3333")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(result.Data.Prefixes) != 1 {
		t.Errorf("Expected 1 prefix, got %d", len(result.Data.Prefixes))
	}
	if result.Data.Prefixes[0].Prefix != "193.0.0.0/21" {
		t.Errorf("Expected prefix 193.0.0.0/21, got %s", result.Data.Prefixes[0].Prefix)
	}
	if len(result.Data.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(result.Data.Imports))
	}
	if result.Data.Imports[0].Peer != 1234 {
		t.Errorf("Expected peer 1234, got %d", result.Data.Imports[0].Peer)
	}

	var _ = asroutingconsistency.GetASRoutingConsistency

	t.Log("Verified GetASRoutingConsistency function signature")
}
