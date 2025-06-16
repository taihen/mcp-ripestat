package asoverview

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"messages": [],
			"see_also": [],
			"version": "1.3",
			"data_call_name": "as-overview",
			"data_call_status": "supported - based on 2.1",
			"cached": true,
			"data": {
				"type": "as",
				"resource": "3333",
				"block": {
					"resource": "3154-3353",
					"desc": "Assigned by RIPE NCC",
					"name": "IANA 16-bit Autonomous System (AS) Numbers Registry"
				},
				"holder": "RIPE-NCC-AS - Reseaux IP Europeens Network Coordination Centre (RIPE NCC)",
				"announced": true,
				"query_starttime": "2025-06-16T16:00:00",
				"query_endtime": "2025-06-16T16:00:00"
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

	client := ts.Client()
	ctx := context.Background()
	resp, err := getWithClient(ctx, "3333", client, ts.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("expected resource 3333, got %s", resp.Data.Resource)
	}
	if !resp.Data.Announced {
		t.Error("expected announced to be true")
	}
}

func TestGet_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getWithClient(ctx, "3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 502") {
		t.Errorf("expected status code error, got %v", err)
	}
}

func TestGet_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not_json":`))
	}))
	defer ts.Close()

	client := ts.Client()
	ctx := context.Background()
	_, err := getWithClient(ctx, "3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestGet_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := &http.Client{Timeout: 50 * time.Millisecond}
	ctx := context.Background()
	_, err := getWithClient(ctx, "3333", client, ts.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGet_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": { "resource": "3333", "announced": true }
		}`))
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
	resp, err := Get(ctx, "3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("expected resource 3333, got %s", resp.Data.Resource)
	}
}

func TestGetWithClient_BadBaseURL(t *testing.T) {
	ctx := context.Background()
	_, err := getWithClient(ctx, "3333", http.DefaultClient, ":bad-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse RIPEstat base URL") {
		t.Errorf("expected URL parse error, got %v", err)
	}
}

func TestGetWithClient_BadRequest(t *testing.T) {
	badCtx, cancel := context.WithCancel(context.Background())
	cancel() // canceled context
	_, err := getWithClient(badCtx, "3333", http.DefaultClient, "http://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
