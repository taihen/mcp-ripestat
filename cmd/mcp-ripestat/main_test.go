package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
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

	if len(manifest.Functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(manifest.Functions))
	}

	funcNames := map[string]bool{}
	for _, f := range manifest.Functions {
		funcNames[f.Name] = true
	}

	if !funcNames["getNetworkInfo"] {
		t.Errorf("manifest is missing function getNetworkInfo")
	}

	if !funcNames["getASOverview"] {
		t.Errorf("manifest is missing function getASOverview")
	}

	if len(manifest.Functions[0].Parameters) != 1 {
		t.Fatalf("expected 1 parameter for getNetworkInfo, got %d", len(manifest.Functions[0].Parameters))
	}

	parameter := manifest.Functions[0].Parameters[0]
	if parameter.Name != "resource" {
		t.Errorf("unexpected parameter name for getNetworkInfo: got %q want %q", parameter.Name, "resource")
	}
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

func TestASOverviewHandler_Success(t *testing.T) {
	// Temporarily replace the real Get with a mock
	originalGet := asoverview.Get
	asoverview.Get = func(ctx context.Context, resource string) (*asoverview.Response, error) {
		return &asoverview.Response{
			Data: asoverview.Data{
				Resource:  "3333",
				Announced: true,
			},
		}, nil
	}
	defer func() { asoverview.Get = originalGet }()

	req, err := http.NewRequest("GET", "/as-overview?resource=3333", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(asOverviewHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp asoverview.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Data.Resource != "3333" {
		t.Errorf("unexpected resource: got %s want %s", resp.Data.Resource, "3333")
	}
}

func TestASOverviewHandler_MissingResource(t *testing.T) {
	req, err := http.NewRequest("GET", "/as-overview", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(asOverviewHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestASOverviewHandler_GetError(t *testing.T) {
	originalGet := asoverview.Get
	asoverview.Get = func(ctx context.Context, resource string) (*asoverview.Response, error) {
		return nil, errors.New("test error")
	}
	defer func() { asoverview.Get = originalGet }()

	req, err := http.NewRequest("GET", "/as-overview?resource=3333", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(asOverviewHandler)
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
