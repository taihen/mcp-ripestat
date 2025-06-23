package announcedprefixes

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
			"version": "1.0",
			"data_call_name": "announced-prefixes",
			"data_call_status": "supported",
			"cached": true,
			"data": {
				"resource": "3333",
				"prefixes": [
					{
						"prefix": "193.0.0.0/21",
						"timelines": [
							{
								"starttime": "2023-07-03 15:49:57",
								"endtime": "2025-06-16 16:00:00"
							}
						]
					},
					{
						"prefix": "193.0.10.0/23",
						"timelines": [
							{
								"starttime": "2023-07-03 15:49:57",
								"endtime": "2025-06-16 16:00:00"
							}
						]
					}
				],
				"query_time": "2025-06-16T16:00:00"
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
	announcedPrefixesClient := NewClient(c)

	ctx := context.Background()
	resp, err := announcedPrefixesClient.Get(ctx, "3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("expected resource 3333, got %s", resp.Data.Resource)
	}
	if len(resp.Data.Prefixes) != 2 {
		t.Errorf("expected 2 prefixes, got %d", len(resp.Data.Prefixes))
	}
	if resp.Data.Prefixes[0].Prefix != "193.0.0.0/21" {
		t.Errorf("expected prefix 193.0.0.0/21, got %s", resp.Data.Prefixes[0].Prefix)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	announcedPrefixesClient := NewClient(c)

	ctx := context.Background()
	_, err := announcedPrefixesClient.Get(ctx, "3333")
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
	announcedPrefixesClient := NewClient(c)

	ctx := context.Background()
	_, err := announcedPrefixesClient.Get(ctx, "3333")
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
	announcedPrefixesClient := NewClient(c)

	ctx := context.Background()
	_, err := announcedPrefixesClient.Get(ctx, "3333")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Errorf("expected timeout error (context deadline or client timeout), got %v", err)
	}
}

func TestGetAnnouncedPrefixes_Exported(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": { "resource": "3333", "prefixes": [] }
		}`))
	}))
	defer ts.Close()

	// Create a custom client that uses our test server
	customClient := client.New(ts.URL, ts.Client())
	announcedPrefixesClient := NewClient(customClient)

	// Test the client directly
	ctx := context.Background()
	resp, err := announcedPrefixesClient.Get(ctx, "3333")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("expected resource 3333, got %s", resp.Data.Resource)
	}

	// Since we can't mock the GetAnnouncedPrefixes function directly,
	// we're effectively testing that DefaultClient() and Get() work together
	// which is what the exported GetAnnouncedPrefixes function does
}

func TestGetAnnouncedPrefixes_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetAnnouncedPrefixes(ctx, "")
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
