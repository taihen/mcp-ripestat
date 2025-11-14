package module

import (
	"context"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
)

type RPCHandler func(ctx context.Context, params interface{}) (interface{}, error)

type Module interface {
	Name() string

	RegisterMethods(handlers map[string]RPCHandler)

	EndpointPath() string
}

type BaseModule struct {
	name         string
	endpointPath string
	client       *client.Client
	cache        *cache.Cache
}

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

func (m *BaseModule) Name() string {
	return m.name
}

func (m *BaseModule) EndpointPath() string {
	return m.endpointPath
}

func (m *BaseModule) Client() *client.Client {
	return m.client
}

func (m *BaseModule) Cache() *cache.Cache {
	return m.cache
}

type Registry struct {
	modules  map[string]Module
	handlers map[string]RPCHandler
}

func NewRegistry() *Registry {
	return &Registry{
		modules:  make(map[string]Module),
		handlers: make(map[string]RPCHandler),
	}
}

func (r *Registry) Register(module Module) {
	r.modules[module.Name()] = module
	module.RegisterMethods(r.handlers)
}

func (r *Registry) GetModule(name string) (Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

func (r *Registry) GetHandler(method string) (RPCHandler, bool) {
	handler, exists := r.handlers[method]
	return handler, exists
}

func (r *Registry) ListModules() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

func (r *Registry) ListMethods() []string {
	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}
