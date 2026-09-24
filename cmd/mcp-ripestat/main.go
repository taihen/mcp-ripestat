package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taihen/mcp-ripestat/internal/mcp"
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

	// Create the SDK-based MCP server
	mcpServer := mcp.NewSDKServer("mcp-ripestat", version, false)

	addr := ":" + port

	server := &http.Server{
		Addr:              addr,
		Handler:           newHTTPHandler(startTime, mcpServer.NewStreamableHTTPHandler()),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	serveErrCh := make(chan error, 1)

	go func() {
		slog.Info("MCP RIPEstat server starting", "addr", server.Addr)
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "err", err)
			select {
			case serveErrCh <- err:
			default:
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serveErrCh:
		return fmt.Errorf("server failed: %w", err)
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

// newHTTPHandler exposes only the public service paths. Operational/debug
// endpoints stay out of the public router, and unknown paths are rejected
// before reaching application handlers.
func newHTTPHandler(startTime time.Time, mcpHandler http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/mcp", mcpHandler)
	mux.HandleFunc("/.well-known/mcp/manifest.json", manifestHandler)
	mux.HandleFunc("/warmup", warmupHandler)
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, startTime)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isPublicPath(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func isPublicPath(path string) bool {
	switch path {
	case "/mcp", "/.well-known/mcp/manifest.json", "/warmup", "/status":
		return true
	default:
		return false
	}
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
	//nolint:gosec // debug logging only; request metadata is not used for decisions.
	slog.Debug("received manifest request", "remote_addr", r.RemoteAddr)

	var functions []Function

	manifest := Manifest{
		Name:        "mcp-ripestat",
		Description: "A server for the RIPEstat Data API, providing network information for IP addresses and prefixes.",
		Functions:   functions,
	}
	writeJSON(w, manifest)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		slog.Error("failed to write json response", "err", err)
	}
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
