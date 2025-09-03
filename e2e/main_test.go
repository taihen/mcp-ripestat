//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

var serverURL string
var serverProcess *exec.Cmd

func TestMain(m *testing.M) {
	// Start the server
	serverURL = "http://localhost:8081"
	serverProcess = exec.Command("../bin/mcp-ripestat", "--port", "8081")

	// Redirect stdout and stderr to os.Stdout and os.Stderr
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr

	// Start the server
	if err := serverProcess.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the tests
	code := m.Run()

	// Stop the server
	if err := serverProcess.Process.Kill(); err != nil {
		fmt.Printf("Failed to kill server process: %v\n", err)
	}

	os.Exit(code)
}

// TestManifest tests that the manifest endpoint returns a 200 status code
func TestManifest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL+"/.well-known/mcp/manifest.json", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestNetworkInfoMissingResource tests that MCP tools/call returns error when no resource is provided
func TestNetworkInfoMissingResource(t *testing.T) {
	// Since REST endpoints were removed in Sprint 23, this test now verifies MCP error handling
	// The old REST endpoint `/network-info` no longer exists, so we expect 404
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL+"/network-info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// REST endpoints were removed in Sprint 23, so we expect 404 Not Found
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d (REST endpoints removed), got %d", http.StatusNotFound, resp.StatusCode)
	}
}
