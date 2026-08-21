package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/taihen/mcp-ripestat/internal/mcp/consolidated"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whatsmyip"
)

// SDKServer wraps the official MCP SDK server with RIPEstat-specific functionality.
type SDKServer struct {
	mcpServer         *mcp.Server
	consolidatedTools *consolidated.Tools
	disableWhatsMyIP  bool
	rateLimiter       *RateLimiter
	allowLegacy       bool
	allowedVersions   []string
}

// NewSDKServer creates a new MCP server using the official SDK.
func NewSDKServer(serverName, serverVersion string, disableWhatsMyIP bool) *SDKServer {
	allowLegacy := legacyProtocolsEnabled()
	allowedVersions := allowedProtocolVersions(allowLegacy)
	impl := &mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}

	mcpServer := mcp.NewServer(impl, &mcp.ServerOptions{
		Logger:       slog.Default(),
		Capabilities: &mcp.ServerCapabilities{},
	})
	mcpServer.AddReceivingMiddleware(protocolResultMiddleware(ToolsListTTLMs, allowedVersions))

	executor := consolidated.NewDirectExecutor()
	consolidatedTools := consolidated.NewTools(executor)

	s := &SDKServer{
		mcpServer:         mcpServer,
		consolidatedTools: consolidatedTools,
		disableWhatsMyIP:  disableWhatsMyIP,
		rateLimiter:       NewRateLimiter(DefaultRateLimitConfig()),
		allowLegacy:       allowLegacy,
		allowedVersions:   allowedVersions,
	}

	s.registerTools()

	return s
}

// protocolResultMiddleware applies deployment policy to discovery and marks
// caller-independent discovery/tool-list results as publicly cacheable.
func protocolResultMiddleware(ttlMs int, allowedVersions []string) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil {
				return result, err
			}
			switch cacheable := result.(type) {
			case *mcp.DiscoverResult:
				cacheable.TTLMs = ttlMs
				cacheable.CacheScope = CacheScopePublic
				cacheable.SupportedVersions = append([]string(nil), allowedVersions...)
			case *mcp.ListToolsResult:
				cacheable.TTLMs = ttlMs
				cacheable.CacheScope = CacheScopePublic
			}
			return result, nil
		}
	}
}

func allowedProtocolVersions(allowLegacy bool) []string {
	versions := []string{ProtocolVersion}
	if allowLegacy {
		versions = append(versions, "2025-11-25", "2025-06-18", "2025-03-26")
	}
	return versions
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
			withPanicRecovery(s.handleGetWhatsMyIP, "getWhatsMyIP"),
		)
	}
}

// withPanicRecovery wraps a tool handler with panic recovery to prevent server crashes.
func withPanicRecovery(handler mcp.ToolHandler, toolName string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (callResult *mcp.CallToolResult, callErr error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in tool handler", "tool", toolName, "panic", r)
				callResult = &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: "Error: internal error occurred"}},
					IsError: true,
				}
				callErr = nil
			}
		}()
		return handler(ctx, req)
	}
}

// createConsolidatedToolHandler creates a handler for consolidated tools.
func (s *SDKServer) createConsolidatedToolHandler(toolName string) mcp.ToolHandler {
	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		return createToolResultFromJSON(result), nil
	}
	return withPanicRecovery(handler, toolName)
}

// handleGetWhatsMyIP handles the getWhatsMyIP tool call.
func (s *SDKServer) handleGetWhatsMyIP(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		return createToolResultFromJSON(result), nil
	}

	// Fallback to server's IP
	result, err := whatsmyip.GetWhatsMyIP(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatToolError(err)}},
			IsError: true,
		}, nil
	}
	return createToolResultFromJSON(result), nil
}

// formatToolError formats an error for tool output.
func formatToolError(err error) string {
	errStr := err.Error()
	if strings.HasPrefix(errStr, "Error: ") {
		return errStr
	}
	return fmt.Sprintf("Error: %v", err)
}

// createToolResultFromJSON creates a CallToolResult from a JSON-serializable value.
func createToolResultFromJSON(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)}},
			IsError: true,
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
	}
}

// NewStreamableHTTPHandler creates an HTTP handler for the MCP server with
// custom middleware for CORS, origin validation, and HTTP context passing.
func (s *SDKServer) NewStreamableHTTPHandler() http.Handler {
	sdkHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return s.mcpServer
	}, &mcp.StreamableHTTPOptions{
		Stateless:                    true,
		JSONResponse:                 true, // Return JSON when no request-scoped notifications need streaming.
		PropagateRequestCancellation: true,
	})

	return &httpHandler{
		sdkHandler:      sdkHandler,
		rateLimiter:     s.rateLimiter,
		allowLegacy:     s.allowLegacy,
		allowedVersions: append([]string(nil), s.allowedVersions...),
	}
}

// httpHandler wraps the SDK handler with custom middleware.
type httpHandler struct {
	sdkHandler      http.Handler
	rateLimiter     *RateLimiter
	allowLegacy     bool
	allowedVersions []string
}

// Default allowed origins for CORS.
var defaultAllowedOrigins = []string{
	"http://localhost",
	"https://localhost",
	"http://127.0.0.1",
	"https://127.0.0.1",
}

var (
	allowedOrigins     []string
	allowedOriginsOnce sync.Once
)

// getAllowedOrigins returns the list of allowed CORS origins.
// It reads from the CORS_ALLOWED_ORIGINS environment variable (comma-separated)
// and merges with the default origins.
func getAllowedOrigins() []string {
	allowedOriginsOnce.Do(func() {
		// Start with defaults
		allowedOrigins = make([]string, len(defaultAllowedOrigins))
		copy(allowedOrigins, defaultAllowedOrigins)

		// Add any origins from environment variable
		if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
			for _, origin := range strings.Split(envOrigins, ",") {
				origin = strings.TrimSpace(origin)
				if origin != "" {
					allowedOrigins = append(allowedOrigins, origin)
				}
			}
		}
	})
	return allowedOrigins
}

// isValidOrigin checks if the origin is in the allowed list.
func isValidOrigin(origin string) bool {
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}

	originScheme := strings.ToLower(originURL.Scheme)
	originHost := strings.ToLower(originURL.Hostname())
	originPort := originURL.Port()

	for _, allowed := range getAllowedOrigins() {
		allowedURL, parseErr := url.Parse(allowed)
		if parseErr != nil || allowedURL.Scheme == "" || allowedURL.Host == "" {
			continue
		}

		allowedScheme := strings.ToLower(allowedURL.Scheme)
		allowedHost := strings.ToLower(allowedURL.Hostname())
		allowedPort := allowedURL.Port()

		if originScheme != allowedScheme || originHost != allowedHost {
			continue
		}

		// Keep local development ergonomics: allow any port for loopback entries
		// defined without explicit ports (e.g. http://localhost).
		if allowedPort == "" && isLoopbackHost(allowedHost) {
			return true
		}

		if effectivePort(originScheme, originPort) == effectivePort(allowedScheme, allowedPort) {
			return true
		}
	}
	return false
}

func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func effectivePort(scheme, port string) string {
	if port != "" {
		return port
	}

	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}

func sanitizeForLog(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		default:
			return r
		}
	}, value)
}

func legacyProtocolsEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("MCP_ENABLE_LEGACY_PROTOCOLS")))
	return value == "1" || value == "true"
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	w.Header().Add("Vary", "Origin")
	if origin != "" && !isValidOrigin(origin) {
		//nolint:gosec // value is sanitized before structured logging.
		slog.Warn("invalid origin rejected", "origin", sanitizeForLog(origin))
		http.Error(w, "Invalid origin", http.StatusForbidden)
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Handle OPTIONS for CORS preflight
	if r.Method == http.MethodOptions {
		h.handleCORS(w, r)
		return
	}

	// Apply rate limiting only to protocol traffic, not browser preflights.
	if h.rateLimiter != nil {
		clientIP := extractClientIPForRateLimit(r)
		if !h.rateLimiter.Allow(clientIP) {
			//nolint:gosec // value is sanitized before structured logging.
			slog.Warn("rate limit exceeded", "client_ip", sanitizeForLog(clientIP))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
	}

	if r.Method == http.MethodPost && !h.enforceProtocolPolicy(w, r) {
		return
	}

	// Inject HTTP request into context for getWhatsMyIP tool
	ctx := WithHTTPRequest(r.Context(), r)
	r = r.WithContext(ctx)

	// Delegate to SDK handler
	h.sdkHandler.ServeHTTP(w, r)
}

const maxPolicyBodyBytes = 4 << 20

type policyEnvelope struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params struct {
		ProtocolVersion string `json:"protocolVersion"`
	} `json:"params"`
}

func (h *httpHandler) enforceProtocolPolicy(w http.ResponseWriter, r *http.Request) bool {
	requested := r.Header.Get("MCP-Protocol-Version")
	if requested == ProtocolVersion {
		return true
	}

	envelopes, ok := readPolicyEnvelopes(w, r)
	if !ok {
		return false
	}
	firstID := firstPolicyID(envelopes)
	if h.allowLegacy && h.supportsVersion(requested) {
		for _, envelope := range envelopes {
			if envelope.Method != "initialize" {
				continue
			}
			bodyVersion := envelope.Params.ProtocolVersion
			if bodyVersion == ProtocolVersion || !h.supportsVersion(bodyVersion) {
				h.rejectProtocolRequest(w, bodyVersion, envelope.ID)
				return false
			}
			if bodyVersion != requested {
				h.rejectLegacyHeaderMismatch(w, requested, bodyVersion, envelope.ID)
				return false
			}
		}
		return true
	}
	if requested == "" && h.allowLegacy {
		// 2025-03-26 predates MCP-Protocol-Version. Headerless non-initialize
		// requests are necessarily handled as that compatibility revision.
		for _, envelope := range envelopes {
			if envelope.Method != "initialize" {
				continue
			}
			bodyVersion := envelope.Params.ProtocolVersion
			if bodyVersion == ProtocolVersion || !h.supportsVersion(bodyVersion) {
				h.rejectProtocolRequest(w, bodyVersion, envelope.ID)
				return false
			}
		}
		return true
	}

	h.rejectProtocolRequest(w, requested, firstID)
	return false
}

func (h *httpHandler) rejectLegacyHeaderMismatch(w http.ResponseWriter, headerVersion, bodyVersion string, id json.RawMessage) {
	response := NewErrorResponse(HeaderMismatchError, "MCP-Protocol-Version header does not match initialize protocolVersion", map[string]interface{}{
		"header": headerVersion,
		"body":   bodyVersion,
	}, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write protocol mismatch response", "err", err)
	}
}

func (h *httpHandler) supportsVersion(version string) bool {
	return slices.Contains(h.allowedVersions, version)
}

func readPolicyEnvelopes(w http.ResponseWriter, r *http.Request) ([]policyEnvelope, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxPolicyBodyBytes+1))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return nil, false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) > maxPolicyBodyBytes {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return nil, false
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var envelopes []policyEnvelope
		_ = json.Unmarshal(trimmed, &envelopes)
		return envelopes, true
	}
	var envelope policyEnvelope
	_ = json.Unmarshal(trimmed, &envelope)
	return []policyEnvelope{envelope}, true
}

func firstPolicyID(envelopes []policyEnvelope) json.RawMessage {
	if len(envelopes) == 0 {
		return nil
	}
	return envelopes[0].ID
}

func (h *httpHandler) rejectProtocolRequest(w http.ResponseWriter, requested string, id json.RawMessage) {
	code := HeaderMismatchError
	message := "MCP-Protocol-Version header is required"
	if requested != "" {
		code = UnsupportedProtocolVersionError
		message = "Unsupported protocol version"
	}
	response := NewErrorResponse(code, message, map[string]interface{}{
		"supported": append([]string(nil), h.allowedVersions...),
		"requested": requested,
		"hint":      "Set MCP_ENABLE_LEGACY_PROTOCOLS=true only when legacy clients are explicitly required.",
	}, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write protocol policy response", "err", err)
	}
}

func (h *httpHandler) handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && isValidOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, MCP-Protocol-Version, Mcp-Method, Mcp-Name, Mcp-Session-Id")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")

	w.WriteHeader(http.StatusOK)
}
