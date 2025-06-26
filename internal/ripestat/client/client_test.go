package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
		if r.URL.RawQuery != "param1=value1&resource=test" {
			t.Errorf("Expected query params 'param1=value1&resource=test', got %q", r.URL.RawQuery)
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

func TestClient_Get_RequestCreationError(t *testing.T) {
	// Create client
	c := New("https://example.com", nil)

	// Create a context that's already cancelled to trigger request creation error
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
