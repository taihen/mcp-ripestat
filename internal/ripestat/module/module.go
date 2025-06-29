// Package module provides a formal interface for RIPEstat API endpoint modules.
package module

import (
	"context"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

// RPCHandler is a function that handles JSON-RPC method calls.
type RPCHandler func(ctx context.Context, params interface{}) (interface{}, error)

// Module represents a RIPEstat API endpoint module.
type Module interface {
	// Name returns the module name (e.g., "network-info", "whois").
	Name() string

	// RegisterMethods registers RPC method handlers with the provided map.
	RegisterMethods(handlers map[string]RPCHandler)

	// EndpointPath returns the API endpoint path for this module.
	EndpointPath() string
}

// BaseModule provides common functionality for all modules.
type BaseModule struct {
	name         string
	endpointPath string
	client       *client.Client
	cache        *cache.Cache
}

// NewBaseModule creates a new BaseModule with dependency injection.
func NewBaseModule(name, endpointPath string, clientParam *client.Client, cacheParam *cache.Cache) *BaseModule {
	if clientParam == nil {
		clientParam = client.DefaultClient()
	}
	if cacheParam == nil {
		cacheParam = cache.New()
	}

	return &BaseModule{
		name:         name,
		endpointPath: endpointPath,
		client:       clientParam,
		cache:        cacheParam,
	}
}

// Name returns the module name.
func (m *BaseModule) Name() string {
	return m.name
}

// EndpointPath returns the API endpoint path.
func (m *BaseModule) EndpointPath() string {
	return m.endpointPath
}

// Client returns the HTTP client.
func (m *BaseModule) Client() *client.Client {
	return m.client
}

// Cache returns the cache instance.
func (m *BaseModule) Cache() *cache.Cache {
	return m.cache
}

// Registry manages module registration and method routing.
type Registry struct {
	modules  map[string]Module
	handlers map[string]RPCHandler
}

// NewRegistry creates a new module registry.
func NewRegistry() *Registry {
	return &Registry{
		modules:  make(map[string]Module),
		handlers: make(map[string]RPCHandler),
	}
}

// Register adds a module to the registry.
func (r *Registry) Register(module Module) {
	r.modules[module.Name()] = module
	module.RegisterMethods(r.handlers)
}

// GetModule returns a module by name.
func (r *Registry) GetModule(name string) (Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

// GetHandler returns a handler by method name.
func (r *Registry) GetHandler(method string) (RPCHandler, bool) {
	handler, exists := r.handlers[method]
	return handler, exists
}

// ListModules returns all registered module names.
func (r *Registry) ListModules() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// ListMethods returns all registered method names.
func (r *Registry) ListMethods() []string {
	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}
