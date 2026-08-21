//go:build e2e

package e2e

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

func TestStreamableHTTP(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	t.Run("GET request with query parameters", func(t *testing.T) {
		// Legacy GET query style is no longer accepted by the SDK handler.
		params := url.Values{}
		params.Set("method", "ping")
		params.Set("id", "test-ping")

		reqURL := mcpURL + "?" + params.Encode()

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for GET, got %d", resp.StatusCode)
		}
	})

	t.Run("GET request with individual parameters", func(t *testing.T) {
		// Legacy GET query style is no longer accepted by the SDK handler.
		params := url.Values{}
		params.Set("method", "tools/call")
		params.Set("name", "getWhatsMyIP")
		params.Set("id", "test-whatsmyip")

		reqURL := mcpURL + "?" + params.Encode()

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for GET, got %d", resp.StatusCode)
		}
	})

	t.Run("GET request with JSON params", func(t *testing.T) {
		// Legacy GET query style is no longer accepted by the SDK handler.
		params := url.Values{}
		params.Set("method", "tools/call")
		params.Set("params", `{"name": "investigateResource", "arguments": {"resource": "8.8.8.8", "operations": ["overview"], "depth": "basic"}}`)
		params.Set("id", "test-json-params")

		reqURL := mcpURL + "?" + params.Encode()

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for GET, got %d", resp.StatusCode)
		}
	})

	t.Run("CORS preflight request", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", mcpURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, MCP-Protocol-Version")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check CORS headers
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "http://localhost:3000",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Accept, Authorization, MCP-Protocol-Version, Mcp-Method, Mcp-Name, Mcp-Session-Id",
			"Access-Control-Max-Age":       "86400",
		}

		for header, expectedValue := range expectedHeaders {
			actualValue := resp.Header.Get(header)
			if actualValue != expectedValue {
				t.Errorf("Expected %s '%s', got '%s'", header, expectedValue, actualValue)
			}
		}
	})

	t.Run("invalid origin rejection", func(t *testing.T) {
		req, err := http.NewRequest("POST", mcpURL, strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "https://malicious.com")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("unsupported protocol version", func(t *testing.T) {
		req, err := http.NewRequest("POST", mcpURL, strings.NewReader(`{"jsonrpc":"2.0","method":"ping","id":1}`))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2024-01-01")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("stateless discovery", func(t *testing.T) {
		discoverReq := mcp.NewRequest("server/discover", nil, 1001)
		resp, discoverResult := sendMCPRequestWithHeaders(t, mcpURL, discoverReq, nil)
		if discoverResult.Error != nil {
			t.Fatalf("server/discover failed: %v", discoverResult.Error)
		}
		if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
			t.Fatalf("stateless response unexpectedly set session ID %q", sessionID)
		}
		result := discoverResult.Result.(map[string]interface{})
		if result["resultType"] != mcp.ResultTypeComplete {
			t.Errorf("resultType = %v", result["resultType"])
		}
	})

	t.Run("supported protocol version", func(t *testing.T) {
		response := sendMCPRequest(t, mcpURL, mcp.NewRequest("server/discover", nil, 1))
		if response.Error != nil {
			t.Fatalf("server/discover failed: %v", response.Error)
		}
		result := response.Result.(map[string]interface{})
		versions := result["supportedVersions"].([]interface{})
		if len(versions) == 0 || versions[0] != mcp.ProtocolVersion {
			t.Errorf("supportedVersions = %v", versions)
		}
	})

	t.Run("GET request missing method parameter", func(t *testing.T) {
		// Stateless 2026 transport does not expose endpoint information over GET.
		req, err := http.NewRequest("GET", mcpURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for GET, got %d", resp.StatusCode)
		}
	})

	t.Run("unsupported HTTP method", func(t *testing.T) {
		req, err := http.NewRequest("PUT", mcpURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestStreamableHTTPCapabilities(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// Modern clients discover capabilities without an initialization handshake.
	req := mcp.NewRequest("server/discover", nil, 1)

	response := sendMCPRequest(t, mcpURL, req)

	if response.Error != nil {
		t.Fatalf("server/discover failed: %v", response.Error)
	}

	// Verify response structure
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Initialize result is not an object")
	}

	capabilities, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Capabilities is not an object")
	}

	if _, hasTools := capabilities["tools"]; !hasTools {
		t.Error("Tools capability not present")
	}

	meta, ok := result["_meta"].(map[string]interface{})
	if !ok {
		t.Fatal("_meta is not an object")
	}
	serverInfo, ok := meta[mcp.MetaKeyServerInfo].(map[string]interface{})
	if !ok {
		t.Fatal("serverInfo is not an object")
	}
	if serverInfo["name"] == "" {
		t.Error("Server name should not be empty")
	}
}

func TestStreamableHTTPFullWorkflow(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// Step 1: List tools directly; modern requests are self-describing.
	listReq := mcp.NewRequest("tools/list", nil, "list-tools")
	listResp, listResponse := sendMCPRequestWithHeaders(t, mcpURL, listReq, nil)
	if listResponse.Error != nil {
		t.Fatalf("Tools list failed: %v", listResponse.Error)
	}
	if sessionID := listResp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		t.Fatalf("stateless response unexpectedly set session ID %q", sessionID)
	}

	// Verify tools list
	result, ok := listResponse.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Tools list result is not an object")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools is not an array")
	}

	if len(tools) == 0 {
		t.Error("Expected at least one tool")
	}

	// Step 2: Call a tool in an independent POST.
	callReq := mcp.NewRequest("tools/call", map[string]interface{}{
		"name":      "getWhatsMyIP",
		"arguments": map[string]interface{}{},
	}, "call-whatsmyip")
	_, callResponse := sendMCPRequestWithHeaders(t, mcpURL, callReq, nil)

	// Tool call should succeed (or fail due to network, but not protocol error)
	if callResponse.Error != nil {
		// Check if it's a tool execution error vs protocol error
		if callResponse.Error.Code < -32000 {
			// MCP protocol error - this is unexpected
			t.Fatalf("Tool call failed with protocol error: %v", callResponse.Error)
		}
		// Network or tool execution error is acceptable in tests
		t.Logf("Tool call failed (likely network issue): %v", callResponse.Error)
	}
}
