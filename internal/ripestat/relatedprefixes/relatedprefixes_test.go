package relatedprefixes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

func TestNewClient(t *testing.T) {
	c := client.DefaultClient()
	relatedClient := NewClient(c)
	if relatedClient == nil {
		t.Fatal("NewClient returned nil")
	}
	if relatedClient.client != c {
		t.Error("NewClient did not use provided client")
	}
}

func TestNewClientWithNil(t *testing.T) {
	relatedClient := NewClient(nil)
	if relatedClient == nil {
		t.Fatal("NewClient returned nil")
	}
	if relatedClient.client == nil {
		t.Error("NewClient did not create default client")
	}
}

func TestDefaultClient(t *testing.T) {
	relatedClient := DefaultClient()
	if relatedClient == nil {
		t.Fatal("DefaultClient returned nil")
	}
	if relatedClient.client == nil {
		t.Error("DefaultClient did not create client")
	}
}

func TestClient_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"resource": "193.0.0.0/21",
				"prefixes": [
					{
						"prefix": "193.0.8.0/23",
						"origin_asn": "197000",
						"asn_name": "RIPE-NCC-AUTHDNS-AS",
						"relationship": "Adjacency - Right"
					}
				],
				"query_time": "2025-07-03T06:44:00"
			},
			"status": "ok"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	relatedClient := NewClient(c)

	ctx := context.Background()
	result, err := relatedClient.Get(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Get() returned nil result")
	}

	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Resource = %v, want %v", result.Data.Resource, "193.0.0.0/21")
	}

	if len(result.Data.Prefixes) != 1 {
		t.Errorf("Number of prefixes = %v, want %v", len(result.Data.Prefixes), 1)
	}

	if result.Data.Prefixes[0].Prefix != "193.0.8.0/23" {
		t.Errorf("First prefix = %v, want %v", result.Data.Prefixes[0].Prefix, "193.0.8.0/23")
	}
}

func TestClient_Get_EmptyResource(t *testing.T) {
	relatedClient := DefaultClient()
	ctx := context.Background()
	_, err := relatedClient.Get(ctx, "")
	if err == nil {
		t.Fatal("Get() expected error for empty resource, got nil")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Get() expected parameter error, got %v", err)
	}
}

func TestClient_Get_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	relatedClient := NewClient(c)

	ctx := context.Background()
	_, err := relatedClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("Get() expected error for server error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get related prefixes") {
		t.Errorf("Get() expected server error, got %v", err)
	}
}

func TestClient_Get_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	relatedClient := NewClient(c)

	ctx := context.Background()
	_, err := relatedClient.Get(ctx, "193.0.0.0/21")
	if err == nil {
		t.Fatal("Get() expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get related prefixes") {
		t.Errorf("Get() expected decode error, got %v", err)
	}
}

func TestGetRelatedPrefixes_Exported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"resource": "193.0.0.0/21",
				"prefixes": [],
				"query_time": "2025-07-03T06:44:00"
			},
			"status": "ok"
		}`))
	}))
	defer ts.Close()

	defaultClient := client.DefaultClient()
	originalBaseURL := defaultClient.BaseURL
	defer func() {
		defaultClient.BaseURL = originalBaseURL
	}()
	defaultClient.BaseURL = ts.URL
	defaultClient.HTTPClient = ts.Client()

	ctx := context.Background()
	result, err := GetRelatedPrefixes(ctx, "193.0.0.0/21")
	if err != nil {
		t.Fatalf("GetRelatedPrefixes() failed: %v", err)
	}

	if result == nil {
		t.Fatal("GetRelatedPrefixes() returned nil result")
	}

	if result.Data.Resource != "193.0.0.0/21" {
		t.Errorf("Resource = %v, want %v", result.Data.Resource, "193.0.0.0/21")
	}
}
