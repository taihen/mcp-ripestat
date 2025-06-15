package networkinfo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetNetworkInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
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

	client := ts.Client()
	ctx := context.Background()
	resp, err := getNetworkInfoWithClient(ctx, "140.78.90.50", client, ts.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Prefix != "140.78.0.0/16" {
		t.Errorf("expected prefix 140.78.0.0/16, got %s", resp.Data.Prefix)
	}
	if len(resp.Data.ASNs) != 1 || resp.Data.ASNs[0] != "1205" {
		t.Errorf("expected ASN 1205, got %v", resp.Data.ASNs)
	}
}

func TestGetNetworkInfo_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getNetworkInfoWithClient(ctx, "140.78.90.50", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestGetNetworkInfo_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"not_json":`))
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getNetworkInfoWithClient(ctx, "140.78.90.50", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestGetNetworkInfo_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"network-info","data_call_status":"supported","cached":false,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 50 * time.Millisecond}
	ctx := context.Background()
	_, err := getNetworkInfoWithClient(ctx, "140.78.90.50", client, ts.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetNetworkInfo_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messages":[],"see_also":[],"version":"1.1","data_call_name":"network-info","data_call_status":"supported","cached":false,"data":{"asns":["1205"],"prefix":"140.78.0.0/16"},"query_id":"test-id","process_time":1,"server_id":"test-server","build_version":"test-build","status":"ok","status_code":200,"time":"2025-06-15T16:31:58.741967"}`))
	}))
	defer ts.Close()

	oldClient := defaultHTTPClient
	oldBase := defaultBaseURL
	defaultHTTPClient = ts.Client()
	defaultBaseURL = ts.URL
	defer func() {
		defaultHTTPClient = oldClient
		defaultBaseURL = oldBase
	}()

	ctx := context.Background()
	resp, err := GetNetworkInfo(ctx, "140.78.90.50")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Prefix != "140.78.0.0/16" {
		t.Errorf("expected prefix 140.78.0.0/16, got %s", resp.Data.Prefix)
	}
}

func TestGetNetworkInfoWithClient_BadBaseURL(t *testing.T) {
	ctx := context.Background()
	_, err := getNetworkInfoWithClient(ctx, "140.78.90.50", http.DefaultClient, ":bad-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse RIPEstat base URL") {
		t.Errorf("expected URL parse error, got %v", err)
	}
}

func TestGetNetworkInfoWithClient_BadRequest(t *testing.T) {
	// Simulate a context that will cause NewRequestWithContext to fail
	badCtx, cancel := context.WithCancel(context.Background())
	cancel() // canceled context
	_, err := getNetworkInfoWithClient(badCtx, "140.78.90.50", http.DefaultClient, "http://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
