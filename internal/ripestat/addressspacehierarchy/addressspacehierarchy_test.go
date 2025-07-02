package addressspacehierarchy

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
			"data_call_name": "address-space-hierarchy",
			"data_call_status": "supported",
			"cached": true,
			"data": {
				"rir": "ripe",
				"resource": "193.0.0.0/21",
				"exact": [
					{
						"inetnum": "193.0.0.0 - 193.0.7.255",
						"netname": "RIPE-NCC",
						"descr": "RIPE Network Coordination Centre, Amsterdam, Netherlands",
						"org": "ORG-RIEN1-RIPE",
						"country": "NL",
						"status": "ASSIGNED PA",
						"created": "2003-03-17T12:15:57Z",
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
					"resource": "193.0.0.0/21",
					"cache": null
				}
			},
			"query_id": "20250702071637-a6674868-a79a-42fd-8cd7-a09d1d38fcd5",
			"process_time": 1,
			"server_id": "app169",
			"build_version": "main-2025.06.24",
			"status": "ok",
			"status_code": 200,
			"time": "2025-07-02T07:16:37.290049"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	addressSpaceClient := NewClient(c)

	ctx := context.Background()
	resp, err := addressSpaceClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "193.0.0.0/21" {
		t.Errorf("expected resource 193.0.0.0/21, got %s", resp.Data.Resource)
	}
	if resp.Data.RIR != "ripe" {
		t.Errorf("expected RIR ripe, got %s", resp.Data.RIR)
	}
	if len(resp.Data.Exact) != 1 {
		t.Errorf("expected 1 exact entry, got %d", len(resp.Data.Exact))
	}
	if resp.Data.Exact[0].Netname != "RIPE-NCC" {
		t.Errorf("expected netname RIPE-NCC, got %s", resp.Data.Exact[0].Netname)
	}
	if len(resp.Data.LessSpecific) != 1 {
		t.Errorf("expected 1 less specific entry, got %d", len(resp.Data.LessSpecific))
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	addressSpaceClient := NewClient(c)

	ctx := context.Background()
	_, err := addressSpaceClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP status: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestClient_Get_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	addressSpaceClient := NewClient(c)

	ctx := context.Background()
	_, err := addressSpaceClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := client.New("http://example.com", http.DefaultClient)
	addressSpaceClient := NewClient(c)

	ctx := context.Background()
	_, err := addressSpaceClient.Get(ctx, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource validation error, got %v", err)
	}
}

func TestClient_Get_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	addressSpaceClient := NewClient(c)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := addressSpaceClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestNewClient_WithNilClient(t *testing.T) {
	addressSpaceClient := NewClient(nil)
	if addressSpaceClient == nil {
		t.Fatal("expected client, got nil")
	}
	if addressSpaceClient.client == nil {
		t.Fatal("expected internal client, got nil")
	}
}

func TestDefaultClient(t *testing.T) {
	addressSpaceClient := DefaultClient()
	if addressSpaceClient == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestGetAddressSpaceHierarchy_EmptyResource(t *testing.T) {
	ctx := context.Background()
	_, err := GetAddressSpaceHierarchy(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource validation error, got %v", err)
	}
}
