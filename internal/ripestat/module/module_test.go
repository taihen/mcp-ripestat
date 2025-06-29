package module

import (
	"context"
	"errors"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

// mockModule implements the Module interface for testing.
type mockModule struct {
	*BaseModule
}

func newMockModule() *mockModule {
	return &mockModule{
		BaseModule: NewBaseModule("test-module", "/data/test", nil, nil),
	}
}

func (m *mockModule) RegisterMethods(handlers map[string]RPCHandler) {
	handlers["test.method"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return "test-response", nil
	}
	handlers["test.error"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("test error")
	}
}

func TestNewBaseModule(t *testing.T) {
	name := "test-module"
	endpointPath := "/data/test"

	module := NewBaseModule(name, endpointPath, nil, nil)
	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.Name() != name {
		t.Errorf("Expected name to be %s, got %s", name, module.Name())
	}

	if module.EndpointPath() != endpointPath {
		t.Errorf("Expected endpoint path to be %s, got %s", endpointPath, module.EndpointPath())
	}

	if module.Client() == nil {
		t.Error("Expected client to be non-nil (should use default)")
	}

	if module.Cache() == nil {
		t.Error("Expected cache to be non-nil (should use default)")
	}
}

func TestNewBaseModuleWithDependencies(t *testing.T) {
	name := "test-module"
	endpointPath := "/data/test"
	client := client.DefaultClient()
	cache := cache.New()

	module := NewBaseModule(name, endpointPath, client, cache)
	if module == nil {
		t.Fatal("Expected module to be non-nil")
	}

	if module.Client() != client {
		t.Error("Expected client to be the same instance")
	}

	if module.Cache() != cache {
		t.Error("Expected cache to be the same instance")
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be non-nil")
	}

	// Should be empty initially
	if len(registry.ListModules()) != 0 {
		t.Errorf("Expected empty module list, got %d modules", len(registry.ListModules()))
	}

	if len(registry.ListMethods()) != 0 {
		t.Errorf("Expected empty method list, got %d methods", len(registry.ListMethods()))
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	module := newMockModule()

	// Register the module
	registry.Register(module)

	// Check that module was registered
	modules := registry.ListModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}

	if modules[0] != "test-module" {
		t.Errorf("Expected module name to be 'test-module', got %s", modules[0])
	}

	// Check that methods were registered
	methods := registry.ListMethods()
	if len(methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(methods))
	}

	// Check specific methods
	expectedMethods := map[string]bool{
		"test.method": false,
		"test.error":  false,
	}

	for _, method := range methods {
		if _, exists := expectedMethods[method]; exists {
			expectedMethods[method] = true
		} else {
			t.Errorf("Unexpected method: %s", method)
		}
	}

	for method, found := range expectedMethods {
		if !found {
			t.Errorf("Expected method %s not found", method)
		}
	}
}

func TestRegistryGetModule(t *testing.T) {
	registry := NewRegistry()
	module := newMockModule()

	// Register the module
	registry.Register(module)

	// Get the module
	retrieved, exists := registry.GetModule("test-module")
	if !exists {
		t.Fatal("Expected module to exist")
	}

	if retrieved != module {
		t.Error("Expected retrieved module to be the same instance")
	}

	// Try to get non-existent module
	_, exists = registry.GetModule("non-existent")
	if exists {
		t.Error("Expected non-existent module to not exist")
	}
}

func TestRegistryGetHandler(t *testing.T) {
	registry := NewRegistry()
	module := newMockModule()

	// Register the module
	registry.Register(module)

	// Get a handler
	handler, exists := registry.GetHandler("test.method")
	if !exists {
		t.Fatal("Expected handler to exist")
	}

	if handler == nil {
		t.Fatal("Expected handler to be non-nil")
	}

	// Test the handler
	ctx := context.Background()
	result, err := handler(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "test-response" {
		t.Errorf("Expected 'test-response', got %v", result)
	}

	// Try to get non-existent handler
	_, exists = registry.GetHandler("non-existent")
	if exists {
		t.Error("Expected non-existent handler to not exist")
	}
}

func TestRegistryHandlerError(t *testing.T) {
	registry := NewRegistry()
	module := newMockModule()

	// Register the module
	registry.Register(module)

	// Get the error handler
	handler, exists := registry.GetHandler("test.error")
	if !exists {
		t.Fatal("Expected error handler to exist")
	}

	// Test the handler
	ctx := context.Background()
	result, err := handler(ctx, nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err.Error())
	}
}

// mockModule1 and mockModule2 with unique method names.
type mockModule1 struct {
	*BaseModule
}

func (m *mockModule1) RegisterMethods(handlers map[string]RPCHandler) {
	handlers["module1.method"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return "module1-response", nil
	}
	handlers["module1.error"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("module1 error")
	}
}

type mockModule2 struct {
	*BaseModule
}

func (m *mockModule2) RegisterMethods(handlers map[string]RPCHandler) {
	handlers["module2.method"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return "module2-response", nil
	}
	handlers["module2.error"] = func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("module2 error")
	}
}

func TestMultipleModuleRegistration(t *testing.T) {
	registry := NewRegistry()

	// Create modules with unique method names
	module1 := &mockModule1{
		BaseModule: NewBaseModule("module1", "/data/test1", nil, nil),
	}
	module2 := &mockModule2{
		BaseModule: NewBaseModule("module2", "/data/test2", nil, nil),
	}

	// Register both modules
	registry.Register(module1)
	registry.Register(module2)

	// Check module count
	modules := registry.ListModules()
	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}

	// Check method count (each mock module registers 2 methods)
	methods := registry.ListMethods()
	if len(methods) != 4 {
		t.Errorf("Expected 4 methods, got %d", len(methods))
	}

	// Verify we can get both modules
	_, exists1 := registry.GetModule("module1")
	_, exists2 := registry.GetModule("module2")

	if !exists1 {
		t.Error("Expected module1 to exist")
	}
	if !exists2 {
		t.Error("Expected module2 to exist")
	}
}
