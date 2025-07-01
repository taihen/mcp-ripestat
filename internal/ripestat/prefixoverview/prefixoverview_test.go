package prefixoverview

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
			"version": "1.3",
			"data_call_name": "prefix-overview",
			"data_call_status": "supported",
			"cached": false,
			"data": {
				"is_less_specific": false,
				"announced": true,
				"asns": [
					{
						"asn": 3333,
						"holder": "RIPE-NCC-AS - Reseaux IP Europeens Network Coordination Centre (RIPE NCC)"
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
			"query_id": "20250701053214-b1e480b2-047a-404c-b6d0-5d7a20669b85",
			"process_time": 89,
			"server_id": "app175",
			"build_version": "main-2025.06.24",
			"status": "ok",
			"status_code": 200,
			"time": "2025-07-01T05:32:14.140666"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	prefixOverviewClient := NewClient(c)

	ctx := context.Background()
	resp, err := prefixOverviewClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
	if !resp.Data.Announced {
		t.Error("expected announced to be true")
	}
	if len(resp.Data.ASNs) != 1 {
		t.Errorf("expected 1 ASN, got %d", len(resp.Data.ASNs))
	}
	if resp.Data.ASNs[0].ASN != 3333 {
		t.Errorf("expected ASN 3333, got %d", resp.Data.ASNs[0].ASN)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	prefixOverviewClient := NewClient(c)

	ctx := context.Background()
	_, err := prefixOverviewClient.Get(ctx, "193.0.0.0/21")
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
	prefixOverviewClient := NewClient(c)

	ctx := context.Background()
	_, err := prefixOverviewClient.Get(ctx, "193.0.0.0/21")
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
	prefixOverviewClient := NewClient(c)

	ctx := context.Background()
	_, err := prefixOverviewClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetPrefixOverview_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": { "resource": "193.0.0.0/21", "announced": true }
		}`))
	}))
	defer ts.Close()

	customClient := client.New(ts.URL, ts.Client())
	prefixOverviewClient := NewClient(customClient)

	ctx := context.Background()
	resp, err := prefixOverviewClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
}

func TestGetPrefixOverview_ConvenienceFunction(t *testing.T) {
	ctx := context.Background()
	_, err := GetPrefixOverview(ctx, "")
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
