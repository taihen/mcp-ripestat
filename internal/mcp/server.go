package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/taihen/mcp-ripestat/internal/ripestat/abusecontactfinder"
	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asnneighbours"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
	"github.com/taihen/mcp-ripestat/internal/ripestat/lookingglass"
	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkivalidation"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whois"
)

// Server represents an MCP server.
type Server struct {
	serverName       string
	serverVersion    string
	initialized      bool
	disableWhatsMyIP bool
}

// NewServer creates a new MCP server.
func NewServer(serverName, serverVersion string, disableWhatsMyIP bool) *Server {
	return &Server{
		serverName:       serverName,
		serverVersion:    serverVersion,
		disableWhatsMyIP: disableWhatsMyIP,
	}
}

// ProcessMessage processes an incoming MCP message.
func (s *Server) ProcessMessage(ctx context.Context, data []byte) (interface{}, error) {
	slog.Debug("processing MCP message", "data", string(data))

	msg, err := ParseMessage(data)
	if err != nil {
		slog.Error("failed to parse message", "err", err)
		return NewErrorResponse(ParseError, "Parse error", err.Error(), nil), nil
	}

	switch m := msg.(type) {
	case *Request:
		return s.handleRequest(ctx, m)
	case *Notification:
		return s.handleNotification(ctx, m)
	default:
		return NewErrorResponse(InvalidRequest, "Invalid request", "Unknown message type", nil), nil
	}
}

// handleRequest handles JSON-RPC requests.
func (s *Server) handleRequest(ctx context.Context, req *Request) (interface{}, error) {
	if err := req.ValidateRequest(); err != nil {
		return NewErrorResponse(InvalidRequest, "Invalid request", err.Error(), req.ID), nil
	}

	slog.Debug("handling request", "method", req.Method, "id", req.ID)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		if !s.initialized {
			return NewErrorResponse(InitializationError, "Server not initialized", "Initialize first", req.ID), nil
		}
		return s.handleToolsList(req)
	case "tools/call":
		if !s.initialized {
			return NewErrorResponse(InitializationError, "Server not initialized", "Initialize first", req.ID), nil
		}
		return s.handleToolsCall(ctx, req)
	case "ping":
		return s.handlePing(req)
	default:
		return NewErrorResponse(MethodNotFound, "Method not found", req.Method, req.ID), nil
	}
}

// handleNotification handles JSON-RPC notifications.
func (s *Server) handleNotification(_ context.Context, notif *Notification) (interface{}, error) {
	slog.Debug("handling notification", "method", notif.Method)

	switch notif.Method {
	case "initialized", "notifications/initialized":
		return s.handleInitialized(notif)
	case "notifications/cancelled":
		// Handle cancellation notifications
		slog.Debug("received cancellation notification")
		return nil, nil
	default:
		slog.Warn("unknown notification method", "method", notif.Method)
		return nil, nil
	}
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(req *Request) (interface{}, error) {
	slog.Info("handling initialize request - server ready for MCP client")

	var params InitializeParams
	if req.Params != nil {
		jsonData, err := json.Marshal(req.Params)
		if err != nil {
			return NewErrorResponse(InvalidParams, "Invalid params", err.Error(), req.ID), nil
		}
		if err := json.Unmarshal(jsonData, &params); err != nil {
			return NewErrorResponse(InvalidParams, "Invalid params", err.Error(), req.ID), nil
		}
	}

	// Validate protocol version
	if params.ProtocolVersion != ProtocolVersion {
		slog.Warn("protocol version mismatch", "client", params.ProtocolVersion, "server", ProtocolVersion)
	}

	// Log server readiness for debugging cold starts
	slog.Info("MCP server responding to initialize request",
		"server_name", s.serverName,
		"version", s.serverVersion,
		"client_protocol", params.ProtocolVersion)

	result := CreateInitializeResult(s.serverName, s.serverVersion)
	return NewResponse(result, req.ID), nil
}

// handleInitialized handles the initialized notification.
func (s *Server) handleInitialized(_ *Notification) (interface{}, error) {
	slog.Debug("handling initialized notification")
	s.initialized = true
	slog.Info("MCP server initialized successfully")
	return nil, nil
}

// handlePing handles ping requests.
func (s *Server) handlePing(req *Request) (interface{}, error) {
	slog.Debug("handling ping request")
	return NewResponse(map[string]string{}, req.ID), nil
}

// handleToolsList handles tools/list requests.
func (s *Server) handleToolsList(req *Request) (interface{}, error) {
	slog.Debug("handling tools/list request")

	toolsList := CreateToolsList()

	// Remove whats-my-ip tool if disabled
	if s.disableWhatsMyIP {
		tools := make([]Tool, 0, len(toolsList.Tools)-1)
		for _, tool := range toolsList.Tools {
			if tool.Name != "getWhatsMyIP" {
				tools = append(tools, tool)
			}
		}
		toolsList.Tools = tools
	}

	return NewResponse(toolsList, req.ID), nil
}

// handleToolsCall handles tools/call requests.
func (s *Server) handleToolsCall(ctx context.Context, req *Request) (interface{}, error) {
	slog.Debug("handling tools/call request")

	params, err := ParseCallToolParams(req.Params)
	if err != nil {
		return NewErrorResponse(InvalidParams, "Invalid params", err.Error(), req.ID), nil
	}

	result, err := s.executeToolCall(ctx, params)
	if err != nil {
		slog.Error("tool execution failed", "tool", params.Name, "err", err)
		return NewErrorResponse(ToolError, "Tool execution failed", err.Error(), req.ID), nil
	}

	return NewResponse(result, req.ID), nil
}

// executeToolCall executes a tool call.
func (s *Server) executeToolCall(ctx context.Context, params *CallToolParams) (*ToolResult, error) {
	slog.Debug("executing tool call", "tool", params.Name)

	// Parse arguments
	args := make(map[string]interface{})
	if params.Arguments != nil {
		jsonData, err := json.Marshal(params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}
		if err := json.Unmarshal(jsonData, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}
	}

	switch params.Name {
	case "getNetworkInfo":
		return s.callNetworkInfo(ctx, args)
	case "getASOverview":
		return s.callASOverview(ctx, args)
	case "getAnnouncedPrefixes":
		return s.callAnnouncedPrefixes(ctx, args)
	case "getRoutingStatus":
		return s.callRoutingStatus(ctx, args)
	case "getWhois":
		return s.callWhois(ctx, args)
	case "getAbuseContactFinder":
		return s.callAbuseContactFinder(ctx, args)
	case "getRPKIValidation":
		return s.callRPKIValidation(ctx, args)
	case "getASNNeighbours":
		return s.callASNNeighbours(ctx, args)
	case "getLookingGlass":
		return s.callLookingGlass(ctx, args)
	case "getWhatsMyIP":
		if s.disableWhatsMyIP {
			return nil, fmt.Errorf("whats-my-ip tool is disabled")
		}
		return s.callWhatsMyIP(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}
}

// Tool call implementations.

func (s *Server) callNetworkInfo(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := networkinfo.GetNetworkInfo(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callASOverview(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := asoverview.GetASOverview(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callAnnouncedPrefixes(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := announcedprefixes.GetAnnouncedPrefixes(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callRoutingStatus(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := routingstatus.GetRoutingStatus(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callWhois(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := whois.GetWhois(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callAbuseContactFinder(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	result, err := abusecontactfinder.GetAbuseContactFinder(ctx, resource)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callRPKIValidation(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	prefix, ok := args["prefix"].(string)
	if !ok {
		return nil, fmt.Errorf("prefix parameter is required")
	}

	result, err := rpkivalidation.GetRPKIValidation(ctx, resource, prefix)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callASNNeighbours(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	lod := 0
	if lodStr, ok := args["lod"].(string); ok {
		var err error
		lod, err = strconv.Atoi(lodStr)
		if err != nil || (lod != 0 && lod != 1) {
			return nil, fmt.Errorf("lod parameter must be 0 or 1")
		}
	}

	queryTime := ""
	if qt, ok := args["query_time"].(string); ok {
		queryTime = qt
	}

	result, err := asnneighbours.GetASNNeighbours(ctx, resource, lod, queryTime)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callLookingGlass(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter is required")
	}

	lookBackLimit := 0
	if lblStr, ok := args["look_back_limit"].(string); ok {
		var err error
		lookBackLimit, err = strconv.Atoi(lblStr)
		if err != nil {
			return nil, fmt.Errorf("look_back_limit parameter must be a valid integer")
		}
	}

	result, err := lookingglass.GetLookingGlass(ctx, resource, lookBackLimit)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

func (s *Server) callWhatsMyIP(ctx context.Context, _ map[string]interface{}) (*ToolResult, error) {
	result, err := whatsmyip.GetWhatsMyIP(ctx)
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error: %v", err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}
