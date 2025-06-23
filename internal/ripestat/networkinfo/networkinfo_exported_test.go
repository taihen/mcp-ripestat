package networkinfo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
)

func TestGetNetworkInfo_Integration(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"asns": [12345],
				"prefix": "192.168.0.0/16",
				"resource": "192.168.0.1"
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	// Create a custom client with our test server URL
	customClient := client.New(server.URL, nil)
	testClient := networkinfo.NewClient(customClient)

	// Call the function directly on our test client instead of the global function
	ctx := context.Background()
	result, err := testClient.Get(ctx, "192.168.0.1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify ASNs are properly converted to strings
	if len(result.Data.ASNs) > 0 {
		_, ok := result.Data.ASNs[0].(string)
		if !ok {
			t.Errorf("Expected ASN to be converted to string, got %T", result.Data.ASNs[0])
		}
	}

	// Check the result
	if len(result.Data.ASNs) != 1 || result.Data.ASNs[0] != "12345" {
		t.Errorf("Expected ASNs [\"12345\"], got %v", result.Data.ASNs)
	}
	if result.Data.Prefix != "192.168.0.0/16" {
		t.Errorf("Expected prefix 192.168.0.0/16, got %s", result.Data.Prefix)
	}
	// Note: Data struct doesn't have a Resource field

	// Skip testing the exported function directly since we can't easily mock it
	// The exported function just calls DefaultClient().Get(), which we've already tested

	// For completeness, let's verify the function exists and has the right signature
	var _ func(context.Context, string) (*networkinfo.Response, error) = networkinfo.GetNetworkInfo

	// This is just a compile-time check, not an actual test execution
	t.Log("Verified GetNetworkInfo function signature")
}
