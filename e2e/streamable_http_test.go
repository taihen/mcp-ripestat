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
		// Test ping method via GET
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check response headers
		if resp.Header.Get("MCP-Protocol-Version") == "" {
			t.Error("Expected MCP-Protocol-Version header in response")
		}

		if resp.Header.Get("MCP-Session-ID") == "" {
			t.Error("Expected MCP-Session-ID header in response")
		}

		// Parse response
		var response mcp.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Error != nil {
			t.Fatalf("Ping failed: %v", response.Error)
		}
	})

	t.Run("GET request with individual parameters", func(t *testing.T) {
		// Test tools/call via GET with individual parameters
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response mcp.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should fail because we need to initialize first, but HTTP routing works
		if response.Error == nil {
			t.Error("Expected error for tools/call without initialization")
		}
	})

	t.Run("GET request with JSON params", func(t *testing.T) {
		// Test with JSON params parameter
		params := url.Values{}
		params.Set("method", "tools/call")
		params.Set("params", `{"name": "getNetworkInfo", "arguments": {"resource": "8.8.8.8"}}`)
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response mcp.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should fail because we need to initialize first
		if response.Error == nil {
			t.Error("Expected error for tools/call without initialization")
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
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, MCP-Protocol-Version, MCP-Session-ID",
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
		req, err := http.NewRequest("POST", mcpURL, strings.NewReader("{}"))
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
		// First request - should get new session
		req1, err := http.NewRequest("GET", mcpURL+"?method=ping&id=1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req1.Header.Set("Origin", "http://localhost:3000")
		req1.Header.Set("MCP-Protocol-Version", "2025-06-18")

		client := &http.Client{}
		resp1, err := client.Do(req1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp1.Body.Close()

		sessionID := resp1.Header.Get("MCP-Session-ID")
		if sessionID == "" {
			t.Fatal("Expected session ID in first response")
		}

		// Second request - should use existing session
		req2, err := http.NewRequest("GET", mcpURL+"?method=ping&id=2", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req2.Header.Set("Origin", "http://localhost:3000")
		req2.Header.Set("MCP-Protocol-Version", "2025-06-18")
		req2.Header.Set("MCP-Session-ID", sessionID)

		resp2, err := client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp2.Body.Close()

		// Should not create new session ID
		newSessionID := resp2.Header.Get("MCP-Session-ID")
		if newSessionID != "" {
			t.Error("Expected no new session ID when existing session provided")
		}
	})

	t.Run("backward compatibility", func(t *testing.T) {
		// Test with old protocol version
		req, err := http.NewRequest("GET", mcpURL+"?method=ping&id=1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("MCP-Protocol-Version", "2025-03-26")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for backward compatible version, got %d", resp.StatusCode)
		}

		// Check protocol version in response
		responseVersion := resp.Header.Get("MCP-Protocol-Version")
		if responseVersion != "2025-03-26" {
			t.Errorf("Expected response protocol version '2025-03-26', got '%s'", responseVersion)
		}
	})

	t.Run("GET request missing method parameter", func(t *testing.T) {
		// Test GET without method parameter
		req, err := http.NewRequest("GET", mcpURL+"?id=1", nil)
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
			t.Errorf("Expected status 400 for missing method, got %d", resp.StatusCode)
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

	// Test that initialize returns transport capabilities
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

	transport, ok := capabilities["transport"].(map[string]interface{})
	if !ok {
		t.Fatal("Transport capability missing")
	}

	httpTransport, ok := transport["http"].(map[string]interface{})
	if !ok {
		t.Fatal("HTTP transport capability missing")
	}

	streamable, ok := httpTransport["streamable"].(bool)
	if !ok || !streamable {
		t.Error("Expected HTTP transport to be streamable")
	}

	methods, ok := httpTransport["methods"].([]interface{})
	if !ok {
		t.Fatal("HTTP methods missing")
	}

	expectedMethods := []string{"POST", "GET"}
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(methods))
	}

	for i, method := range methods {
		methodStr, ok := method.(string)
		if !ok {
			t.Errorf("Method %d is not a string", i)
			continue
		}
		if methodStr != expectedMethods[i] {
			t.Errorf("Expected method %s at index %d, got %s", expectedMethods[i], i, methodStr)
		}
	}
}

func TestStreamableHTTPFullWorkflow(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// Step 1: Initialize via POST
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	initResponse := sendMCPRequest(t, mcpURL, initReq)
	if initResponse.Error != nil {
		t.Fatalf("Initialize failed: %v", initResponse.Error)
	}

	// Step 2: Send initialized notification via POST
	initializedNotif := mcp.NewNotification("initialized", nil)
	sendMCPNotificationViaHTTP(t, mcpURL, initializedNotif)

	// Step 3: List tools via GET
	params := url.Values{}
	params.Set("method", "tools/list")
	params.Set("id", "list-tools")

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

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var listResponse mcp.Response
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

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

	// Step 4: Call a tool via GET
	params = url.Values{}
	params.Set("method", "tools/call")
	params.Set("name", "getWhatsMyIP")
	params.Set("id", "call-whatsmyip")

	reqURL = mcpURL + "?" + params.Encode()

	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("MCP-Protocol-Version", "2025-06-18")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var callResponse mcp.Response
	if err := json.NewDecoder(resp.Body).Decode(&callResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204 for notification, got %d", resp.StatusCode)
	}
}
