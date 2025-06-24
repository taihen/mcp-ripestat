package lookingglass_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/lookingglass"
)

func TestLookingGlassIntegration(t *testing.T) {
	// Create a test server that mimics the RIPEstat API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path
		expectedPath := "/data/looking-glass/data.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify query parameters
		resource := r.URL.Query().Get("resource")
		if resource == "" {
			t.Errorf("Expected resource parameter, got empty")
		}

		lookBackLimit := r.URL.Query().Get("look_back_limit")
		if lookBackLimit == "" {
			t.Errorf("Expected look_back_limit parameter, got empty")
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"messages": [],
			"see_also": [],
			"version": "2.1",
			"data_call_name": "looking-glass",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"rrcs": [
					{
						"rrc": "RRC00",
						"location": "Amsterdam, Netherlands",
						"peers": [
							{
								"asn_origin": "1205",
								"as_path": "34854 6939 1853 1853 1205",
								"community": "34854:1000",
								"largeCommunity": "",
								"extendedCommunity": "",
								"last_updated": "2025-06-18T10:54:36",
								"prefix": "140.78.0.0/16",
								"peer": "2.56.11.1",
								"origin": "IGP",
								"next_hop": "2.56.11.1",
								"latest_time": "2025-06-24T03:07:19"
							}
						]
					},
					{
						"rrc": "RRC01",
						"location": "London, United Kingdom",
						"peers": [
							{
								"asn_origin": "1205",
								"as_path": "34854 6939 1853 1853 1205",
								"community": "34854:1000",
								"largeCommunity": "",
								"extendedCommunity": "",
								"last_updated": "2025-06-18T10:54:36",
								"prefix": "140.78.0.0/16",
								"peer": "2.56.11.2",
								"origin": "IGP",
								"next_hop": "2.56.11.2",
								"latest_time": "2025-06-24T03:07:19"
							}
						]
					}
				]
			},
			"query_time": "2025-06-24T03:07:22",
			"latest_time": "2025-06-24T03:06:23",
			"parameters": {
				"resource": "140.78.0.0/16",
				"look_back_limit": 3600,
				"cache": null
			},
			"query_id": "20250624030722-9dd4958b-8287-445b-8936-f055e008bef0",
			"process_time": 54,
			"server_id": "app188",
			"build_version": "main-2025.06.23",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-24T03:07:22.385069"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create a client that points to our test server
	httpClient := &http.Client{}
	c := lookingglass.NewClient(client.New(server.URL, httpClient))

	// Test the Get method
	ctx := context.Background()
	result, err := c.Get(ctx, "140.78.0.0/16", 3600)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify the response structure
	if len(result.RRCs) != 2 {
		t.Errorf("Expected 2 RRCs, got %d", len(result.RRCs))
	}

	if result.FetchedAt == "" {
		t.Error("Expected FetchedAt to be set")
	}

	// Verify first RRC
	if result.RRCs[0].RRC != "RRC00" {
		t.Errorf("Expected first RRC to be 'RRC00', got '%s'", result.RRCs[0].RRC)
	}

	if result.RRCs[0].Location != "Amsterdam, Netherlands" {
		t.Errorf("Expected first RRC location to be 'Amsterdam, Netherlands', got '%s'", result.RRCs[0].Location)
	}

	if len(result.RRCs[0].Peers) != 1 {
		t.Errorf("Expected 1 peer in first RRC, got %d", len(result.RRCs[0].Peers))
	}

	// Verify peer information
	peer := result.RRCs[0].Peers[0]
	if peer.ASNOrigin != "1205" {
		t.Errorf("Expected ASN origin '1205', got '%s'", peer.ASNOrigin)
	}

	if peer.Prefix != "140.78.0.0/16" {
		t.Errorf("Expected prefix '140.78.0.0/16', got '%s'", peer.Prefix)
	}

	if peer.Peer != "2.56.11.1" {
		t.Errorf("Expected peer '2.56.11.1', got '%s'", peer.Peer)
	}

	// Verify second RRC
	if result.RRCs[1].RRC != "RRC01" {
		t.Errorf("Expected second RRC to be 'RRC01', got '%s'", result.RRCs[1].RRC)
	}

	if result.RRCs[1].Location != "London, United Kingdom" {
		t.Errorf("Expected second RRC location to be 'London, United Kingdom', got '%s'", result.RRCs[1].Location)
	}
}

func TestLookingGlassErrorHandling(t *testing.T) {
	// Test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	// Create a client that points to our test server
	httpClient := &http.Client{}
	c := lookingglass.NewClient(client.New(server.URL, httpClient))

	// Test the Get method with server error
	ctx := context.Background()
	result, err := c.Get(ctx, "140.78.0.0/16", 3600)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestLookingGlassParameterValidation(t *testing.T) {
	// Create a basic test server (won't be called due to parameter validation)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"rrcs": []}}`))
	}))
	defer server.Close()

	httpClient := &http.Client{}
	c := lookingglass.NewClient(client.New(server.URL, httpClient))
	ctx := context.Background()

	// Test empty resource
	_, err := c.Get(ctx, "", 3600)
	if err == nil {
		t.Error("Expected error for empty resource")
	}

	// Test negative look back limit
	_, err = c.Get(ctx, "140.78.0.0/16", -1)
	if err == nil {
		t.Error("Expected error for negative look back limit")
	}

	// Test look back limit too large
	_, err = c.Get(ctx, "140.78.0.0/16", lookingglass.MaxLookBackLimit+1)
	if err == nil {
		t.Error("Expected error for look back limit too large")
	}
}

func TestLookingGlassDefaultClient(t *testing.T) {
	// Test that DefaultClient returns a non-nil client
	c := lookingglass.DefaultClient()
	if c == nil {
		t.Error("DefaultClient() returned nil")
	}
}

func TestLookingGlassNewClientWithNil(t *testing.T) {
	// Test that NewClient with nil parameter uses default client
	c := lookingglass.NewClient(nil)
	if c == nil {
		t.Error("NewClient(nil) returned nil")
	}
}
