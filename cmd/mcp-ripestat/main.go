package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

// version is set via -ldflags during build time.
var version = "dev"

func main() {
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	showVersion := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Print all possible flags")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("mcp-ripestat version %s\n", version)
		os.Exit(0)
	}

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	slog.SetDefault(logger)

	if err := run(context.Background(), *port); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, port string) error {
	startTime := time.Now()
	mux := http.NewServeMux()

	// Create MCP server
	mcpServer := mcp.NewServer("mcp-ripestat", version, false)

	// Add MCP JSON-RPC endpoint
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		mcpHandler(w, r, mcpServer)
	})

	mux.HandleFunc("/.well-known/mcp/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		manifestHandler(w, r)
	})

	// Warmup endpoint to prevent cold starts
	mux.HandleFunc("/warmup", warmupHandler)

	// Status endpoint for debugging cold starts
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, startTime)
	})

	addr := ":" + port

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	go func() {
		slog.Info("MCP RIPEstat server starting", "addr", server.Addr)
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "err", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("shutting down server...")
	case <-ctx.Done():
		slog.Info("shutting down server due to context cancellation...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	slog.Info("server exited gracefully")

	return nil
}

// Manifest represents the structure of the .well-known/mcp/manifest.json file.
type Manifest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Functions   []Function `json:"functions"`
}

// Function represents a single function in the manifest.
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	Returns     Return      `json:"returns"`
}

// Parameter represents a single parameter for a function.
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Return represents the return type of a function.
type Return struct {
	Type string `json:"type"`
}

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("received manifest request", "remote_addr", r.RemoteAddr)

	var functions []Function

	manifest := Manifest{
		Name:        "mcp-ripestat",
		Description: "A server for the RIPEstat Data API, providing network information for IP addresses and prefixes.",
		Functions:   functions,
	}
	writeJSON(w, manifest, http.StatusOK)
}

// mcpHandler handles MCP JSON-RPC requests with streamable HTTP support.
func mcpHandler(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	origin := r.Header.Get("Origin")
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	slog.Debug("received MCP request", "method", r.Method, "remote_addr", r.RemoteAddr, "origin", origin, "protocol_version", protocolVersion)

	// Determine if this is a streamable HTTP request
	// Only treat as streamable if:
	// 1. It's a GET request (explicit streamable call)
	// 2. It's an OPTIONS request (CORS preflight)
	// 3. It's a POST request with MCP-Protocol-Version >= 2025-06-18 and Origin header
	isStreamableHTTP := false

	switch r.Method {
	case http.MethodGet:
		// GET requests are always streamable HTTP
		isStreamableHTTP = true
	case http.MethodOptions:
		// OPTIONS requests are for CORS
		isStreamableHTTP = true
	case http.MethodPost:
		// POST with Origin header and new protocol version (2025-06-18+)
		if origin != "" {
			// Check if protocol version supports streamable HTTP
			// Be strict: only 2025-06-18+ or empty (default to latest) should get streamable HTTP
			supportsStreamableHTTP := protocolVersion == "" || protocolVersion == "2025-06-18"
			isStreamableHTTP = supportsStreamableHTTP
			slog.Debug("protocol version check", "version", protocolVersion, "supports_streamable", supportsStreamableHTTP)
		}
	}

	slog.Debug("request classification", "is_streamable", isStreamableHTTP, "method", r.Method, "has_origin", origin != "", "protocol_version", protocolVersion)

	if isStreamableHTTP {
		slog.Debug("processing as streamable HTTP request")
		// Validate streamable HTTP requirements
		if !validateStreamableHTTP(w, r) {
			return
		}

		// Handle session management
		sessionID := getOrCreateSession(r, w)
		slog.Debug("session management", "session_id", sessionID)

		// Route based on HTTP method
		switch r.Method {
		case http.MethodPost:
			slog.Debug("routing to handleMCPRequest for streamable POST")
			handleMCPRequest(w, r, server, sessionID)
		case http.MethodGet:
			slog.Debug("routing to handleMCPQuery for GET")
			handleMCPQuery(w, r, server, sessionID)
		case http.MethodOptions:
			slog.Debug("routing to handleCORS for OPTIONS")
			handleCORS(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		slog.Debug("processing as regular MCP client")
		// Handle regular MCP clients (POST without streamable support)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleMCPRequest(w, r, server, "")
	}
}

// validateStreamableHTTP validates HTTP transport requirements.
func validateStreamableHTTP(w http.ResponseWriter, r *http.Request) bool {
	// Origin validation (required by MCP spec).
	if origin := r.Header.Get("Origin"); origin != "" {
		if !isValidOrigin(origin) {
			slog.Warn("invalid origin rejected", "origin", origin)
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return false
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Protocol version handling.
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "2025-06-18"
	}
	if !isSupportedProtocolVersion(protocolVersion) {
		slog.Warn("unsupported protocol version", "version", protocolVersion)
		http.Error(w, "Unsupported protocol version", http.StatusBadRequest)
		return false
	}
	w.Header().Set("MCP-Protocol-Version", protocolVersion)

	return true
}

// isValidOrigin validates the origin header.
func isValidOrigin(origin string) bool {
	// Allow common development origins.
	allowedOrigins := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"https://cursor.sh",
		"https://claude.ai",
	}

	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(origin, allowed) {
			return true
		}
	}

	return false
}

// isSupportedProtocolVersion checks if protocol version is supported.
func isSupportedProtocolVersion(version string) bool {
	supportedVersions := []string{
		"2025-06-18",
		"2025-03-26", // Backward compatibility.
	}

	for _, supported := range supportedVersions {
		if version == supported {
			return true
		}
	}

	return false
}

// getOrCreateSession manages session IDs.
func getOrCreateSession(r *http.Request, w http.ResponseWriter) string {
	sessionID := r.Header.Get("MCP-Session-ID")
	if sessionID == "" {
		sessionID = generateSessionID()
		w.Header().Set("MCP-Session-ID", sessionID)
	}
	return sessionID
}

// generateSessionID creates a new session ID.
func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID.
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// handleMCPRequest handles POST requests (standard JSON-RPC).
func handleMCPRequest(w http.ResponseWriter, r *http.Request, server *mcp.Server, sessionID string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "err", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Extended timeout for cold start scenarios.
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Store HTTP request and session in context.
	ctx = mcp.WithHTTPRequest(ctx, r)
	ctx = mcp.WithSessionID(ctx, sessionID)

	response, err := server.ProcessMessage(ctx, body)
	if err != nil {
		slog.Error("failed to process MCP message", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If no response (notification), return 204 No Content.
	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write MCP response", "err", err)
	}
}

// handleMCPQuery handles GET requests (query parameters to JSON-RPC).
func handleMCPQuery(w http.ResponseWriter, r *http.Request, server *mcp.Server, sessionID string) {
	query := r.URL.Query()
	slog.Debug("handling GET request", "query_params", query, "has_method", query.Get("method") != "")

	// Check if this is a valid MCP query request (has method parameter)
	if query.Get("method") == "" {
		slog.Debug("GET request to MCP endpoint without method parameter, returning endpoint info", "query", query, "user_agent", r.Header.Get("User-Agent"))
		// Return basic endpoint information for health checks and discovery
		// This helps MCP clients and tooling understand the endpoint capabilities
		response := map[string]interface{}{
			"service":     "mcp-ripestat",
			"protocol":    "MCP",
			"version":     "2025-06-18",
			"methods":     []string{"POST", "GET"},
			"description": "RIPEstat Data API MCP Server",
			"endpoints": map[string]interface{}{
				"mcp": map[string]interface{}{
					"url":         "/mcp",
					"methods":     []string{"POST", "GET"},
					"description": "Main MCP JSON-RPC endpoint",
					"usage": map[string]string{
						"POST": "Send JSON-RPC 2.0 requests",
						"GET":  "Use query parameters: ?method=<method>&params=<json>",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to write endpoint info response", "err", err)
		}
		return
	}

	// Convert query parameters to JSON-RPC request.
	requestData, err := server.ParseQueryToRequest(query)
	if err != nil {
		slog.Error("failed to parse query parameters", "err", err)
		http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		return
	}

	// Extended timeout for cold start scenarios.
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Store HTTP request and session in context.
	ctx = mcp.WithHTTPRequest(ctx, r)
	ctx = mcp.WithSessionID(ctx, sessionID)

	response, err := server.ProcessMessage(ctx, requestData)
	if err != nil {
		slog.Error("failed to process MCP query", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If no response (notification), return 204 No Content.
	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write MCP query response", "err", err)
	}
}

// handleCORS handles CORS preflight requests.
func handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && isValidOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, MCP-Protocol-Version, MCP-Session-ID")
	w.Header().Set("Access-Control-Max-Age", "86400")

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		slog.Error("failed to write json response", "err", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}

func warmupHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"server":    "mcp-ripestat",
	}); err != nil {
		slog.Error("failed to encode warmup response", "err", err)
	}
}

func statusHandler(w http.ResponseWriter, _ *http.Request, startTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"server":    "mcp-ripestat",
		"version":   version,
		"mcp_ready": true,
		"uptime":    time.Since(startTime).String(),
	}); err != nil {
		slog.Error("failed to encode status response", "err", err)
	}
}
