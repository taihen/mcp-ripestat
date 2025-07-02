package networkinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/module"
)

func TestNewModule(t *testing.T) {
	m := NewModule(nil, nil)
	if m == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if m.Name() != "network-info" {
		t.Errorf("Expected module name to be 'network-info', got %s", m.Name())
	}

	if m.EndpointPath() != EndpointPath {
		t.Errorf("Expected endpoint path to be %s, got %s", EndpointPath, m.EndpointPath())
	}
}

func TestModuleImplementsInterface(_ *testing.T) {
	m := NewModule(nil, nil)

	// Verify it implements the Module interface
	var _ module.Module = m
}

func TestRegisterMethods(t *testing.T) {
	m := NewModule(nil, nil)
	handlers := make(map[string]module.RPCHandler)

	// Register methods
	m.RegisterMethods(handlers)

	// Should have registered the getNetworkInfo method
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}

	handler, exists := handlers["getNetworkInfo"]
	if !exists {
		t.Error("Expected getNetworkInfo handler to be registered")
	}

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
}

func TestHandleGetNetworkInfo_InvalidParams(t *testing.T) {
	m := NewModule(nil, nil)
	ctx := context.Background()

	// Test with nil params
	result, err := m.handleGetNetworkInfo(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil params")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	// Test with invalid params type
	result, err = m.handleGetNetworkInfo(ctx, "invalid")
	if err == nil {
		t.Error("Expected error for invalid params type")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestHandleGetNetworkInfo_MissingResource(t *testing.T) {
	m := NewModule(nil, nil)
	ctx := context.Background()

	// Test with empty params map
	params := map[string]interface{}{}
	result, err := m.handleGetNetworkInfo(ctx, params)
	if err == nil {
		t.Error("Expected error for missing resource")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	// Test with nil resource
	params = map[string]interface{}{
		"resource": nil,
	}
	result, err = m.handleGetNetworkInfo(ctx, params)
	if err == nil {
		t.Error("Expected error for nil resource")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	// Test with empty resource
	params = map[string]interface{}{
		"resource": "",
	}
	result, err = m.handleGetNetworkInfo(ctx, params)
	if err == nil {
		t.Error("Expected error for empty resource")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

func TestHandleGetNetworkInfo_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"prefix": "193.0.0.0/21",
				"resource": "193.0.0.0/21"
			},
			"status": "ok"
		}`))
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	m := NewModule(c, nil)
	ctx := context.Background()

	params := map[string]interface{}{
		"resource": "193.0.0.0/21",
	}

	result, err := m.handleGetNetworkInfo(ctx, params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected result to be *Response, got %T", result)
	}
	if response.Data.Prefix != "193.0.0.0/21" {
		t.Errorf("Expected prefix 193.0.0.0/21, got %s", response.Data.Prefix)
	}
}

func TestHandleGetNetworkInfo_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	c := client.New(ts.URL, ts.Client())
	m := NewModule(c, nil)
	ctx := context.Background()

	params := map[string]interface{}{
		"resource": "193.0.0.0/21",
	}

	result, err := m.handleGetNetworkInfo(ctx, params)
	if err == nil {
		t.Fatal("Expected error for HTTP error")
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}
