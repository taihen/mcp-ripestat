package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
)

func main() {
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/network-info", networkInfoHandler)

	addr := ":" + *port

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		slog.Info("MCP RIPEstat server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "err", err)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("received network-info request", "remote_addr", r.RemoteAddr, "query", r.URL.RawQuery)
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		slog.Warn("missing resource parameter")
		writeJSONError(w, "missing resource parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := networkinfo.GetNetworkInfo(ctx, resource)
	if err != nil {
		slog.Error("network-info call failed", "err", err)
		http.Error(w, `{"error":"failed to fetch network info"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to write response", "err", err)
	}
}
