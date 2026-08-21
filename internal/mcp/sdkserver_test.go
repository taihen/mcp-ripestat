package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		{"cursor.sh", "https://cursor.sh", false},
		{"cursor.sh with path", "https://cursor.sh/app", false},
		{"cursor.sh malicious suffix domain", "https://cursor.sh.evil.com", false},
		{"cursor.sh non-default port", "https://cursor.sh:8443", false},
		{"claude.ai", "https://claude.ai", false},
		{"scheme mismatch", "http://cursor.sh", false},
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

func TestEffectivePort(t *testing.T) {
	tests := []struct {
		name   string
		scheme string
		port   string
		want   string
	}{
		{name: "explicit port", scheme: "https", port: "8443", want: "8443"},
		{name: "default HTTP", scheme: "http", want: "80"},
		{name: "default HTTPS", scheme: "https", want: "443"},
		{name: "unknown scheme", scheme: "custom", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectivePort(tt.scheme, tt.port); got != tt.want {
				t.Fatalf("effectivePort(%q, %q) = %q, want %q", tt.scheme, tt.port, got, tt.want)
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
		result := createToolResultFromJSON(data)
		if result.IsError {
			t.Error("Expected IsError to be false")
		}
		if len(result.Content) == 0 {
			t.Fatal("Expected content to be non-empty")
		}
		// Check content text contains expected data
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(interface{ GetText() string }); ok {
				if !strings.Contains(textContent.GetText(), "key") {
					t.Errorf("Expected JSON output to contain 'key'")
				}
			}
		}
	})

	t.Run("unmarshallable data returns error result", func(t *testing.T) {
		// channels cannot be marshaled to JSON
		data := make(chan int)
		result := createToolResultFromJSON(data)
		if !result.IsError {
			t.Error("Expected IsError to be true for unmarshallable data")
		}
	})
}

func TestWithPanicRecovery(t *testing.T) {
	t.Run("passes through normal handler", func(t *testing.T) {
		handler := func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "success"}},
			}, nil
		}

		wrapped := withPanicRecovery(handler, "testTool")
		result, err := wrapped(context.Background(), &mcp.CallToolRequest{})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.IsError {
			t.Error("Expected IsError to be false")
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		handler := func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			panic("test panic")
		}

		wrapped := withPanicRecovery(handler, "testTool")
		result, err := wrapped(context.Background(), &mcp.CallToolRequest{})

		if err != nil {
			t.Errorf("Expected no error after recovery, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result after recovery")
		}
		if !result.IsError {
			t.Error("Expected IsError to be true after panic recovery")
		}
	})
}

func TestHTTPHandler_ServeHTTP(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	t.Run("unsupported protocol version", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2099-01-01","io.modelcontextprotocol/clientCapabilities":{}}}}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("MCP-Protocol-Version", "2099-01-01")
		req.Header.Set("Mcp-Method", "server/discover")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(strings.ToLower(w.Body.String()), "unsupported protocol version") {
			t.Errorf("Expected error message about protocol version, got %s", w.Body.String())
		}
	})

	t.Run("CORS preflight request", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/mcp", nil)
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("valid origin accepted", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"ping","id":1}`))
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Error("Expected Access-Control-Allow-Origin to be set")
		}
	})

	t.Run("GET is rejected by stateless transport", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/mcp", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
		if w.Header().Get("Allow") != "POST" {
			t.Errorf("Expected Allow POST, got %q", w.Header().Get("Allow"))
		}
	})

}

func TestHTTPHandler_Protocol20260728(t *testing.T) {
	t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "")
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	post := func(t *testing.T, version, method, name string, params map[string]interface{}) *httptest.ResponseRecorder {
		t.Helper()
		if params == nil {
			params = make(map[string]interface{})
		}
		params["_meta"] = map[string]interface{}{
			MetaKeyProtocolVersion:    version,
			MetaKeyClientCapabilities: map[string]interface{}{},
			MetaKeyClientInfo:         map[string]interface{}{"name": "test-client", "version": "1.0.0"},
		}
		body, err := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  method,
			"params":  params,
		})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("MCP-Protocol-Version", version)
		req.Header.Set("Mcp-Method", method)
		if name != "" {
			req.Header.Set("Mcp-Name", name)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}

	decode := func(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
		t.Helper()
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response %q: %v", w.Body.String(), err)
		}
		return response
	}

	t.Run("discover advertises modern protocol without session", func(t *testing.T) {
		w := post(t, ProtocolVersion, "server/discover", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		if got := w.Header().Get("Mcp-Session-Id"); got != "" {
			t.Errorf("stateless response set Mcp-Session-Id %q", got)
		}
		result := decode(t, w)["result"].(map[string]interface{})
		if result["resultType"] != ResultTypeComplete {
			t.Errorf("resultType = %v", result["resultType"])
		}
		if result["cacheScope"] != CacheScopePublic {
			t.Errorf("cacheScope = %v", result["cacheScope"])
		}
		if result["ttlMs"] != float64(ToolsListTTLMs) {
			t.Errorf("discover ttlMs = %v", result["ttlMs"])
		}
		meta, ok := result["_meta"].(map[string]interface{})
		if !ok || meta[MetaKeyServerInfo] == nil {
			t.Errorf("discover result missing server info: %v", result["_meta"])
		}
		versions := result["supportedVersions"].([]interface{})
		if len(versions) != 1 || versions[0] != ProtocolVersion {
			t.Errorf("supportedVersions = %v", versions)
		}
	})

	t.Run("tools list is complete cacheable and deterministic", func(t *testing.T) {
		w := post(t, ProtocolVersion, "tools/list", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		result := decode(t, w)["result"].(map[string]interface{})
		if result["resultType"] != ResultTypeComplete || result["cacheScope"] != CacheScopePublic {
			t.Errorf("unexpected result metadata: %v", result)
		}
		if result["ttlMs"] != float64(ToolsListTTLMs) {
			t.Errorf("tools/list ttlMs = %v", result["ttlMs"])
		}
		tools := result["tools"].([]interface{})
		for i := 1; i < len(tools); i++ {
			previous := tools[i-1].(map[string]interface{})["name"].(string)
			current := tools[i].(map[string]interface{})["name"].(string)
			if previous > current {
				t.Fatalf("tools are not sorted: %q before %q", previous, current)
			}
		}
	})

	t.Run("tool result carries complete type and server identity", func(t *testing.T) {
		w := post(t, ProtocolVersion, "tools/call", "getWhatsMyIP", map[string]interface{}{
			"name":      "getWhatsMyIP",
			"arguments": map[string]interface{}{},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		result := decode(t, w)["result"].(map[string]interface{})
		if result["resultType"] != ResultTypeComplete {
			t.Errorf("resultType = %v", result["resultType"])
		}
		meta := result["_meta"].(map[string]interface{})
		if meta[MetaKeyServerInfo] == nil {
			t.Error("tool result omitted serverInfo")
		}
	})

	t.Run("header body mismatch is rejected", func(t *testing.T) {
		w := post(t, ProtocolVersion, "tools/call", "differentTool", map[string]interface{}{
			"name":      "getWhatsMyIP",
			"arguments": map[string]interface{}{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		errorObject := decode(t, w)["error"].(map[string]interface{})
		if errorObject["code"] != float64(HeaderMismatchError) {
			t.Errorf("error code = %v", errorObject["code"])
		}
	})

	t.Run("unsupported version returns protocol error", func(t *testing.T) {
		w := post(t, "2099-01-01", "server/discover", "", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		errorObject := decode(t, w)["error"].(map[string]interface{})
		if errorObject["code"] != float64(UnsupportedProtocolVersionError) {
			t.Errorf("error code = %v", errorObject["code"])
		}
	})

	t.Run("removed ping method is not found", func(t *testing.T) {
		w := post(t, ProtocolVersion, "ping", "", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		errorObject := decode(t, w)["error"].(map[string]interface{})
		if errorObject["code"] != float64(MethodNotFound) {
			t.Errorf("error code = %v", errorObject["code"])
		}
	})
}

func TestHTTPHandler_LegacyProtocolPolicy(t *testing.T) {
	initializeBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"legacy-test","version":"1.0.0"}}}`

	t.Run("legacy is rejected by default", func(t *testing.T) {
		t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "")
		handler := NewSDKServer("test-server", "1.0.0", false).NewStreamableHTTPHandler()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(initializeBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		var response Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Error == nil || response.Error.Code != HeaderMismatchError {
			t.Fatalf("error = %+v", response.Error)
		}
		if response.ID != float64(1) {
			t.Fatalf("response id = %v, want 1", response.ID)
		}
	})

	t.Run("legacy can be explicitly enabled", func(t *testing.T) {
		t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "true")
		handler := NewSDKServer("test-server", "1.0.0", false).NewStreamableHTTPHandler()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(initializeBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		result := response["result"].(map[string]interface{})
		if result["protocolVersion"] != "2025-06-18" {
			t.Fatalf("protocolVersion = %v", result["protocolVersion"])
		}

		discoverBody := `{"jsonrpc":"2.0","id":2,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{}}}}`
		discoverReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(discoverBody))
		discoverReq.Header.Set("Content-Type", "application/json")
		discoverReq.Header.Set("Accept", "application/json, text/event-stream")
		discoverReq.Header.Set("MCP-Protocol-Version", ProtocolVersion)
		discoverReq.Header.Set("Mcp-Method", "server/discover")
		discoverRecorder := httptest.NewRecorder()
		handler.ServeHTTP(discoverRecorder, discoverReq)
		if discoverRecorder.Code != http.StatusOK {
			t.Fatalf("discover status = %d, body = %s", discoverRecorder.Code, discoverRecorder.Body.String())
		}
		var discoverResponse map[string]interface{}
		if err := json.Unmarshal(discoverRecorder.Body.Bytes(), &discoverResponse); err != nil {
			t.Fatal(err)
		}
		versions := discoverResponse["result"].(map[string]interface{})["supportedVersions"].([]interface{})
		wantVersions := []string{"2026-07-28", "2025-11-25", "2025-06-18", "2025-03-26"}
		if len(versions) != len(wantVersions) {
			t.Fatalf("supportedVersions = %v", versions)
		}
		for i, want := range wantVersions {
			if versions[i] != want {
				t.Fatalf("supportedVersions = %v", versions)
			}
		}
	})

	t.Run("retired 2024 revision remains rejected when legacy is enabled", func(t *testing.T) {
		t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "true")
		handler := NewSDKServer("test-server", "1.0.0", false).NewStreamableHTTPHandler()
		body := `{"jsonrpc":"2.0","id":"legacy-2024","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"legacy-test","version":"1.0.0"}}}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		var response Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Error == nil || response.Error.Code != UnsupportedProtocolVersionError {
			t.Fatalf("error = %+v", response.Error)
		}
		if response.ID != "legacy-2024" {
			t.Fatalf("response id = %v", response.ID)
		}
		supported := response.Error.Data.(map[string]interface{})["supported"].([]interface{})
		wantSupported := []string{"2026-07-28", "2025-11-25", "2025-06-18", "2025-03-26"}
		if len(supported) != len(wantSupported) {
			t.Fatalf("error supported versions = %v", supported)
		}
		for i, want := range wantSupported {
			if supported[i] != want {
				t.Fatalf("error supported versions = %v", supported)
			}
		}
	})

	for _, tt := range []struct {
		name        string
		bodyVersion string
		requestID   int
		wantCode    int
	}{
		{"legacy header cannot conceal retired body version", "2024-11-05", 7, UnsupportedProtocolVersionError},
		{"legacy initialize header and body must match", "2025-06-18", 8, HeaderMismatchError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "true")
			handler := NewSDKServer("test-server", "1.0.0", false).NewStreamableHTTPHandler()
			body := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"initialize","params":{"protocolVersion":%q,"capabilities":{},"clientInfo":{"name":"legacy-test","version":"1.0.0"}}}`, tt.requestID, tt.bodyVersion)
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			req.Header.Set("MCP-Protocol-Version", "2025-03-26")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusBadRequest || response.Error == nil || response.Error.Code != tt.wantCode {
				t.Fatalf("status = %d, error = %+v", w.Code, response.Error)
			}
			if response.ID != float64(tt.requestID) {
				t.Fatalf("response id = %v", response.ID)
			}
		})
	}

}

func TestHTTPHandler_LegacyBatchProtocolPolicy(t *testing.T) {
	t.Setenv("MCP_ENABLE_LEGACY_PROTOCOLS", "true")
	handler := NewSDKServer("test-server", "1.0.0", false).NewStreamableHTTPHandler()
	body := `[{"jsonrpc":"2.0","id":"batch-smuggle","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"legacy-test","version":"1.0.0"}}}]`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-03-26")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest || response.Error == nil || response.Error.Code != UnsupportedProtocolVersionError {
		t.Fatalf("status = %d, error = %+v", w.Code, response.Error)
	}
	if response.ID != "batch-smuggle" {
		t.Fatalf("response id = %v", response.ID)
	}
}

func TestHTTPHandler_HandleCORS(t *testing.T) {
	server := NewSDKServer("test-server", "1.0.0", false)
	handler := server.NewStreamableHTTPHandler()

	t.Run("CORS with valid origin", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/mcp", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/mcp", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("Expected Access-Control-Allow-Origin to be empty for invalid origin")
		}
	})

	t.Run("CORS without origin", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/mcp", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
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
