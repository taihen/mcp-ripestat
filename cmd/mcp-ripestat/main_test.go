package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestHandleRIPEstatRequest_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?resource=123", nil)
	rr := httptest.NewRecorder()

	mockFunc := func(ctx context.Context, resource string) (interface{}, error) {
		if resource != "123" {
			t.Errorf("expected resource 123, got %s", resource)
		}
		return map[string]string{"data": "success"}, nil
	}

	handleRIPEstatRequest(rr, req, "test", mockFunc)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp["data"] != "success" {
		t.Errorf("unexpected response body: got %v", resp)
	}
}

func TestHandleRIPEstatRequest_MissingResource(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	mockFunc := func(ctx context.Context, resource string) (interface{}, error) {
		t.Fatal("mock function should not be called")
		return nil, nil
	}

	handleRIPEstatRequest(rr, req, "test", mockFunc)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleRIPEstatRequest_BackendError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?resource=123", nil)
	rr := httptest.NewRecorder()

	mockFunc := func(ctx context.Context, resource string) (interface{}, error) {
		return nil, errors.New("backend failure")
	}

	handleRIPEstatRequest(rr, req, "test", mockFunc)

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
