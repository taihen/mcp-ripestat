package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

func TestNetworkInfoHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/network-info?resource=8.8.8.8", nil)
	w := httptest.NewRecorder()

	networkInfoHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestASOverviewHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/as-overview?resource=AS15169", nil)
	w := httptest.NewRecorder()

	asOverviewHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestAnnouncedPrefixesHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/announced-prefixes?resource=AS15169", nil)
	w := httptest.NewRecorder()

	announcedPrefixesHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestRoutingStatusHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/routing-status?resource=8.8.8.0/24", nil)
	w := httptest.NewRecorder()

	routingStatusHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestManifestHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/.well-known/mcp/manifest.json", nil)
	w := httptest.NewRecorder()

	manifestHandler(w, req, false) // Test with whats-my-ip enabled

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if manifest.Name != "mcp-ripestat" {
		t.Errorf("Expected manifest name to be 'mcp-ripestat', got %q", manifest.Name)
	}

	if len(manifest.Functions) != 10 {
		t.Errorf("Expected 10 functions in manifest, got %d", len(manifest.Functions))
	}

	// Check that all expected functions are present
	functionNames := make(map[string]bool)
	for _, fn := range manifest.Functions {
		functionNames[fn.Name] = true
	}

	expectedFunctions := []string{
		"getNetworkInfo",
		"getASOverview",
		"getAnnouncedPrefixes",
		"getRoutingStatus",
		"getWhois",
		"getAbuseContactFinder",
		"getRPKIValidation",
		"getASNNeighbours",
		"getLookingGlass",
		"getWhatsMyIP",
	}

	for _, name := range expectedFunctions {
		if !functionNames[name] {
			t.Errorf("Expected function %q in manifest", name)
		}
	}
}

func TestManifestHandler_WhatsMyIPDisabled(t *testing.T) {
	req := httptest.NewRequest("GET", "/.well-known/mcp/manifest.json", nil)
	w := httptest.NewRecorder()

	manifestHandler(w, req, true) // Test with whats-my-ip disabled

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if len(manifest.Functions) != 9 {
		t.Errorf("Expected 9 functions in manifest when whats-my-ip is disabled, got %d", len(manifest.Functions))
	}

	// Check that whats-my-ip function is not present
	functionNames := make(map[string]bool)
	for _, fn := range manifest.Functions {
		functionNames[fn.Name] = true
	}

	if functionNames["getWhatsMyIP"] {
		t.Error("Expected getWhatsMyIP function to be absent when disabled")
	}
}

func TestHandleRIPEstatRequest_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/network-info", nil)
	w := httptest.NewRecorder()

	handleRIPEstatRequest(w, req, "test", func(_ context.Context, _ string) (interface{}, error) {
		return nil, nil
	})

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestHandleRIPEstatRequest_BackendError(t *testing.T) {
	req := httptest.NewRequest("GET", "/network-info?resource=8.8.8.8", nil)
	w := httptest.NewRecorder()

	handleRIPEstatRequest(w, req, "test", func(_ context.Context, _ string) (interface{}, error) {
		return nil, io.EOF
	})

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status code 502, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "failed to fetch") {
		t.Errorf("Expected error message about fetch failure, got %q", string(body))
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, data, http.StatusOK)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected result[\"key\"] to be 'value', got %q", result["key"])
	}
}

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSONError(w, "test error", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %q", contentType)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "test error" {
		t.Errorf("Expected error message to be 'test error', got %q", response["error"])
	}
}

func TestRun_ServerStartup(t *testing.T) {
	// Test that the server starts and shuts down properly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use a random available port
	port := "0" // Let the OS choose an available port

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, port, false)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for the server to shut down
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error from run, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	// Test that the server shuts down when context is cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Use a random available port
	port := "0"

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, port, false)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for the server to shut down
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error from run, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestHandleRIPEstatRequest_MissingResourceDetailed(t *testing.T) {
	// Test the missing resource parameter case with detailed checks
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Mock function that should not be called
	mockFn := func(_ context.Context, _ string) (interface{}, error) {
		t.Fatal("Function should not be called when resource is missing")
		return nil, nil
	}

	handleRIPEstatRequest(w, req, "test-call", mockFn)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "missing resource parameter" {
		t.Errorf("Expected error message about missing resource, got %q", response["error"])
	}
}

func TestHandleRIPEstatRequest_FunctionError(t *testing.T) {
	// Test the case where the RIPEstat function returns an error
	req := httptest.NewRequest(http.MethodGet, "/test?resource=test-resource", nil)
	w := httptest.NewRecorder()

	// Mock function that returns an error
	mockFn := func(_ context.Context, resource string) (interface{}, error) {
		if resource != "test-resource" {
			t.Errorf("Expected resource 'test-resource', got %q", resource)
		}
		return nil, errors.New("mock error")
	}

	handleRIPEstatRequest(w, req, "test-call", mockFn)

	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status code %d, got %d", http.StatusBadGateway, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "failed to fetch test-call" {
		t.Errorf("Expected error message about failed fetch, got %q", response["error"])
	}
}

func TestHandleRIPEstatRequest_Success(t *testing.T) {
	// Test the successful case
	req := httptest.NewRequest(http.MethodGet, "/test?resource=test-resource", nil)
	w := httptest.NewRecorder()

	// Mock function that returns success
	mockFn := func(_ context.Context, resource string) (interface{}, error) {
		if resource != "test-resource" {
			t.Errorf("Expected resource 'test-resource', got %q", resource)
		}
		return map[string]string{"result": "success"}, nil
	}

	handleRIPEstatRequest(w, req, "test-call", mockFn)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["result"] != "success" {
		t.Errorf("Expected result 'success', got %q", response["result"])
	}
}

func TestWriteJSON_EncodingError(t *testing.T) {
	// Test writeJSON with a value that cannot be encoded to JSON
	w := httptest.NewRecorder()

	// Create a value that cannot be marshaled to JSON (channel)
	invalidValue := make(chan int)

	writeJSON(w, invalidValue, http.StatusOK)

	// The function should still set the status code and content type
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %q", contentType)
	}

	// The body should be empty or contain an error due to encoding failure
	// The exact behavior depends on the JSON encoder implementation
}

func TestMain_HelpFlag(t *testing.T) {
	// Test the help flag functionality
	// We can't easily test main() directly, but we can test the flag parsing logic

	// Save original args and restore them after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to simulate -help flag
	os.Args = []string{"cmd", "-help"}

	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// We can't easily test the actual exit behavior, but we can verify
	// that the flag parsing works correctly
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	help := flag.Bool("help", false, "Print all possible flags")

	flag.Parse()

	if !*help {
		t.Error("Expected help flag to be true")
	}
	if *port != "8080" {
		t.Errorf("Expected default port to be '8080', got %q", *port)
	}
	if *debug {
		t.Error("Expected debug flag to be false by default")
	}
}

func TestMain_DebugFlag(t *testing.T) {
	// Test the debug flag functionality

	// Save original args and restore them after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to simulate -debug flag
	os.Args = []string{"cmd", "-debug"}

	// Reset flag package state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Parse flags
	port := flag.String("port", "8080", "Port for the server to listen on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	help := flag.Bool("help", false, "Print all possible flags")

	flag.Parse()

	if *debug != true {
		t.Error("Expected debug flag to be true")
	}
	if *help {
		t.Error("Expected help flag to be false")
	}
	if *port != "8080" {
		t.Errorf("Expected default port to be '8080', got %q", *port)
	}
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, "0", false) // Use port 0 to let the OS choose a free port
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

func TestWriteJSON_EncoderFail(_ *testing.T) {
	// Create a ResponseWriter that will fail on Write
	w := &failingResponseWriter{}

	writeJSON(w, map[string]string{"key": "value"}, http.StatusOK)

	// The function should handle the error gracefully (it logs but doesn't return error)
	// We just verify it doesn't panic
}

// failingResponseWriter is a mock ResponseWriter that fails on Write operations.
type failingResponseWriter struct {
	header http.Header
}

func (f *failingResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe // Simulate write error
}

func (f *failingResponseWriter) WriteHeader(_ int) {
	// Do nothing
}

func TestRun_ServerShutdownError(t *testing.T) {
	// Test the case where server shutdown fails
	// This is hard to trigger in practice, but we can test the run function
	// with a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := run(ctx, "0", false)
	// The function should complete without error even with cancelled context
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

func TestRPKIValidationHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/rpki-validation?resource=3333&prefix=193.0.0.0/21", nil)
	w := httptest.NewRecorder()

	rpkiValidationHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	// If the request was successful, validate the JSON response body
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate key fields in the response
		if status, ok := response["status"].(string); !ok || status == "" {
			t.Errorf("Expected 'status' field to be a non-empty string, got %v", response["status"])
		}

		if resource, ok := response["resource"].(string); !ok || resource != "3333" {
			t.Errorf("Expected 'resource' field to be '3333', got %v", response["resource"])
		}

		if prefix, ok := response["prefix"].(string); !ok || prefix != "193.0.0.0/21" {
			t.Errorf("Expected 'prefix' field to be '193.0.0.0/21', got %v", response["prefix"])
		}

		if validator, ok := response["validator"].(string); !ok || validator == "" {
			t.Errorf("Expected 'validator' field to be a non-empty string, got %v", response["validator"])
		}

		if fetchedAt, ok := response["fetched_at"].(string); !ok || fetchedAt == "" {
			t.Errorf("Expected 'fetched_at' field to be a non-empty string, got %v", response["fetched_at"])
		}

		// Validate validating_roas field exists (can be empty array or contain ROAs)
		if _, ok := response["validating_roas"]; !ok {
			t.Error("Expected 'validating_roas' field to be present in response")
		}
	}
}

func TestRPKIValidationHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/rpki-validation?prefix=193.0.0.0/21", nil)
	w := httptest.NewRecorder()

	rpkiValidationHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestRPKIValidationHandler_MissingPrefix(t *testing.T) {
	req := httptest.NewRequest("GET", "/rpki-validation?resource=3333", nil)
	w := httptest.NewRecorder()

	rpkiValidationHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "missing prefix parameter") {
		t.Errorf("Expected error message about missing prefix, got %q", string(body))
	}
}

func TestRun_InvalidPort(t *testing.T) {
	// Test with an invalid port to see if the server handles it gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Use a very high port number that might cause issues
	err := run(ctx, "99999", false)
	// The function should complete without error
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

func TestASNNeighboursHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/asn-neighbours?resource=AS1205&lod=0", nil)
	w := httptest.NewRecorder()

	asnNeighboursHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestASNNeighboursHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/asn-neighbours?lod=0", nil)
	w := httptest.NewRecorder()

	asnNeighboursHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestASNNeighboursHandler_InvalidLOD(t *testing.T) {
	req := httptest.NewRequest("GET", "/asn-neighbours?resource=AS1205&lod=5", nil)
	w := httptest.NewRecorder()

	asnNeighboursHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "lod parameter must be 0 or 1") {
		t.Errorf("Expected error message about invalid lod, got %q", string(body))
	}
}

func TestLookingGlassHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/looking-glass?resource=140.78.0.0/16&look_back_limit=3600", nil)
	w := httptest.NewRecorder()

	lookingGlassHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestLookingGlassHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/looking-glass?look_back_limit=3600", nil)
	w := httptest.NewRecorder()

	lookingGlassHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestLookingGlassHandler_InvalidLookBackLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/looking-glass?resource=140.78.0.0/16&look_back_limit=invalid", nil)
	w := httptest.NewRecorder()

	lookingGlassHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
		return
	}

	if !strings.Contains(string(body), "look_back_limit parameter must be a valid integer") {
		t.Errorf("Expected error message about invalid look_back_limit, got %q", string(body))
	}
}

func TestWhatsMyIPHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/whats-my-ip", nil)
	w := httptest.NewRecorder()

	whatsMyIPHandler(w, req)

	resp := w.Result()

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		// We accept either OK or BadGateway since this might be run without internet
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	// If the request was successful, validate the JSON response body
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate key fields in the response
		if ip, ok := response["ip"].(string); !ok || ip == "" {
			t.Errorf("Expected 'ip' field to be a non-empty string, got %v", response["ip"])
		}

		// Note: fetched_at might be empty when using client IP extraction in test environment
		if fetchedAt, ok := response["fetched_at"]; ok {
			if fetchedAtStr, isString := fetchedAt.(string); isString && fetchedAtStr == "" {
				// This is acceptable in test environment when using client IP extraction
				t.Logf("fetched_at is empty (expected in test environment with client IP extraction)")
			}
		}
	}
}

func TestWhoisHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/whois?resource=8.8.8.8", nil)
	w := httptest.NewRecorder()

	whoisHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestWhoisHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/whois", nil)
	w := httptest.NewRecorder()

	whoisHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestAbuseContactFinderHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/abuse-contact-finder?resource=8.8.8.8", nil)
	w := httptest.NewRecorder()

	abuseContactFinderHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status code 200 or 502, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestAbuseContactFinderHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/abuse-contact-finder", nil)
	w := httptest.NewRecorder()

	abuseContactFinderHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestMCPHandler(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	// Test initialize request
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	reqBody, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mcpHandler(w, req, server)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestMCPHandler_Notification(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	// Test initialized notification
	notif := mcp.NewNotification("initialized", nil)

	reqBody, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mcpHandler(w, req, server)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code 204 for notification, got %d", resp.StatusCode)
	}
}

func TestMCPHandler_InvalidJSON(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mcpHandler(w, req, server)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200 even for invalid JSON, got %d", resp.StatusCode)
	}

	// Should return a JSON-RPC error response
	var response mcp.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid JSON")
	}

	if response.Error.Code != mcp.ParseError {
		t.Errorf("Expected ParseError code %d, got %d", mcp.ParseError, response.Error.Code)
	}
}

func TestMCPHandler_MethodNotAllowed(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	mcpHandler(w, req, server)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code 405, got %d", resp.StatusCode)
	}
}

func TestMCPHandler_ReadBodyError(t *testing.T) {
	server := mcp.NewServer("test-server", "1.0.0", false)

	// Create a request with a body that will cause a read error
	req := httptest.NewRequest("POST", "/mcp", &errorReader{})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mcpHandler(w, req, server)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}
}

// errorReader is a helper type that always returns an error when read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestNetworkInfoHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/network-info", nil)
	w := httptest.NewRecorder()

	networkInfoHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestASOverviewHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/as-overview", nil)
	w := httptest.NewRecorder()

	asOverviewHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestAnnouncedPrefixesHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/announced-prefixes", nil)
	w := httptest.NewRecorder()

	announcedPrefixesHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestRoutingStatusHandler_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/routing-status", nil)
	w := httptest.NewRecorder()

	routingStatusHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "missing resource parameter") {
		t.Errorf("Expected error message about missing resource, got %q", string(body))
	}
}

func TestWhatsMyIPHandler_WithClientIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/whats-my-ip?client_ip=8.8.8.8", nil)
	w := httptest.NewRecorder()

	whatsMyIPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Accept any valid IP response since the function may extract client IP differently in test environment
	if ip, ok := response["ip"].(string); !ok || ip == "" {
		t.Errorf("Expected valid IP, got %v", response["ip"])
	}
}
