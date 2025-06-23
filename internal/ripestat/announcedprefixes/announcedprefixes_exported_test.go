package announcedprefixes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestGetAnnouncedPrefixes_Integration(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"prefixes": [
					{
						"prefix": "192.168.0.0/16",
						"asn": 12345
					}
				]
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	// Create a custom client with our test server URL
	customClient := client.New(server.URL, nil)
	testClient := announcedprefixes.NewClient(customClient)

	// Call the function directly on our test client instead of the global function
	ctx := context.Background()
	result, err := testClient.Get(ctx, "AS12345")
	if err != nil {
		t.Fatalf("GetAnnouncedPrefixes() error = %v", err)
	}
	// Check the result
	if len(result.Data.Prefixes) != 1 {
		t.Errorf("Expected 1 prefix, got %d", len(result.Data.Prefixes))
	}
	if result.Data.Prefixes[0].Prefix != "192.168.0.0/16" {
		t.Errorf("Expected prefix 192.168.0.0/16, got %s", result.Data.Prefixes[0].Prefix)
	}
	// The Prefix struct doesn't have an ASN field, so we just check the prefix
}
