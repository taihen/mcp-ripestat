// Package networkinfo provides a modular implementation for the RIPEstat network-info API.
package networkinfo

import (
	"context"
	"fmt"
	"net/url"

	"github.com/taihen/mcp-ripestat/internal/ripestat/cache"
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
	"github.com/taihen/mcp-ripestat/internal/ripestat/errors"
	"github.com/taihen/mcp-ripestat/internal/ripestat/module"
)

// Module implements the Module interface for network-info API.
type Module struct {
	*module.BaseModule
}

// NewModule creates a new Module with dependency injection.
func NewModule(client *client.Client, cache *cache.Cache) *Module {
	return &Module{
		BaseModule: module.NewBaseModule("network-info", EndpointPath, client, cache),
	}
}

// RegisterMethods registers RPC method handlers for the network-info module.
func (m *Module) RegisterMethods(handlers map[string]module.RPCHandler) {
	handlers["getNetworkInfo"] = m.handleGetNetworkInfo
}

// handleGetNetworkInfo handles the getNetworkInfo RPC method.
func (m *Module) handleGetNetworkInfo(ctx context.Context, params interface{}) (interface{}, error) {
	// Parse parameters
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("invalid parameters"))
	}

	resource, ok := paramsMap["resource"].(string)
	if !ok || resource == "" {
		return nil, errors.ErrInvalidParameter.WithError(fmt.Errorf("resource parameter is required"))
	}

	// Build URL parameters
	urlParams := url.Values{}
	urlParams.Set("resource", resource)

	// Make the API request
	var response Response
	if err := m.Client().GetJSON(ctx, m.EndpointPath(), urlParams, &response); err != nil {
		return nil, errors.ErrServerError.WithError(fmt.Errorf("failed to get network information: %w", err))
	}

	return &response, nil
}
