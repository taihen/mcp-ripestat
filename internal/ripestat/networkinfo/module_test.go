package networkinfo

import (
	"context"
	"testing"

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
