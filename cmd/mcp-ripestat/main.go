package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"expvar"
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
	"github.com/taihen/mcp-ripestat/internal/ripestat/metrics"
)


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


	mcpServer := mcp.NewServer("mcp-ripestat", version, false)


	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		mcpHandler(w, r, mcpServer)
	})

	mux.HandleFunc("/.well-known/mcp/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		manifestHandler(w, r)
	})


	mux.HandleFunc("/warmup", warmupHandler)


	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, startTime)
	})


	mux.Handle("/debug/vars", expvar.Handler())
	mux.HandleFunc("/metrics", metricsHandler)

	addr := ":" + port

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("MCP RIPEstat server starting", "addr", server.Addr)
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "err", err)
		}
	}()


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


type Manifest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Functions   []Function `json:"functions"`
}


type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	Returns     Return      `json:"returns"`
}


type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}


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


func mcpHandler(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	origin := r.Header.Get("Origin")
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "2025-06-18"
	}

	slog.Debug("received MCP request", "method", r.Method, "remote_addr", r.RemoteAddr, "origin", origin, "protocol_version", protocolVersion)


	if !isSupportedProtocolVersion(protocolVersion) {
		slog.Warn("unsupported protocol version", "version", protocolVersion)
		http.Error(w, "Unsupported protocol version", http.StatusBadRequest)
		return
	}


	supportsStreamableHTTP := isProtocolVersionAtLeast(protocolVersion, "2025-06-18")


	if !supportsStreamableHTTP {
		slog.Debug("using simplified handling for older protocol version", "version", protocolVersion)
		handleLegacyMCPClient(w, r, server)
		return
	}


	isStreamableHTTP := false

	switch r.Method {
	case http.MethodGet:

		isStreamableHTTP = true
	case http.MethodOptions:

		isStreamableHTTP = true
	case http.MethodPost:

		if origin != "" {
			isStreamableHTTP = true
			slog.Debug("POST with origin detected as streamable HTTP", "origin", origin)
		}
	}

	slog.Debug("request classification", "is_streamable", isStreamableHTTP, "method", r.Method, "has_origin", origin != "", "protocol_version", protocolVersion)

	if isStreamableHTTP {
		slog.Debug("processing as streamable HTTP request")

		if !validateStreamableHTTP(w, r) {
			return
		}


		sessionID := getOrCreateSession(r, w)
		slog.Debug("session management", "session_id", sessionID)


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
		slog.Debug("processing as regular MCP client (POST-only)")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleMCPRequest(w, r, server, "")
	}
}


func validateStreamableHTTP(w http.ResponseWriter, r *http.Request) bool {

	if origin := r.Header.Get("Origin"); origin != "" {
		if !isValidOrigin(origin) {
			slog.Warn("invalid origin rejected", "origin", origin)
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return false
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}


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


func isValidOrigin(origin string) bool {

	allowedOrigins := []string{
		"http:
		"https:
		"http:
		"https:
		"https:
		"https:
	}

	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(origin, allowed) {
			return true
		}
	}

	return false
}


func isSupportedProtocolVersion(version string) bool {
	supportedVersions := []string{
		"2025-06-18",
		"2025-03-26",
	}

	for _, supported := range supportedVersions {
		if version == supported {
			return true
		}
	}

	return false
}


func getOrCreateSession(r *http.Request, w http.ResponseWriter) string {
	sessionID := r.Header.Get("MCP-Session-ID")
	if sessionID == "" {
		sessionID = generateSessionID()
		w.Header().Set("MCP-Session-ID", sessionID)
	}
	return sessionID
}


func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {

		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}


func handleMCPRequest(w http.ResponseWriter, r *http.Request, server *mcp.Server, sessionID string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "err", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()


	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()


	ctx = mcp.WithHTTPRequest(ctx, r)
	ctx = mcp.WithSessionID(ctx, sessionID)

	response, err := server.ProcessMessage(ctx, body)
	if err != nil {
		slog.Error("failed to process MCP message", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}


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


func handleMCPQuery(w http.ResponseWriter, r *http.Request, server *mcp.Server, sessionID string) {
	query := r.URL.Query()
	slog.Debug("handling GET request", "query_params", query, "has_method", query.Get("method") != "")


	if query.Get("method") == "" {
		slog.Debug("GET request to MCP endpoint without method parameter, returning endpoint info", "query", query, "user_agent", r.Header.Get("User-Agent"))


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


	requestData, err := server.ParseQueryToRequest(query)
	if err != nil {
		slog.Error("failed to parse query parameters", "err", err)
		http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		return
	}


	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()


	ctx = mcp.WithHTTPRequest(ctx, r)
	ctx = mcp.WithSessionID(ctx, sessionID)

	response, err := server.ProcessMessage(ctx, requestData)
	if err != nil {
		slog.Error("failed to process MCP query", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}


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


func isProtocolVersionAtLeast(version, minVersion string) bool {

	return version >= minVersion
}


func handleLegacyMCPClient(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	origin := r.Header.Get("Origin")
	protocolVersion := r.Header.Get("MCP-Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "2025-03-26"
	}


	w.Header().Set("MCP-Protocol-Version", protocolVersion)


	if origin != "" && isValidOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, MCP-Protocol-Version")
	}


	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}


	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	handleSimpleMCPRequest(w, r, server)
}


func handleSimpleMCPRequest(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "err", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()


	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()


	ctx = mcp.WithHTTPRequest(ctx, r)

	response, err := server.ProcessMessage(ctx, body)
	if err != nil {
		slog.Error("failed to process MCP message", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}


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

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	summary := metrics.Summary()
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		slog.Error("failed to encode metrics response", "err", err)
	}
}
