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
	"strconv"
	"syscall"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/abusecontactfinder"
	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asnneighbours"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkivalidation"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whois"
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
	mux.HandleFunc("/whois", whoisHandler)
	mux.HandleFunc("/abuse-contact-finder", abuseContactFinderHandler)
	mux.HandleFunc("/rpki-validation", rpkiValidationHandler)
	mux.HandleFunc("/asn-neighbours", asnNeighboursHandler)
	mux.HandleFunc("/.well-known/mcp/manifest.json", manifestHandler)

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
			{
				Name:        "getWhois",
				Description: "Get whois information for an IP address, prefix, or ASN.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The IP address, prefix, or ASN to query.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getAbuseContactFinder",
				Description: "Get abuse contact information for an IP address or prefix.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The IP address or prefix to query for abuse contacts.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getRPKIValidation",
				Description: "Get RPKI validation status for a resource (ASN) and prefix combination.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The ASN to validate against the prefix.",
					},
					{
						Name:        "prefix",
						Type:        "string",
						Required:    true,
						Description: "The IP prefix to validate.",
					},
				},
				Returns: Return{
					Type: "object",
				},
			},
			{
				Name:        "getASNNeighbours",
				Description: "Get ASN neighbours for an Autonomous System. Left neighbours are downstream providers, right neighbours are upstream providers.",
				Parameters: []Parameter{
					{
						Name:        "resource",
						Type:        "string",
						Required:    true,
						Description: "The AS number to query for neighbours.",
					},
					{
						Name:        "lod",
						Type:        "string",
						Required:    false,
						Description: "Level of detail: 0 (basic) or 1 (detailed with power, v4_peers, v6_peers). Default is 0.",
					},
					{
						Name:        "query_time",
						Type:        "string",
						Required:    false,
						Description: "Query time in ISO8601 format for historical data. If omitted, uses latest snapshot.",
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
		return asoverview.GetASOverview(ctx, resource)
	})
}

func announcedPrefixesHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "announced-prefixes", func(ctx context.Context, resource string) (interface{}, error) {
		return announcedprefixes.GetAnnouncedPrefixes(ctx, resource)
	})
}

func routingStatusHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "routing-status", func(ctx context.Context, resource string) (interface{}, error) {
		return routingstatus.GetRoutingStatus(ctx, resource)
	})
}

func whoisHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "whois", func(ctx context.Context, resource string) (interface{}, error) {
		return whois.GetWhois(ctx, resource)
	})
}

func abuseContactFinderHandler(w http.ResponseWriter, r *http.Request) {
	handleRIPEstatRequest(w, r, "abuse-contact-finder", func(ctx context.Context, resource string) (interface{}, error) {
		return abusecontactfinder.GetAbuseContactFinder(ctx, resource)
	})
}

func rpkiValidationHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("received rpki-validation request", "remote_addr", r.RemoteAddr, "query", r.URL.RawQuery)

	resource := r.URL.Query().Get("resource")
	prefix := r.URL.Query().Get("prefix")

	if resource == "" {
		slog.Warn("missing resource parameter", "call_name", "rpki-validation")
		writeJSONError(w, "missing resource parameter", http.StatusBadRequest)
		return
	}

	if prefix == "" {
		slog.Warn("missing prefix parameter", "call_name", "rpki-validation")
		writeJSONError(w, "missing prefix parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := rpkivalidation.GetRPKIValidation(ctx, resource, prefix)
	if err != nil {
		slog.Error("RIPEstat call failed", "call_name", "rpki-validation", "err", err)
		writeJSONError(w, "failed to fetch rpki-validation", http.StatusBadGateway)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func asnNeighboursHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("received asn-neighbours request", "remote_addr", r.RemoteAddr, "query", r.URL.RawQuery)

	resource := r.URL.Query().Get("resource")
	lodStr := r.URL.Query().Get("lod")
	queryTime := r.URL.Query().Get("query_time")

	if resource == "" {
		slog.Warn("missing resource parameter", "call_name", "asn-neighbours")
		writeJSONError(w, "missing resource parameter", http.StatusBadRequest)
		return
	}

	lod := 0 // default value
	if lodStr != "" {
		var err error
		lod, err = strconv.Atoi(lodStr)
		if err != nil || (lod != 0 && lod != 1) {
			slog.Warn("invalid lod parameter", "call_name", "asn-neighbours", "lod", lodStr)
			writeJSONError(w, "lod parameter must be 0 or 1", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := asnneighbours.GetASNNeighbours(ctx, resource, lod, queryTime)
	if err != nil {
		slog.Error("RIPEstat call failed", "call_name", "asn-neighbours", "err", err)
		writeJSONError(w, "failed to fetch asn-neighbours", http.StatusBadGateway)
		return
	}

	writeJSON(w, resp, http.StatusOK)
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

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		slog.Error("failed to write json response", "err", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSON(w, map[string]string{"error": message}, statusCode)
}
