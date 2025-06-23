package routingstatus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
)

func TestGetRoutingStatus_Integration(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"announced": true,
				"resource": "192.168.0.0/24"
			},
			"status": "ok",
			"status_code": 200
		}`))
	}))
	defer server.Close()

	// Create a custom client with our test server URL
	customClient := client.New(server.URL, nil)
	testClient := routingstatus.NewClient(customClient)

	// Call the function directly on our test client instead of the global function
	ctx := context.Background()
	result, err := testClient.Get(ctx, "192.168.0.0/24")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Check the result
	if !result.Data.Announced {
		t.Errorf("Expected announced to be true")
	}
	if result.Data.Resource != "192.168.0.0/24" {
		t.Errorf("Expected resource 192.168.0.0/24, got %s", result.Data.Resource)
	}

	// Skip testing the exported function directly since we can't easily mock it
	// The exported function just calls DefaultClient().Get(), which we've already tested

	// For completeness, let's verify the function exists and has the right signature
	var _ func(context.Context, string) (*routingstatus.Response, error) = routingstatus.GetRoutingStatus

	// This is just a compile-time check, not an actual test execution
	t.Log("Verified GetRoutingStatus function signature")
}
