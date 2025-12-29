package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/taihen/mcp-ripestat/internal/mcp/consolidated"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
)

// SDKServer wraps the official MCP SDK server with RIPEstat-specific functionality.
type SDKServer struct {
	mcpServer         *mcp.Server
	consolidatedTools *consolidated.Tools
	disableWhatsMyIP  bool
}

// NewSDKServer creates a new MCP server using the official SDK.
func NewSDKServer(serverName, serverVersion string, disableWhatsMyIP bool) *SDKServer {
	impl := &mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}

	mcpServer := mcp.NewServer(impl, &mcp.ServerOptions{
		Logger: slog.Default(),
	})

	executor := consolidated.NewDirectExecutor()
	consolidatedTools := consolidated.NewTools(executor)

	s := &SDKServer{
		mcpServer:         mcpServer,
		consolidatedTools: consolidatedTools,
		disableWhatsMyIP:  disableWhatsMyIP,
	}

	s.registerTools()

	return s
}

// MCPServer returns the underlying MCP SDK server.
func (s *SDKServer) MCPServer() *mcp.Server {
	return s.mcpServer
}

// registerTools registers all RIPEstat tools with the MCP server.
// Only consolidated tools + getWhatsMyIP are exposed (7 tools total).
func (s *SDKServer) registerTools() {
	// Register consolidated tools (6 tools)
	for toolName, schema := range consolidated.ConsolidatedSchemas {
		description := consolidated.ConsolidatedToolDescriptions[toolName]
		if description == "" {
			description = "Consolidated tool for RIPEstat operations"
		}

		schemaJSON, err := json.Marshal(schema)
		if err != nil {
			slog.Error("failed to marshal schema", "tool", toolName, "err", err)
			continue
		}

		s.mcpServer.AddTool(
			&mcp.Tool{
				Name:        toolName,
				Description: description,
				InputSchema: json.RawMessage(schemaJSON),
			},
			s.createConsolidatedToolHandler(toolName),
		)
	}

	// Register getWhatsMyIP tool
	if !s.disableWhatsMyIP {
		s.mcpServer.AddTool(
			&mcp.Tool{
				Name:        "getWhatsMyIP",
				Description: "Get the caller's public IP address. Respects X-Forwarded-For headers when behind a proxy.",
				InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
			},
			s.handleGetWhatsMyIP,
		)
	}
}

// createConsolidatedToolHandler creates a handler for consolidated tools.
func (s *SDKServer) createConsolidatedToolHandler(toolName string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := make(map[string]interface{})
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: failed to parse arguments: %v", err)}},
					IsError: true,
				}, nil
			}
		}

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
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: unknown consolidated tool: %s", toolName)}},
				IsError: true,
			}, nil
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: formatToolError(err)}},
				IsError: true,
			}, nil
		}

		return createToolResultFromJSON(result)
	}
}

// handleGetWhatsMyIP handles the getWhatsMyIP tool call.
func (s *SDKServer) handleGetWhatsMyIP(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Try to extract client IP from HTTP request context
	if httpReq, ok := HTTPRequestFromContext(ctx); ok {
		clientIP := whatsmyip.ExtractClientIP(httpReq)
		slog.Debug("extracted client IP from HTTP request", "client_ip", clientIP, "remote_addr", httpReq.RemoteAddr)

		result, err := whatsmyip.GetWhatsMyIPWithClientIP(ctx, clientIP)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: formatToolError(err)}},
				IsError: true,
			}, nil
		}
		return createToolResultFromJSON(result)
	}

	// Fallback to server's IP
	result, err := whatsmyip.GetWhatsMyIP(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatToolError(err)}},
			IsError: true,
		}, nil
	}
	return createToolResultFromJSON(result)
}

// formatToolError formats an error for tool output.
func formatToolError(err error) string {
	errStr := err.Error()
	if len(errStr) > 6 && errStr[:6] == "Error:" {
		return errStr
	}
	return fmt.Sprintf("Error: %v", err)
}

// createToolResultFromJSON creates a CallToolResult from a JSON-serializable value.
func createToolResultFromJSON(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
	}, nil
}

// NewStreamableHTTPHandler creates an HTTP handler for the MCP server with
// custom middleware for CORS, origin validation, and HTTP context passing.
func (s *SDKServer) NewStreamableHTTPHandler() http.Handler {
	sdkHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, &mcp.StreamableHTTPOptions{
		JSONResponse: true, // Return JSON instead of SSE for simpler client compatibility
	})

	return &httpHandler{
		sdkHandler: sdkHandler,
	}
}

// httpHandler wraps the SDK handler with custom middleware.
type httpHandler struct {
	sdkHandler http.Handler
}

// Allowed origins for CORS
var allowedOrigins = []string{
	"http://localhost",
	"https://localhost",
	"http://127.0.0.1",
	"https://127.0.0.1",
	"https://cursor.sh",
	"https://claude.ai",
}

// isValidOrigin checks if the origin is in the allowed list.
func isValidOrigin(origin string) bool {
	for _, allowed := range allowedOrigins {
		if len(origin) >= len(allowed) && origin[:len(allowed)] == allowed {
			return true
		}
	}
	return false
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	protocolVersion := r.Header.Get("MCP-Protocol-Version")

	// Default to the only supported protocol version
	if protocolVersion == "" {
		protocolVersion = "2025-06-18"
	}

	// Only support 2025-06-18 protocol version
	if protocolVersion != "2025-06-18" {
		slog.Warn("unsupported protocol version", "version", protocolVersion)
		http.Error(w, "Unsupported protocol version. Only 2025-06-18 is supported.", http.StatusBadRequest)
		return
	}

	// Set protocol version header
	w.Header().Set("MCP-Protocol-Version", protocolVersion)

	// Handle OPTIONS for CORS preflight
	if r.Method == http.MethodOptions {
		h.handleCORS(w, r)
		return
	}

	// Validate origin for requests with Origin header
	if origin != "" {
		if !isValidOrigin(origin) {
			slog.Warn("invalid origin rejected", "origin", origin)
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Handle GET without method parameter (endpoint info)
	if r.Method == http.MethodGet && r.URL.Query().Get("method") == "" {
		h.handleEndpointInfo(w, r)
		return
	}

	// Inject HTTP request into context for getWhatsMyIP tool
	ctx := WithHTTPRequest(r.Context(), r)
	r = r.WithContext(ctx)

	// Delegate to SDK handler
	h.sdkHandler.ServeHTTP(w, r)
}

func (h *httpHandler) handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && isValidOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, MCP-Protocol-Version, MCP-Session-ID, Accept")
	w.Header().Set("Access-Control-Max-Age", "86400")

	w.WriteHeader(http.StatusOK)
}

func (h *httpHandler) handleEndpointInfo(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":     "mcp-ripestat",
		"protocol":    "MCP",
		"version":     "2025-06-18",
		"methods":     []string{"POST", "GET", "DELETE"},
		"description": "RIPEstat Data API MCP Server",
		"endpoints": map[string]interface{}{
			"mcp": map[string]interface{}{
				"url":         "/mcp",
				"methods":     []string{"POST", "GET", "DELETE"},
				"description": "Main MCP JSON-RPC endpoint",
				"usage": map[string]string{
					"POST":   "Send JSON-RPC 2.0 requests",
					"GET":    "Open SSE stream or use query parameters: ?method=<method>&params=<json>",
					"DELETE": "Terminate session",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write endpoint info response", "err", err)
	}
}

