package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewSDKServer(t *testing.T) {
	t.Run("creates server with WhatsMyIP enabled", func(t *testing.T) {
		server := NewSDKServer("test-server", "1.0.0", false)
		if server == nil {
			t.Fatal("Expected server to be created")
		}
		if server.mcpServer == nil {
			t.Error("Expected MCP server to be initialized")
		}
		if server.consolidatedTools == nil {
			t.Error("Expected consolidated tools to be initialized")
		}
		if server.disableWhatsMyIP != false {
			t.Error("Expected disableWhatsMyIP to be false")
		}
	})

	t.Run("creates server with WhatsMyIP disabled", func(t *testing.T) {
		server := NewSDKServer("test-server", "1.0.0", true)
		if server == nil {
			t.Fatal("Expected server to be created")
		}
		if server.disableWhatsMyIP != true {
			t.Error("Expected disableWhatsMyIP to be true")
		}
	})
}

func TestMCPServer(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	mcpServer := server.MCPServer()
	if mcpServer == nil {
		t.Error("Expected MCPServer() to return non-nil server")
	}
	if mcpServer != server.mcpServer {
		t.Error("Expected MCPServer() to return the same server instance")
	}
}

func TestIsValidOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"localhost http", "http://localhost", true},
		{"localhost https", "https://localhost", true},
		{"localhost with port", "http://localhost:3000", true},
		{"127.0.0.1 http", "http://127.0.0.1", true},
		{"127.0.0.1 https", "https://127.0.0.1", true},
		{"127.0.0.1 with port", "http://127.0.0.1:8080", true},
		{"cursor.sh", "https://cursor.sh", true},
		{"cursor.sh with path", "https://cursor.sh/app", true},
		{"claude.ai", "https://claude.ai", true},
		{"invalid origin", "https://evil.com", false},
		{"empty origin", "", false},
		{"partial match fail", "http://local", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidOrigin(tt.origin)
			if result != tt.expected {
				t.Errorf("isValidOrigin(%q) = %v, want %v", tt.origin, result, tt.expected)
			}
		})
	}
}

func TestFormatToolError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"already has Error prefix", errors.New("Error: something went wrong"), "Error: something went wrong"},
		{"no Error prefix", errors.New("something went wrong"), "Error: something went wrong"},
		{"short error", errors.New("fail"), "Error: fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatToolError(tt.err)
			if result != tt.expected {
				t.Errorf("formatToolError(%v) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}

func TestSDKServerCreateToolResultFromJSON(t *testing.T) {
	t.Run("valid JSON data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
			"num": 42,
		}
		result, err := createToolResultFromJSON(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.IsError {
			t.Error("Expected IsError to be false")
		}
		if len(result.Content) == 0 {
			t.Fatal("Expected content to be non-empty")
		}
		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		if !strings.Contains(textContent.Text, "key") {
			t.Errorf("Expected JSON output to contain 'key', got %s", textContent.Text)
		}
	})

	t.Run("unmarshallable data returns error result", func(t *testing.T) {
		// channels cannot be marshaled to JSON
		data := make(chan int)
		result, err := createToolResultFromJSON(data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("Expected IsError to be true for unmarshallable data")
		}
	})
}

func TestHTTPHandler_ServeHTTP(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	t.Run("unsupported protocol version", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("MCP-Protocol-Version", "2024-01-01")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Unsupported protocol version") {
			t.Errorf("Expected error message about protocol version, got %s", w.Body.String())
		}
	})

	t.Run("CORS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Expected CORS headers to be set")
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Expected Access-Control-Allow-Origin to be set")
		}
	})

	t.Run("invalid origin rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("valid origin accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"ping","id":1}`))
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Expected Access-Control-Allow-Origin to be set")
		}
	})

	t.Run("GET without method returns endpoint info", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if response["service"] != "mcp-ripestat" {
			t.Errorf("Expected service 'mcp-ripestat', got %v", response["service"])
		}
	})

	t.Run("protocol version header is set", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("MCP-Protocol-Version") != "2025-06-18" {
			t.Errorf("Expected MCP-Protocol-Version header, got %s", w.Header().Get("MCP-Protocol-Version"))
		}
	})

	t.Run("default protocol version when not specified", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		// No MCP-Protocol-Version header set
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should default to 2025-06-18 and succeed
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestHTTPHandler_HandleCORS(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	t.Run("CORS with valid origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
		req.Header.Set("Origin", "https://cursor.sh")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "https://cursor.sh" {
			t.Error("Expected Access-Control-Allow-Origin for valid origin")
		}
		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Expected Access-Control-Allow-Methods to be set")
		}
		if w.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("Expected Access-Control-Allow-Headers to be set")
		}
		if w.Header().Get("Access-Control-Max-Age") != "86400" {
			t.Error("Expected Access-Control-Max-Age to be set")
		}
	})

	t.Run("CORS with invalid origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// CORS preflight still succeeds but without Allow-Origin
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("Expected Access-Control-Allow-Origin to be empty for invalid origin")
		}
	})

	t.Run("CORS without origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestHTTPHandler_HandleEndpointInfo(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["service"] != "mcp-ripestat" {
		t.Errorf("Expected service 'mcp-ripestat', got %v", response["service"])
	}
	if response["protocol"] != "MCP" {
		t.Errorf("Expected protocol 'MCP', got %v", response["protocol"])
	}
	if response["version"] != "2025-06-18" {
		t.Errorf("Expected version '2025-06-18', got %v", response["version"])
	}

	endpoints, ok := response["endpoints"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected endpoints to be a map")
	}
	if _, ok := endpoints["mcp"]; !ok {
		t.Error("Expected 'mcp' endpoint in response")
	}
}

func TestNewStreamableHTTPHandler(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}

	// Verify it's an httpHandler
	_, ok := handler.(*httpHandler)
	if !ok {
		t.Error("Expected handler to be of type *httpHandler")
	}
}

func TestHTTPRequestContext(t *testing.T) {
	t.Run("WithHTTPRequest and HTTPRequestFromContext", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := WithHTTPRequest(context.Background(), req)

		extractedReq, ok := HTTPRequestFromContext(ctx)
		if !ok {
			t.Error("Expected to extract request from context")
		}
		if extractedReq != req {
			t.Error("Expected extracted request to match original")
		}
	})

	t.Run("HTTPRequestFromContext with no request", func(t *testing.T) {
		ctx := context.Background()
		_, ok := HTTPRequestFromContext(ctx)
		if ok {
			t.Error("Expected ok to be false for context without request")
		}
	})
}
