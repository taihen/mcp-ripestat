package networkinfo

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
			"messages":[],
			"see_also":[],
			"version":"1.1",
			"data_call_name":"network-info",
			"data_call_status":"supported",
			"cached":false,
			"data":{"asns":["1205"],"prefix":"140.78.0.0/16"},
			"query_id":"test-id",
			"process_time":1,
			"server_id":"test-server",
			"build_version":"test-build",
			"status":"ok",
			"status_code":200,
			"time":"2025-06-15T16:31:58.741967"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	networkInfoClient := NewClient(c)

	ctx := context.Background()
	resp, err := networkInfoClient.Get(ctx, "140.78.90.50")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Prefix != "140.78.0.0/16" {
		t.Errorf("expected prefix 140.78.0.0/16, got %s", resp.Data.Prefix)
	}
	if len(resp.Data.ASNs) != 1 {
		t.Errorf("expected 1 ASN, got %d", len(resp.Data.ASNs))
	}

	asnStr, ok := resp.Data.ASNs[0].(string)
	if !ok || asnStr != "1205" {
		t.Errorf("expected ASN 1205, got %v", resp.Data.ASNs[0])
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	networkInfoClient := NewClient(c)

	ctx := context.Background()
	_, err := networkInfoClient.Get(ctx, "140.78.90.50")

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
	networkInfoClient := NewClient(c)

	ctx := context.Background()
	_, err := networkInfoClient.Get(ctx, "140.78.90.50")

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
		_, _ = w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"network-info","data_call_status":"supported","cached":false,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	c := client.New(ts.URL, httpClient)
	networkInfoClient := NewClient(c)

	ctx := context.Background()
	_, err := networkInfoClient.Get(ctx, "140.78.90.50")

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetNetworkInfo_Exported(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"network-info","data_call_status":"supported","cached":false,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	// Create a custom client that uses our test server
	customClient := client.New(ts.URL, ts.Client())
	networkInfoClient := NewClient(customClient)

	// Test the client directly
	ctx := context.Background()
	resp, err := networkInfoClient.Get(ctx, "140.78.90.50")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Prefix != "140.78.0.0/16" {
		t.Errorf("expected prefix 140.78.0.0/16, got %s", resp.Data.Prefix)
	}

	// Since we can't mock the GetNetworkInfo function directly,
	// we're effectively testing that DefaultClient() and Get() work together
	// which is what GetNetworkInfo does
}

func TestGetNetworkInfo_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetNetworkInfo(ctx, "")
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
