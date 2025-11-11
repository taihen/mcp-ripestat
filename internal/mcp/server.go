package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/taihen/mcp-ripestat/internal/mcp/consolidated"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// httpRequestKey is the context key for storing HTTP request information.
const httpRequestKey contextKey = "http_request"

// sessionIDKey is the context key for storing session ID information.
const sessionIDKey contextKey = "session_id"

// WithHTTPRequest stores an HTTP request in the context.
func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

// HTTPRequestFromContext retrieves an HTTP request from the context.
func HTTPRequestFromContext(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(httpRequestKey).(*http.Request)
	return r, ok
}

// WithSessionID stores a session ID in the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SessionIDFromContext retrieves a session ID from the context.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// Error message constants for parameter validation.
const (
	ErrResourceRequired     = "Error: resource parameter is required"
	ErrPrefixRequired       = "Error: prefix parameter is required"
	ErrLODParameterInvalid  = "Error: lod parameter must be 0 or 1"
	ErrLookBackLimitInvalid = "Error: look_back_limit parameter must be a valid integer"
)

// formatErrorMessage formats an error for tool results, avoiding duplicate "Error:" prefixes.
func formatErrorMessage(err error) string {
	errStr := err.Error()
	// If the error already starts with "Error:", don't add another prefix
	if strings.HasPrefix(errStr, "Error:") {
		return errStr
	}
	return fmt.Sprintf("Error: %v", err)
}

// getRequiredStringParam extracts a required string parameter from args.
func getRequiredStringParam(args map[string]interface{}, key, errorMsg string) (string, *ToolResult) {
	value, ok := args[key].(string)
	if !ok {
		return "", CreateToolResult(errorMsg, true)
	}
	return value, nil
}

// getOptionalStringParam extracts an optional string parameter from args.
func getOptionalStringParam(args map[string]interface{}, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

// validateLODParam validates and extracts the LOD parameter (0 or 1).
func validateLODParam(args map[string]interface{}) (int, *ToolResult) {
	lodStr, ok := args["lod"].(string)
	if !ok {
		return 0, nil // Default value when not provided
	}

	lod, err := strconv.Atoi(lodStr)
	if err != nil || (lod != 0 && lod != 1) {
		return 0, CreateToolResult(ErrLODParameterInvalid, true)
	}
	return lod, nil
}

// validateLookBackLimitParam validates and extracts the look_back_limit parameter.
func validateLookBackLimitParam(args map[string]interface{}) (int, *ToolResult) {
	lblStr, ok := args["look_back_limit"].(string)
	if !ok {
		return 0, nil // Default value when not provided
	}

	lookBackLimit, err := strconv.Atoi(lblStr)
	if err != nil {
		return 0, CreateToolResult(ErrLookBackLimitInvalid, true)
	}
	return lookBackLimit, nil
}

// Server represents an MCP server.
type Server struct {
	serverName          string
	serverVersion       string
	initialized         bool
	disableWhatsMyIP    bool
	globallyInitialized bool // For compatibility with older protocol versions
	consolidatedTools   *consolidated.Tools
}

// NewServer creates a new MCP server.
func NewServer(serverName, serverVersion string, disableWhatsMyIP bool) *Server {
	// Initialize consolidated tools with direct executor
	executor := consolidated.NewDirectExecutor()
	consolidatedTools := consolidated.NewTools(executor)

	return &Server{
		serverName:        serverName,
		serverVersion:     serverVersion,
		disableWhatsMyIP:  disableWhatsMyIP,
		consolidatedTools: consolidatedTools,
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
		if !s.initialized && !s.globallyInitialized {
			return NewErrorResponse(InitializationError, "Server not initialized", "Initialize first", req.ID), nil
		}
		return s.handleToolsList(req)
	case "tools/call":
		if !s.initialized && !s.globallyInitialized {
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

	// Determine if this is a legacy client based on protocol version
	isLegacyClient := params.ProtocolVersion != "" && params.ProtocolVersion < "2025-06-18"

	// Log server readiness for debugging cold starts
	slog.Info("MCP server responding to initialize request",
		"server_name", s.serverName,
		"version", s.serverVersion,
		"client_protocol", params.ProtocolVersion,
		"is_legacy", isLegacyClient)

	// For legacy protocol versions (< 2025-06-18), use simplified initialization
	if isLegacyClient {
		s.initialized = true
		s.globallyInitialized = true // Global initialization for compatibility
		slog.Info("auto-initialized server for legacy protocol version", "version", params.ProtocolVersion)

		result := CreateLegacyInitializeResult(s.serverName, s.serverVersion)
		return NewResponse(result, req.ID), nil
	}

	// For current protocol versions, validate and use full capabilities
	if params.ProtocolVersion != ProtocolVersion {
		slog.Warn("protocol version mismatch", "client", params.ProtocolVersion, "server", ProtocolVersion)
	}

	// Auto-initialize for better client compatibility
	s.initialized = true
	slog.Info("auto-initialized server for protocol version", "version", params.ProtocolVersion)

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

	// Route to consolidated tools
	switch params.Name {
	case "investigateResource":
		result, err := s.consolidatedTools.InvestigateResource(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "analyzeRouting":
		result, err := s.consolidatedTools.AnalyzeRouting(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "queryRegistry":
		result, err := s.consolidatedTools.QueryRegistry(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "validateSecurity":
		result, err := s.consolidatedTools.ValidateSecurity(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "exploreRelationships":
		result, err := s.consolidatedTools.ExploreRelationships(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "searchByLocation":
		result, err := s.consolidatedTools.SearchByLocation(ctx, args)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil

	case "getWhatsMyIP":
		if s.disableWhatsMyIP {
			return nil, fmt.Errorf("whats-my-ip tool is disabled")
		}
		return s.callWhatsMyIP(ctx, args)

	default:
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}
}

// Keep only the callWhatsMyIP implementation for the special case

func (s *Server) callWhatsMyIP(ctx context.Context, _ map[string]interface{}) (*ToolResult, error) {
	// Check if we have HTTP request context for client IP extraction
	if httpReq, ok := HTTPRequestFromContext(ctx); ok {
		// Extract client IP from HTTP headers for proxy scenarios
		clientIP := whatsmyip.ExtractClientIP(httpReq)
		slog.Debug("extracted client IP from HTTP request", "client_ip", clientIP, "remote_addr", httpReq.RemoteAddr)

		// Use the extracted client IP for whats-my-ip query
		result, err := whatsmyip.GetWhatsMyIPWithClientIP(ctx, clientIP)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil
	}

	// Fallback to standard behavior if no HTTP context available
	result, err := whatsmyip.GetWhatsMyIP(ctx)
	if err != nil {
		return CreateToolResult(formatErrorMessage(err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

// ParseQueryToRequest converts URL query parameters to JSON-RPC request.
func (s *Server) ParseQueryToRequest(query url.Values) ([]byte, error) {
	method := query.Get("method")
	if method == "" {
		return nil, fmt.Errorf("method parameter is required")
	}

	request := &Request{
		JSONRPC: "2.0",
		Method:  method,
		ID:      query.Get("id"),
	}

	if request.ID == "" {
		request.ID = "1"
	}

	// Parse parameters if provided.
	if paramsStr := query.Get("params"); paramsStr != "" {
		var params interface{}
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			return nil, fmt.Errorf("invalid params JSON: %w", err)
		}
		request.Params = params
	} else {
		// Convert individual query parameters to params object.
		params := make(map[string]interface{})
		for key, values := range query {
			if key == "method" || key == "id" || key == "params" {
				continue
			}
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
		if len(params) > 0 {
			request.Params = params
		}
	}

	return json.Marshal(request)
}
