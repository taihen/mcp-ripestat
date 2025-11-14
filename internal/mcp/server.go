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

type contextKey string

const httpRequestKey contextKey = "http_request"

const sessionIDKey contextKey = "session_id"

func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

func HTTPRequestFromContext(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(httpRequestKey).(*http.Request)
	return r, ok
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

const (
	ErrResourceRequired      = "Error: resource parameter is required"
	ErrPrefixRequired        = "Error: prefix parameter is required"
	ErrLODParameterInvalid   = "Error: lod parameter must be 0 or 1"
	ErrLookBackLimitInvalid  = "Error: look_back_limit parameter must be a valid integer"
	ErrMaxResultsInvalid     = "Error: max_results parameter must be a valid integer"
	ErrMaxResultsNonNegative = "Error: max_results parameter must be non-negative"
)

func formatErrorMessage(err error) string {
	errStr := err.Error()

	if strings.HasPrefix(errStr, "Error:") {
		return errStr
	}
	return fmt.Sprintf("Error: %v", err)
}

func getRequiredStringParam(args map[string]interface{}, key, errorMsg string) (string, *ToolResult) {
	value, ok := args[key].(string)
	if !ok {
		return "", CreateToolResult(errorMsg, true)
	}
	return value, nil
}

func getOptionalStringParam(args map[string]interface{}, key string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return ""
}

func validateLODParam(args map[string]interface{}) (int, *ToolResult) {
	lodStr, ok := args["lod"].(string)
	if !ok {
		return 0, nil
	}

	lod, err := strconv.Atoi(lodStr)
	if err != nil || (lod != 0 && lod != 1) {
		return 0, CreateToolResult(ErrLODParameterInvalid, true)
	}
	return lod, nil
}

func validateLookBackLimitParam(args map[string]interface{}) (int, *ToolResult) {
	lblStr, ok := args["look_back_limit"].(string)
	if !ok {
		return 0, nil
	}

	lookBackLimit, err := strconv.Atoi(lblStr)
	if err != nil {
		return 0, CreateToolResult(ErrLookBackLimitInvalid, true)
	}
	return lookBackLimit, nil
}

func validateMaxResultsParam(args map[string]interface{}) (int, *ToolResult) {
	maxResultsVal, ok := args["max_results"]
	if !ok {
		return 0, nil
	}

	var maxResults int
	switch v := maxResultsVal.(type) {
	case string:
		var err error
		maxResults, err = strconv.Atoi(v)
		if err != nil {
			return 0, CreateToolResult(ErrMaxResultsInvalid, true)
		}
	case float64:
		maxResults = int(v)
	case int:
		maxResults = v
	case int64:
		maxResults = int(v)
	default:
		return 0, CreateToolResult(ErrMaxResultsInvalid, true)
	}

	if maxResults < 0 {
		return 0, CreateToolResult(ErrMaxResultsNonNegative, true)
	}

	return maxResults, nil
}

type Server struct {
	serverName          string
	serverVersion       string
	initialized         bool
	disableWhatsMyIP    bool
	globallyInitialized bool
	consolidatedTools   *consolidated.Tools
}

func NewServer(serverName, serverVersion string, disableWhatsMyIP bool) *Server {

	executor := consolidated.NewDirectExecutor()
	consolidatedTools := consolidated.NewTools(executor)

	return &Server{
		serverName:        serverName,
		serverVersion:     serverVersion,
		disableWhatsMyIP:  disableWhatsMyIP,
		consolidatedTools: consolidatedTools,
	}
}

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

func (s *Server) handleNotification(_ context.Context, notif *Notification) (interface{}, error) {
	slog.Debug("handling notification", "method", notif.Method)

	switch notif.Method {
	case "initialized", "notifications/initialized":
		return s.handleInitialized(notif)
	case "notifications/cancelled":

		slog.Debug("received cancellation notification")
		return nil, nil
	default:
		slog.Warn("unknown notification method", "method", notif.Method)
		return nil, nil
	}
}

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

	isLegacyClient := params.ProtocolVersion != "" && params.ProtocolVersion < "2025-06-18"

	slog.Info("MCP server responding to initialize request",
		"server_name", s.serverName,
		"version", s.serverVersion,
		"client_protocol", params.ProtocolVersion,
		"is_legacy", isLegacyClient)

	if isLegacyClient {
		s.initialized = true
		s.globallyInitialized = true
		slog.Info("auto-initialized server for legacy protocol version", "version", params.ProtocolVersion)

		result := CreateLegacyInitializeResult(s.serverName, s.serverVersion)
		return NewResponse(result, req.ID), nil
	}

	if params.ProtocolVersion != ProtocolVersion {
		slog.Warn("protocol version mismatch", "client", params.ProtocolVersion, "server", ProtocolVersion)
	}

	s.initialized = true
	slog.Info("auto-initialized server for protocol version", "version", params.ProtocolVersion)

	result := CreateInitializeResult(s.serverName, s.serverVersion)
	return NewResponse(result, req.ID), nil
}

func (s *Server) handleInitialized(_ *Notification) (interface{}, error) {
	slog.Debug("handling initialized notification")
	s.initialized = true
	slog.Info("MCP server initialized successfully")
	return nil, nil
}

func (s *Server) handlePing(req *Request) (interface{}, error) {
	slog.Debug("handling ping request")
	return NewResponse(map[string]string{}, req.ID), nil
}

func (s *Server) handleToolsList(req *Request) (interface{}, error) {
	slog.Debug("handling tools/list request")

	toolsList := CreateToolsList()

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

func (s *Server) executeToolCall(ctx context.Context, params *CallToolParams) (*ToolResult, error) {
	slog.Debug("executing tool call", "tool", params.Name)

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

	if result := s.handleConsolidatedTool(ctx, params.Name, args); result != nil {
		return result, nil
	}

	if params.Name == "getWhatsMyIP" {
		if s.disableWhatsMyIP {
			return nil, fmt.Errorf("whats-my-ip tool is disabled")
		}
		return s.callWhatsMyIP(ctx, args)
	}

	if strings.HasPrefix(params.Name, "get") {
		return s.handleGetEndpoint(ctx, params.Name, args)
	}

	return nil, fmt.Errorf("unknown tool: %s", params.Name)
}

func (s *Server) handleConsolidatedTool(ctx context.Context, toolName string, args map[string]interface{}) *ToolResult {
	var result interface{}
	var err error

	switch toolName {
	case "investigateResource":
		result, err = s.consolidatedTools.InvestigateResource(ctx, args)
	case "analyzeRouting":
		result, err = s.consolidatedTools.AnalyzeRouting(ctx, args)
	case "queryRegistry":
		result, err = s.consolidatedTools.QueryRegistry(ctx, args)
	case "validateSecurity":
		result, err = s.consolidatedTools.ValidateSecurity(ctx, args)
	case "exploreRelationships":
		result, err = s.consolidatedTools.ExploreRelationships(ctx, args)
	case "searchByLocation":
		result, err = s.consolidatedTools.SearchByLocation(ctx, args)
	default:
		return nil
	}

	if err != nil {
		return CreateToolResult(formatErrorMessage(err), true)
	}
	return CreateToolResultFromJSON(result)
}

func (s *Server) handleGetEndpoint(ctx context.Context, endpointName string, args map[string]interface{}) (*ToolResult, error) {
	resource, ok := args["resource"].(string)
	if !ok || resource == "" {
		return CreateToolResult(ErrResourceRequired, true), nil
	}
	_ = resource // resource is validated but not used directly here

	if validationResult := s.validateGetEndpointParams(endpointName, args); validationResult != nil {
		return validationResult, nil
	}

	result, err := s.consolidatedTools.ExecuteIndividualEndpoint(ctx, endpointName, args)
	if err != nil {
		return CreateToolResult(formatErrorMessage(err), true), nil
	}
	return CreateToolResultFromJSON(result), nil
}

func (s *Server) validateGetEndpointParams(endpointName string, args map[string]interface{}) *ToolResult {
	switch endpointName {
	case "getRPKIValidation":
		prefix, ok := args["prefix"].(string)
		if !ok || prefix == "" {
			return CreateToolResult(ErrPrefixRequired, true)
		}
	case "getASNNeighbours":
		if _, lodResult := validateLODParam(args); lodResult != nil {
			return lodResult
		}
	case "getLookingGlass":
		if _, lblResult := validateLookBackLimitParam(args); lblResult != nil {
			return lblResult
		}
	case "getRoutingHistory":
		if _, maxResultsResult := validateMaxResultsParam(args); maxResultsResult != nil {
			return maxResultsResult
		}
	}
	return nil
}

func (s *Server) callWhatsMyIP(ctx context.Context, _ map[string]interface{}) (*ToolResult, error) {

	if httpReq, ok := HTTPRequestFromContext(ctx); ok {

		clientIP := whatsmyip.ExtractClientIP(httpReq)
		slog.Debug("extracted client IP from HTTP request", "client_ip", clientIP, "remote_addr", httpReq.RemoteAddr)

		result, err := whatsmyip.GetWhatsMyIPWithClientIP(ctx, clientIP)
		if err != nil {
			return CreateToolResult(formatErrorMessage(err), true), nil
		}
		return CreateToolResultFromJSON(result), nil
	}

	result, err := whatsmyip.GetWhatsMyIP(ctx)
	if err != nil {
		return CreateToolResult(formatErrorMessage(err), true), nil
	}

	return CreateToolResultFromJSON(result), nil
}

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

	if paramsStr := query.Get("params"); paramsStr != "" {
		var params interface{}
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			return nil, fmt.Errorf("invalid params JSON: %w", err)
		}
		request.Params = params
	} else {

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
