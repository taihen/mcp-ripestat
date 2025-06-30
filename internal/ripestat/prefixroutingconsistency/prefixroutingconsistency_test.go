package prefixroutingconsistency

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "0.7",
			"data_call_name": "prefix-routing-consistency",
			"data_call_status": "supported - based on version 1.1",
			"cached": false,
			"data": {
				"resource": "193.0.0.0/21",
				"routes": [
					{
						"in_bgp": true,
						"in_whois": true,
						"prefix": "193.0.0.0/21",
						"origin": 3333,
						"irr_sources": ["RIPE"],
						"asn_name": "RIPE-NCC-AS - Reseaux IP Europeens Network Coordination Centre (RIPE NCC)"
					}
				],
				"parameters": {
					"resource": "193.0.0.0/21",
					"data_overload_limit": ""
				},
				"query_starttime": "2025-06-30T16:00:00",
				"query_endtime": "2025-06-30T16:00:00"
			},
			"query_id": "20250630214254-e0fb9b3a-0845-4e30-bec9-c790f0a5d6df",
			"process_time": 118,
			"server_id": "app176",
			"build_version": "main-2025.06.24",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-30T21:42:54.651331"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	resp, err := client.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
	if len(resp.Data.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(resp.Data.Routes))
	}
	route := resp.Data.Routes[0]
	if !route.InBGP {
		t.Error("expected route in_bgp to be true")
	}
	if !route.InWHOIS {
		t.Error("expected route in_whois to be true")
	}
	if route.Origin != 3333 {
		t.Errorf("expected origin 3333, got %d", route.Origin)
	}
	if len(route.IRRSources) != 1 || route.IRRSources[0] != "RIPE" {
		t.Errorf("expected IRR sources [RIPE], got %v", route.IRRSources)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	_, err := client.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP status: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestClient_Get_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not_json":`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	client := NewClient(c)

	ctx := context.Background()
	_, err := client.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestClient_Get_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	c := client.New(ts.URL, httpClient)
	client := NewClient(c)

	ctx := context.Background()
	_, err := client.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetPrefixRoutingConsistency_Exported(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
			}
		}`))
	}))
	defer ts.Close()

	// Create a custom client that uses our test server
	c := client.New(ts.URL, ts.Client())
	customClient := NewClient(c)

	// Test the client directly
	ctx := context.Background()
	resp, err := customClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
}

func TestGetPrefixRoutingConsistency_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetPrefixRoutingConsistency(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := DefaultClient()
	ctx := context.Background()
	_, err := c.Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}

func TestNewClient_NilClient(t *testing.T) {
	client := NewClient(nil)
	if client == nil {
		t.Fatal("expected client to be created, got nil")
	}
	if client.client == nil {
		t.Fatal("expected client.client to be set")
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()
	if client == nil {
		t.Fatal("expected client to be created, got nil")
	}
	if client.client == nil {
		t.Fatal("expected client.client to be set")
	}
}
