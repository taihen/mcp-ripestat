package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

// mcpHandler handles MCP JSON-RPC requests.
func mcpHandler(w http.ResponseWriter, r *http.Request, server *mcp.Server) {
	slog.Debug("received MCP request", "method", r.Method, "remote_addr", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "err", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Extended timeout for cold start scenarios
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Store HTTP request in context for tools that need access to headers
	ctx = mcp.WithHTTPRequest(ctx, r)

	response, err := server.ProcessMessage(ctx, body)
	if err != nil {
		slog.Error("failed to process MCP message", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If no response (notification), return 204 No Content
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
