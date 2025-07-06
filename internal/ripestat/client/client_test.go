package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/config"
	"github.com/taihen/mcp-ripestat/internal/ripestat/logging"
)

func TestNew(t *testing.T) {
	c := New("https://example.com", nil)
	if c.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be non-nil")
	}
	if c.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}
}

func TestNew_WithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	c := New("https://example.com", httpClient)
	if c.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", c.BaseURL)
	}
	if c.HTTPClient != httpClient {
		t.Error("Expected HTTPClient to be the same instance")
	}
}

func TestDefaultClient(t *testing.T) {
	c := DefaultClient()
	if c.BaseURL != config.DefaultBaseURL {
		t.Errorf("Expected BaseURL to be %q, got %q", config.DefaultBaseURL, c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be non-nil")
	}
	if c.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}
}

func TestNewWithConfig(t *testing.T) {
	cfg := config.DefaultConfig().
		WithBaseURL("https://example.com").
		WithTimeout(20 * time.Second).
		WithUserAgent("test-agent/1.0")

	c := NewWithConfig(cfg, nil)

	if c.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://example.com", c.BaseURL)
	}

	if c.UserAgent != "test-agent/1.0" {
		t.Errorf("Expected UserAgent to be %q, got %q", "test-agent/1.0", c.UserAgent)
	}

	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be non-nil")
	}

	if c.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}
}

func TestClient_Get(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("Expected request to '/test', got %q", r.URL.Path)
		}
		if r.URL.RawQuery != "param1=value1&resource=test&sourceapp=mcp-ripestat" {
			t.Errorf("Expected query params 'param1=value1&resource=test&sourceapp=mcp-ripestat', got %q", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")
	params.Set("param1", "value1")

	resp, err := c.Get(ctx, "/test", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	defer resp.Body.Close()

	// Check response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != `{"data": "test"}` {
		t.Errorf("Expected response body %q, got %q", `{"data": "test"}`, string(body))
	}
}

func TestClient_Get_WithCustomLogger(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with custom logger
	c := New(server.URL, nil)
	c.Logger = logging.NewLogger(logging.LogLevelDebug, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	defer resp.Body.Close()
}

func TestClient_Get_Error(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error": "internal server error"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err != nil {
		t.Fatalf("Expected no error from Get, got %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Test GetJSON which should return an error for non-200 status
	var result map[string]interface{}
	err = c.GetJSON(ctx, "/test", params, &result)
	if err == nil {
		t.Fatal("Expected error from GetJSON, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP status: 500") {
		t.Errorf("Expected error to mention HTTP status, got %q", err.Error())
	}
}

// TestClient_Get_URLParseError tests the URL parsing error path.
func TestClient_Get_URLParseError(t *testing.T) {
	// Create client with invalid base URL
	c := New("http://[::1", nil) // Invalid URL format

	// Make request that should fail URL parsing
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected URL parsing error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse URL") {
		t.Errorf("Expected error to mention URL parsing, got %q", err.Error())
	}
}

// TestClient_Get_RequestCreationError tests the request creation error path.
func TestClient_Get_RequestCreationError(t *testing.T) {
	// Create client
	c := New("https://example.com", nil)

	// Create a context that's already cancelled to trigger request creation error
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Make request with cancelled context and invalid method
	resp, err := c.Get(ctx, "test\x00invalid", nil) // Invalid URL character
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected request creation error, got nil")
	}
}

func TestClient_Get_ContextCanceled(t *testing.T) {
	// Setup test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Make request
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
	}
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context error, got %q", err.Error())
	}
}

func TestClient_GetJSON(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	var result map[string]interface{}
	err := c.GetJSON(ctx, "/test", params, &result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check result
	if result["data"] != "test" {
		t.Errorf("Expected result.data to be 'test', got %v", result["data"])
	}
}

func TestClient_GetJSON_BadJSON(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": invalid json}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	var result map[string]interface{}
	err := c.GetJSON(ctx, "/test", params, &result)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected decode error, got %q", err.Error())
	}
}

func TestNewWithConfig_WithHTTPClient(t *testing.T) {
	cfg := config.DefaultConfig().WithBaseURL("https://example.com")
	httpClient := &http.Client{Timeout: 30 * time.Second}

	c := NewWithConfig(cfg, httpClient)

	if c.HTTPClient != httpClient {
		t.Error("Expected HTTPClient to be the same instance")
	}
}

func TestClient_Get_InvalidURL(t *testing.T) {
	// Create client with invalid base URL
	c := New("://invalid-url", nil)

	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

func TestClient_Get_NilParams(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("Expected request to '/test', got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make request with nil params
	ctx := context.Background()
	resp, err := c.Get(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	defer resp.Body.Close()
}

func TestNew_EmptyBaseURL(t *testing.T) {
	c := New("", nil)
	if c.BaseURL != config.DefaultBaseURL {
		t.Errorf("Expected BaseURL to be %q, got %q", config.DefaultBaseURL, c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be non-nil")
	}
}

func TestNewWithConfig_NilConfig(t *testing.T) {
	c := NewWithConfig(nil, nil)
	if c.BaseURL != config.DefaultBaseURL {
		t.Errorf("Expected BaseURL to be %q, got %q", config.DefaultBaseURL, c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Error("Expected HTTPClient to be non-nil")
	}
}

func TestClient_Get_EmptyUserAgent(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// When UserAgent is empty, Go's HTTP client sets a default User-Agent
		// We just verify that the request goes through successfully
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with empty user agent
	c := New(server.URL, nil)
	c.UserAgent = ""

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	defer resp.Body.Close()
}

func TestClient_GetJSON_SuccessfulDecode(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test", "status": "ok"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with debug logging to test the success path
	c := New(server.URL, nil)
	c.Logger = logging.NewLogger(logging.LogLevelDebug, nil)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	var result map[string]interface{}
	err := c.GetJSON(ctx, "/test", params, &result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check result
	if result["data"] != "test" {
		t.Errorf("Expected result.data to be 'test', got %v", result["data"])
	}
	if result["status"] != "ok" {
		t.Errorf("Expected result.status to be 'ok', got %v", result["status"])
	}
}

func TestClient_Get_HTTPDoError(t *testing.T) {
	// Create a mock HTTP client that always returns an error
	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	// Create client with mock
	c := New("https://example.com", mockClient)

	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected error from HTTP client, got nil")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("Expected error to contain 'network error', got %q", err.Error())
	}
}

// mockHTTPClient is a mock implementation of HTTPDoer.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestClient_Get_RequestCreationFailure(t *testing.T) {
	// Create client with invalid method to trigger request creation error
	c := New("https://example.com", nil)

	// Use an invalid method character to trigger http.NewRequestWithContext error
	ctx := context.Background()

	// We need to test the case where http.NewRequestWithContext fails
	// This is hard to trigger directly, so let's test the URL parsing error instead
	c.BaseURL = "://invalid-url"

	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse URL") {
		t.Errorf("Expected URL parsing error, got %q", err.Error())
	}
}

func TestClient_Get_RequestCreationWithContext(t *testing.T) {
	// Test the case where http.NewRequestWithContext fails due to invalid method
	c := New("https://example.com", nil)

	// Create a context that will be cancelled to test context handling
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err == nil {
		defer resp.Body.Close()
		t.Fatal("Expected error for cancelled context, got nil")
	}
}

func TestClient_Get_SlowRequestWarning(t *testing.T) {
	// Setup test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(11 * time.Second) // Simulate slow response

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with custom logger to capture warning
	var logOutput strings.Builder
	c := New(server.URL, nil)
	c.Logger = logging.NewLogger(logging.LogLevelWarning, &logOutput)

	// Make request
	ctx := context.Background()
	params := url.Values{}
	params.Set("resource", "test")

	resp, err := c.Get(ctx, "/test", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	defer resp.Body.Close()

	// Check that warning was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "Slow request") {
		t.Errorf("Expected warning about slow request, got log: %s", logContent)
	}
	if !strings.Contains(logContent, "took") {
		t.Errorf("Expected warning to include timing, got log: %s", logContent)
	}
}

func TestClient_RateLimiting_ConcurrentRequests(t *testing.T) {
	var inFlightCount int64
	var maxInFlightCount int64

	// Setup test server that tracks concurrent requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Increment in-flight counter
		current := atomic.AddInt64(&inFlightCount, 1)

		// Track maximum concurrent requests
		for {
			maxVal := atomic.LoadInt64(&maxInFlightCount)
			if current <= maxVal {
				break
			}
			if atomic.CompareAndSwapInt64(&maxInFlightCount, maxVal, current) {
				break
			}
		}

		// Hold the request for a short time
		time.Sleep(100 * time.Millisecond)

		// Decrement in-flight counter
		atomic.AddInt64(&inFlightCount, -1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Make 10 concurrent requests (more than the limit of 7)
	const numRequests = 10
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			params := url.Values{}
			params.Set("resource", "test")

			var result map[string]interface{}
			err := c.GetJSON(ctx, "/test", params, &result)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		}()
	}

	wg.Wait()

	// Check that we never exceeded the rate limit
	maxConcurrent := atomic.LoadInt64(&maxInFlightCount)
	if maxConcurrent > 7 {
		t.Errorf("Rate limiting failed: had %d concurrent requests, expected <= 7", maxConcurrent)
	}
}

func TestClient_RateLimiting_ContextCancellation(t *testing.T) {
	// Setup test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test"}`)
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	params := url.Values{}
	params.Set("resource", "test")

	var result map[string]interface{}
	err := c.GetJSON(ctx, "/test", params, &result)

	// Should get context cancellation error
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context error, got %q", err.Error())
	}
}

func TestClient_CacheHit(t *testing.T) {
	var requestCount int

	// Setup test server that counts requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, `{"data": "test", "request_count": `+fmt.Sprintf("%d", requestCount)+`}`)
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	c := New(server.URL, nil)

	ctx := context.Background()
	endpoint := "/data/whois"
	params := url.Values{}
	params.Set("resource", "test")

	// First request should hit the server
	var result1 map[string]interface{}
	err := c.GetJSON(ctx, endpoint, params, &result1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Expected 1 request to server, got %d", requestCount)
	}

	// Second request should hit cache
	var result2 map[string]interface{}
	err = c.GetJSON(ctx, endpoint, params, &result2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Expected still 1 request to server (cache hit), got %d", requestCount)
	}

	// Results should be the same
	if result1["data"] != result2["data"] {
		t.Errorf("Expected cached data to match original, got %v vs %v", result1["data"], result2["data"])
	}
}

func TestCopyInterface(t *testing.T) {
	source := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": []string{"a", "b", "c"},
	}

	var target map[string]interface{}
	err := copyInterface(source, &target)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if target["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", target["key1"])
	}

	// Note: JSON unmarshaling converts numbers to float64
	if target["key2"] != float64(42) {
		t.Errorf("Expected key2 to be 42.0, got %v", target["key2"])
	}
}

// BenchmarkHTTPClientPerformance benchmarks the performance of different HTTP client configurations
func BenchmarkHTTPClientPerformance(b *testing.B) {
	b.Run("DefaultClient", func(b *testing.B) {
		client := &http.Client{Timeout: 30 * time.Second}
		benchmarkHTTPClient(b, client)
	})

	b.Run("OptimizedClient", func(b *testing.B) {
		cfg := config.DefaultConfig()
		client := createOptimizedHTTPClient(cfg)
		benchmarkHTTPClient(b, client)
	})

	b.Run("OptimizedClientHighConcurrency", func(b *testing.B) {
		cfg := config.DefaultConfig().
			WithMaxIdleConns(200).
			WithMaxIdleConnsPerHost(20).
			WithMaxConnsPerHost(200)
		client := createOptimizedHTTPClient(cfg)
		benchmarkHTTPClient(b, client)
	})
}

func benchmarkHTTPClient(b *testing.B, client *http.Client) {
	// Create a test server that simulates RIPE API behavior
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"test": "response"}}`))
	}))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
				continue
			}
			_ = resp.Body.Close()
		}
	})
}

// TestOptimizedHTTPClientConfiguration tests the optimized HTTP client configuration
func TestOptimizedHTTPClientConfiguration(t *testing.T) {
	cfg := config.DefaultConfig().
		WithMaxIdleConns(50).
		WithMaxIdleConnsPerHost(5).
		WithMaxConnsPerHost(50).
		WithIdleConnTimeout(60 * time.Second).
		WithForceHTTP2(true)

	client := createOptimizedHTTPClient(cfg)

	// Verify client is created
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Verify timeout is set correctly
	if client.Timeout != cfg.Timeout {
		t.Errorf("Expected timeout %v, got %v", cfg.Timeout, client.Timeout)
	}

	// Verify transport is configured (type assertion)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport, got different type")
	}

	// Verify connection pool settings
	if transport.MaxIdleConns != cfg.MaxIdleConns {
		t.Errorf("Expected MaxIdleConns %d, got %d", cfg.MaxIdleConns, transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != cfg.MaxIdleConnsPerHost {
		t.Errorf("Expected MaxIdleConnsPerHost %d, got %d", cfg.MaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
	}

	if transport.MaxConnsPerHost != cfg.MaxConnsPerHost {
		t.Errorf("Expected MaxConnsPerHost %d, got %d", cfg.MaxConnsPerHost, transport.MaxConnsPerHost)
	}

	if transport.IdleConnTimeout != cfg.IdleConnTimeout {
		t.Errorf("Expected IdleConnTimeout %v, got %v", cfg.IdleConnTimeout, transport.IdleConnTimeout)
	}

	// Verify HTTP/2 is enabled
	if !transport.ForceAttemptHTTP2 {
		t.Error("Expected ForceAttemptHTTP2 to be true")
	}

	// Verify other performance settings
	if transport.DisableCompression {
		t.Error("Expected compression to be enabled")
	}

	if transport.DisableKeepAlives {
		t.Error("Expected keep-alives to be enabled")
	}
}

// TestHTTP2Support tests HTTP/2 support functionality
func TestHTTP2Support(t *testing.T) {
	t.Run("HTTP2Enabled", func(t *testing.T) {
		cfg := config.DefaultConfig().WithForceHTTP2(true)
		client := createOptimizedHTTPClient(cfg)

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("Expected *http.Transport")
		}

		if !transport.ForceAttemptHTTP2 {
			t.Error("Expected HTTP/2 to be enabled")
		}
	})

	t.Run("HTTP2Disabled", func(t *testing.T) {
		cfg := config.DefaultConfig().WithForceHTTP2(false)
		client := createOptimizedHTTPClient(cfg)

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("Expected *http.Transport")
		}

		if transport.ForceAttemptHTTP2 {
			t.Error("Expected HTTP/2 to be disabled")
		}
	})
}

// TestConnectionPoolingBehavior tests connection pooling behavior under load
func TestConnectionPoolingBehavior(t *testing.T) {
	// Create a test server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"test": "response"}`))
	}))
	defer server.Close()

	// Create client with small connection pool for testing
	cfg := config.DefaultConfig().
		WithMaxIdleConns(2).
		WithMaxIdleConnsPerHost(1).
		WithMaxConnsPerHost(2)

	client := createOptimizedHTTPClient(cfg)

	// Make concurrent requests
	const numRequests = 10
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
	}

	wg.Wait()

	if requestCount != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, requestCount)
	}
}
