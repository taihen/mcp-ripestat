package aspathlength

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestClient_Get_Success(t *testing.T) {
	mockResponse := `{
		"data": {
			"stats": [
				{
					"number": 0,
					"count": 308,
					"location": "Amsterdam, Netherlands",
					"stripped": {
						"sum": 1109,
						"min": 1,
						"max": 6,
						"avg": 3.6
					},
					"unstripped": {
						"sum": 1134,
						"min": 1,
						"max": 6,
						"avg": 3.68
					}
				}
			],
			"resource": "3333",
			"query_time": "2025-07-04T00:00:00",
			"sort_by": "number"
		},
		"status": "ok",
		"status_code": 200
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/as-path-length/data.json" {
			t.Errorf("Expected path '/data/as-path-length/data.json', got %s", r.URL.Path)
		}
		if r.URL.Query().Get("resource") != "AS3333" {
			t.Errorf("Expected resource 'AS3333', got %s", r.URL.Query().Get("resource"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx := context.Background()
	resp, err := c.Get(ctx, "AS3333")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("Expected resource '3333', got %s", resp.Data.Resource)
	}
	if len(resp.Data.Stats) != 1 {
		t.Errorf("Expected 1 stat entry, got %d", len(resp.Data.Stats))
	}
	if resp.Data.Stats[0].Number != 0 {
		t.Errorf("Expected number 0, got %d", resp.Data.Stats[0].Number)
	}
	if resp.Data.Stats[0].Count != 308 {
		t.Errorf("Expected count 308, got %d", resp.Data.Stats[0].Count)
	}
}

func TestClient_Get_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx := context.Background()
	_, err := c.Get(ctx, "AS3333")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get AS path length data") {
		t.Errorf("Expected error message to contain 'failed to get AS path length data', got %v", err)
	}
}

func TestClient_Get_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	httpClient := client.New(server.URL, server.Client())
	c := NewClient(httpClient)

	ctx := context.Background()
	_, err := c.Get(ctx, "AS3333")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_Get_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	c := NewClient(client.New(server.URL, httpClient))

	ctx := context.Background()
	_, err := c.Get(ctx, "AS3333")

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	c := NewClient(nil)

	ctx := context.Background()
	_, err := c.Get(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected error message to contain 'resource parameter is required', got %v", err)
	}
}

func TestGetASPathLength_Exported(t *testing.T) {
	// Test with empty resource to verify error handling
	ctx := context.Background()
	_, err := GetASPathLength(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected error message to contain 'resource parameter is required', got %v", err)
	}
}

func TestGetASPathLength_ConvenienceFunction(t *testing.T) {
	ctx := context.Background()
	_, err := GetASPathLength(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected error message to contain 'resource parameter is required', got %v", err)
	}
}

func TestNewClient(t *testing.T) {
	httpClient := client.DefaultClient()
	c := NewClient(httpClient)

	if c == nil {
		t.Fatal("Expected client, got nil")
	}
	if c.client != httpClient {
		t.Error("Expected client to be set")
	}
}

func TestNewClient_NilClient(t *testing.T) {
	c := NewClient(nil)

	if c == nil {
		t.Fatal("Expected client, got nil")
	}
	if c.client == nil {
		t.Error("Expected default client to be set")
	}
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()

	if c == nil {
		t.Fatal("Expected client, got nil")
	}
	if c.client == nil {
		t.Error("Expected client to be set")
	}
}
