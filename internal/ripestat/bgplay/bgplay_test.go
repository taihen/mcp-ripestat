package bgplay

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
			"data_call_name": "bgplay",
			"data_call_status": "supported",
			"cached": true,
			"data": {
				"resource": "8.8.8.8",
				"query_starttime": "2025-06-16T16:00:00",
				"query_endtime": "2025-06-16T17:00:00",
				"target_prefix": "8.8.8.0/24",
				"rrcs": [0, 1, 3],
				"initial_state": [
					{
						"target_prefix": "8.8.8.0/24",
						"source_id": "rrc00-1",
						"path": [15169],
						"community": []
					}
				],
				"events": [
					{
						"type": "A",
						"timestamp": "2025-06-16T16:30:00",
						"attrs": {
							"target_prefix": "8.8.8.0/24",
							"source_id": "rrc00-2",
							"path": [1299, 15169],
							"community": []
						}
					}
				]
			},
			"query_id": "20250616201149-d1dc0028-1b4d-4809-9d22-b8cba055b6a9",
			"process_time": 5,
			"server_id": "app195",
			"build_version": "main-2025.05.26",
			"status": "ok",
			"status_code": 200,
			"time": "2025-06-16T20:11:49.678721"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	bgplayClient := New(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := bgplayClient.Get(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Data.Resource != "8.8.8.8" {
		t.Errorf("Expected resource '8.8.8.8', got '%s'", result.Data.Resource)
	}

	if result.Data.TargetPrefix != "8.8.8.0/24" {
		t.Errorf("Expected target prefix '8.8.8.0/24', got '%s'", result.Data.TargetPrefix)
	}

	if len(result.Data.InitialState) != 1 {
		t.Errorf("Expected 1 initial state record, got %d", len(result.Data.InitialState))
	}

	if len(result.Data.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(result.Data.Events))
	}

	if result.Data.Events[0].Type != "A" {
		t.Errorf("Expected event type 'A', got '%s'", result.Data.Events[0].Type)
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := client.New("http://example.com", http.DefaultClient)
	bgplayClient := New(c)

	ctx := context.Background()
	_, err := bgplayClient.Get(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}

	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected error message about required resource parameter, got %v", err)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	bgplayClient := New(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := bgplayClient.Get(ctx, "8.8.8.8")
	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}
}

func TestClient_Get_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	bgplayClient := New(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := bgplayClient.Get(ctx, "8.8.8.8")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestClient_Get_ContextTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {}}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	bgplayClient := New(c)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := bgplayClient.Get(ctx, "8.8.8.8")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestDefaultClient(t *testing.T) {
	defaultClient := DefaultClient()
	if defaultClient == nil {
		t.Fatal("Expected default client, got nil")
	}

	if defaultClient.client == nil {
		t.Fatal("Expected default client to have underlying client, got nil")
	}
}

func TestGetBGPlay_ConvenienceFunction(t *testing.T) {
	// Test the convenience function with an empty resource to trigger error path
	ctx := context.Background()
	_, err := GetBGPlay(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("expected resource required error, got %v", err)
	}
}
