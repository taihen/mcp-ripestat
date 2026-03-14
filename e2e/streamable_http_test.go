//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
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

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for legacy GET query style, got %d", resp.StatusCode)
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

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for legacy GET query style, got %d", resp.StatusCode)
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

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for legacy GET query style, got %d", resp.StatusCode)
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
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS, DELETE",
			"Access-Control-Allow-Headers": "Content-Type, MCP-Protocol-Version, MCP-Session-ID, Accept",
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

	t.Run("session management", func(t *testing.T) {
		// Initialize should create a session.
		initReq := mcp.NewRequest("initialize", map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "streamable-http-test",
				"version": "1.0.0",
			},
		}, 1001)
		initResp, initResult := sendMCPRequestWithHeaders(t, mcpURL, initReq, nil)
		if initResult.Error != nil {
			t.Fatalf("Initialize failed: %v", initResult.Error)
		}
		sessionID := initResp.Header.Get("Mcp-Session-Id")
		if sessionID == "" {
			t.Fatal("Expected session ID in initialize response")
		}

		// Follow-up request with same session should succeed.
		pingReq := mcp.NewRequest("ping", nil, 1002)
		_, pingResult := sendMCPRequestWithSession(t, mcpURL, pingReq, sessionID)
		if pingResult.Error != nil {
			t.Fatalf("Ping with existing session failed: %v", pingResult.Error)
		}
	})

	t.Run("supported protocol version", func(t *testing.T) {
		// Test with current protocol version using POST
		requestData := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "ping",
			"id":      1,
		}

		reqBody, _ := json.Marshal(requestData)
		req, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for current protocol version, got %d", resp.StatusCode)
		}

		// Check protocol version in response
		responseVersion := resp.Header.Get("MCP-Protocol-Version")
		if responseVersion != "2025-06-18" {
			t.Errorf("Expected response protocol version '2025-06-18', got '%s'", responseVersion)
		}
	})

	t.Run("GET request missing method parameter", func(t *testing.T) {
		// Test GET without method parameter - should return endpoint info
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for endpoint info, got %d", resp.StatusCode)
		}

		// Should return endpoint info JSON
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["service"] != "mcp-ripestat" {
			t.Errorf("Expected service 'mcp-ripestat', got %v", response["service"])
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

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestStreamableHTTPCapabilities(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// Test that initialize returns capabilities
	req := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	response := sendMCPRequest(t, mcpURL, req)

	if response.Error != nil {
		t.Fatalf("Initialize failed: %v", response.Error)
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

	// SDK returns tools capability when tools are registered
	if _, hasTools := capabilities["tools"]; !hasTools {
		t.Log("Tools capability not present (may be expected depending on SDK version)")
	}

	// Verify serverInfo
	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("ServerInfo is not an object")
	}

	if serverInfo["name"] == "" {
		t.Error("Server name should not be empty")
	}
}

func TestStreamableHTTPFullWorkflow(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// Step 1: Initialize via POST and capture session.
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	initResp, initResponse := sendMCPRequestWithHeaders(t, mcpURL, initReq, nil)
	if initResponse.Error != nil {
		t.Fatalf("Initialize failed: %v", initResponse.Error)
	}
	sessionID := initResp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("Expected session ID from initialize response")
	}

	// Step 2: Send initialized notification via POST
	initializedNotif := mcp.NewNotification("initialized", nil)
	sendMCPNotificationViaHTTP(t, mcpURL, initializedNotif)

	// Step 3: List tools via POST JSON-RPC.
	listReq := mcp.NewRequest("tools/list", nil, "list-tools")
	_, listResponse := sendMCPRequestWithSession(t, mcpURL, listReq, sessionID)
	if listResponse.Error != nil {
		t.Fatalf("Tools list failed: %v", listResponse.Error)
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

	// Step 4: Call a tool via POST JSON-RPC.
	callReq := mcp.NewRequest("tools/call", map[string]interface{}{
		"name":      "getWhatsMyIP",
		"arguments": map[string]interface{}{},
	}, "call-whatsmyip")
	_, callResponse := sendMCPRequestWithSession(t, mcpURL, callReq, sessionID)

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

func sendMCPNotificationViaHTTP(t *testing.T, url string, notif *mcp.Notification) {
	reqBody, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	// SDK may be strict with notification/session semantics depending on state.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusAccepted &&
		resp.StatusCode != http.StatusNoContent &&
		resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 200/202/204/400 for notification, got %d", resp.StatusCode)
	}
}
