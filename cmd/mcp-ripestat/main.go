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

	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
)

func main() {
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
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
	mux := http.NewServeMux()
	mux.HandleFunc("/network-info", networkInfoHandler)
	mux.HandleFunc("/as-overview", asOverviewHandler)
	mux.HandleFunc("/announced-prefixes", announcedPrefixesHandler)
	mux.HandleFunc("/routing-status", routingStatusHandler)
	mux.HandleFunc("/.well-known/mcp/manifest.json", manifestHandler)

	addr := ":" + port

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		slog.Info("MCP RIPEstat server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	manifest := Manifest{
		Name:        "mcp-ripestat",
		Description: "A server for the RIPEstat Data API, providing network information for IP addresses and prefixes.",
		Functions: []Function{
			{
				Name:        "getNetworkInfo",
				Description: "Get network information for an IP address or prefix.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The IP address or prefix to query.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getASOverview",
				Description: "Get an overview of an Autonomous System (AS).",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The AS number to query.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getAnnouncedPrefixes",
				Description: "Get a list of prefixes announced by an Autonomous System (AS).",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The AS number to query.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getRoutingStatus",
				Description: "Get the routing status for an IP prefix.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The IP prefix to query.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
		},
	}
	writeJSON(w, manifest, http.StatusOK)
}

func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "network-info", func(ctx context.Context, resource string) (interface{}, error) {
		return networkinfo.GetNetworkInfo(ctx, resource)
	})
}

func asOverviewHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "as-overview", func(ctx context.Context, resource string) (interface{}, error) {
		return asoverview.Get(ctx, resource)
	})
}

func announcedPrefixesHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "announced-prefixes", func(ctx context.Context, resource string) (interface{}, error) {
		return announcedprefixes.Get(ctx, resource)
	})
}

func routingStatusHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "routing-status", func(ctx context.Context, resource string) (interface{}, error) {
		client := routingstatus.NewClient("https://stat.ripe.net", http.DefaultClient)
		return client.Get(ctx, resource)
	})
}

type ripeStatFunc func(ctx context.Context, resource string) (interface{}, error)

func handleRIPEstatRequest(w http.ResponseWriter, r *http.Request, callName string, fn ripeStatFunc) {
	slog.Debug("received request", "call_name", callName, "remote_addr", r.RemoteAddr, "query", r.URL.RawQuery)
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		slog.Warn("missing resource parameter", "call_name", callName)
		writeJSONError(w, `missing resource parameter`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := fn(ctx, resource)
	if err != nil {
		slog.Error("RIPEstat call failed", "call_name", callName, "err", err)
		writeJSONError(w, fmt.Sprintf("failed to fetch %s", callName), http.StatusBadGateway)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "err", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}
