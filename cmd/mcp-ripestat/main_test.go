package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
)

func TestManifestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/.well-known/mcp/manifest.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(manifestHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var manifest Manifest
	if err := json.NewDecoder(rr.Body).Decode(&manifest); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if manifest.Name != "mcp-ripestat" {
		t.Errorf("unexpected manifest name: got %q want %q", manifest.Name, "mcp-ripestat")
	}

	if len(manifest.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(manifest.Functions))
	}

	function := manifest.Functions[0]
	if function.Name != "getNetworkInfo" {
		t.Errorf("unexpected function name: got %q want %q", function.Name, "getNetworkInfo")
	}

	if len(function.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(function.Parameters))
	}

	parameter := function.Parameters[0]
	if parameter.Name != "resource" {
		t.Errorf("unexpected parameter name: got %q want %q", parameter.Name, "resource")
	}
}

// Mock GetNetworkInfo function for testing
var mockGetNetworkInfo func(ctx context.Context, resource string) (*networkinfo.NetworkInfoResponse, error)

type mockNetworkInfo struct{}

func (m *mockNetworkInfo) GetNetworkInfo(ctx context.Context, resource string) (*networkinfo.NetworkInfoResponse, error) {
	return mockGetNetworkInfo(ctx, resource)
}

func TestNetworkInfoHandler_Success(t *testing.T) {
	// Temporarily replace the real GetNetworkInfo with a mock
	originalGetNetworkInfo := networkinfo.GetNetworkInfo
	networkinfo.GetNetworkInfo = func(ctx context.Context, resource string) (*networkinfo.NetworkInfoResponse, error) {
		return &networkinfo.NetworkInfoResponse{
			Data: networkinfo.NetworkInfoData{
				Prefix: "1.1.1.0/24",
				ASNs:   []string{"13335"},
			},
		}, nil
	}
	defer func() { networkinfo.GetNetworkInfo = originalGetNetworkInfo }()

	req, err := http.NewRequest("GET", "/network-info?resource=1.1.1.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(networkInfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp networkinfo.NetworkInfoResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Data.Prefix != "1.1.1.0/24" {
		t.Errorf("unexpected prefix: got %s want %s", resp.Data.Prefix, "1.1.1.0/24")
	}
}

func TestNetworkInfoHandler_MissingResource(t *testing.T) {
	req, err := http.NewRequest("GET", "/network-info", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(networkInfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestNetworkInfoHandler_GetNetworkInfoError(t *testing.T) {
	originalGetNetworkInfo := networkinfo.GetNetworkInfo
	networkinfo.GetNetworkInfo = func(ctx context.Context, resource string) (*networkinfo.NetworkInfoResponse, error) {
		return nil, errors.New("test error")
	}
	defer func() { networkinfo.GetNetworkInfo = originalGetNetworkInfo }()

	req, err := http.NewRequest("GET", "/network-info?resource=1.1.1.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(networkInfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadGateway {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadGateway)
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	writeJSON(rr, data, http.StatusCreated)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("writeJSON returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var target map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&target); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if target["foo"] != "bar" {
		t.Errorf("unexpected body: got %v want %v", rr.Body.String(), `{"foo":"bar"}`)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("unexpected content type: got %s want %s", rr.Header().Get("Content-Type"), "application/json")
	}
}

func TestWriteJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, "test error", http.StatusInternalServerError)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("writeJSONError returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	var target map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&target); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if target["error"] != "test error" {
		t.Errorf("unexpected error message: got %s want %s", target["error"], "test error")
	}
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel the context after a short delay to allow the server to start
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Use a random port
	err := run(ctx, "0")
	if err != nil {
		t.Errorf("run function returned an error: %v", err)
	}
}
