package main

import (
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

	manifestHandler(w, req)

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

	if len(manifest.Functions) != 8 {
		t.Errorf("Expected 8 functions in manifest, got %d", len(manifest.Functions))
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
	}

	for _, name := range expectedFunctions {
		if !functionNames[name] {
			t.Errorf("Expected function %q in manifest", name)
		}
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
		errCh <- run(ctx, port)
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
		errCh <- run(ctx, port)
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

	err := run(ctx, "0") // Use port 0 to let the OS choose a free port
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

	err := run(ctx, "0")
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
	err := run(ctx, "99999")
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
