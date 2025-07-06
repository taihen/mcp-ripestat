package asroutingconsistency

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
			"version": "1.2",
			"data_call_name": "as-routing-consistency",
			"data_call_status": "supported - based on 2.1",
			"cached": true,
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
			"query_id": "20250616201149-d1dc0028-1b4d-4809-9d22-b8cba055b6a9",
			"process_time": 3,
			"server_id": "app195",
			"build_version": "main-2025.05.26",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-16T20:11:49.678721"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	asRoutingConsistencyClient := NewClient(c)

	ctx := context.Background()
	resp, err := asRoutingConsistencyClient.Get(ctx, "AS3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Data.Prefixes) != 1 {
		t.Errorf("expected 1 prefix, got %d", len(resp.Data.Prefixes))
	}
	if resp.Data.Prefixes[0].Prefix != "193.0.0.0/21" {
		t.Errorf("expected prefix 193.0.0.0/21, got %s", resp.Data.Prefixes[0].Prefix)
	}
	if len(resp.Data.Imports) != 1 {
		t.Errorf("expected 1 import, got %d", len(resp.Data.Imports))
	}
	if resp.Data.Imports[0].Peer != 1234 {
		t.Errorf("expected peer 1234, got %d", resp.Data.Imports[0].Peer)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	asRoutingConsistencyClient := NewClient(c)

	ctx := context.Background()
	_, err := asRoutingConsistencyClient.Get(ctx, "AS3333")
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
	asRoutingConsistencyClient := NewClient(c)

	ctx := context.Background()
	_, err := asRoutingConsistencyClient.Get(ctx, "AS3333")
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
	asRoutingConsistencyClient := NewClient(c)

	ctx := context.Background()
	_, err := asRoutingConsistencyClient.Get(ctx, "AS3333")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetASRoutingConsistency_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": { "prefixes": [], "imports": [] }
		}`))
	}))
	defer ts.Close()

	customClient := client.New(ts.URL, ts.Client())
	asRoutingConsistencyClient := NewClient(customClient)

	ctx := context.Background()
	resp, err := asRoutingConsistencyClient.Get(ctx, "AS3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Data.Prefixes) != 0 {
		t.Errorf("expected 0 prefixes, got %d", len(resp.Data.Prefixes))
	}
	if len(resp.Data.Imports) != 0 {
		t.Errorf("expected 0 imports, got %d", len(resp.Data.Imports))
	}
}

func TestGetASRoutingConsistency_ConvenienceFunction(t *testing.T) {
	ctx := context.Background()
	_, err := GetASRoutingConsistency(ctx, "")
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
