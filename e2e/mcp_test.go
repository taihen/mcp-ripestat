//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

func TestMCPProtocol(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	t.Run("Initialize", func(t *testing.T) {
		req := mcp.NewRequest("initialize", map[string]interface{}{
			"protocolVersion": "2025-03-26",
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

		if result["protocolVersion"] != "2025-03-26" {
			t.Errorf("Expected protocolVersion '2025-03-26', got %v", result["protocolVersion"])
		}

		if _, ok := result["capabilities"]; !ok {
			t.Error("Initialize result missing capabilities")
		}

		if _, ok := result["serverInfo"]; !ok {
			t.Error("Initialize result missing serverInfo")
		}
	})

	t.Run("Initialized", func(t *testing.T) {
		notif := mcp.NewNotification("initialized", nil)

		reqBody, err := json.Marshal(notif)
		if err != nil {
			t.Fatalf("Failed to marshal notification: %v", err)
		}

		resp, err := http.Post(mcpURL, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to send initialized notification: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204 for notification, got %d", resp.StatusCode)
		}
	})

	t.Run("ToolsList", func(t *testing.T) {
		req := mcp.NewRequest("tools/list", nil, 2)
		response := sendMCPRequest(t, mcpURL, req)

		if response.Error != nil {
			t.Fatalf("Tools/list failed: %v", response.Error)
		}

		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Tools/list result is not an object")
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatal("Tools/list result missing tools array")
		}

		if len(tools) == 0 {
			t.Error("Expected at least one tool")
		}

		// Verify each tool has required fields
		for i, toolData := range tools {
			tool, ok := toolData.(map[string]interface{})
			if !ok {
				t.Errorf("Tool %d is not an object", i)
				continue
			}

			if _, ok := tool["name"]; !ok {
				t.Errorf("Tool %d missing name", i)
			}
			if _, ok := tool["description"]; !ok {
				t.Errorf("Tool %d missing description", i)
			}
			if _, ok := tool["inputSchema"]; !ok {
				t.Errorf("Tool %d missing inputSchema", i)
			}
		}
	})

	t.Run("ToolsCall_GetNetworkInfo", func(t *testing.T) {
		req := mcp.NewRequest("tools/call", map[string]interface{}{
			"name": "getNetworkInfo",
			"arguments": map[string]interface{}{
				"resource": "8.8.8.8",
			},
		}, 3)

		response := sendMCPRequest(t, mcpURL, req)

		if response.Error != nil {
			t.Fatalf("Tools/call failed: %v", response.Error)
		}

		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Tools/call result is not an object")
		}

		content, ok := result["content"].([]interface{})
		if !ok {
			t.Fatal("Tools/call result missing content array")
		}

		if len(content) == 0 {
			t.Error("Expected at least one content item")
		}

		firstContent, ok := content[0].(map[string]interface{})
		if !ok {
			t.Fatal("First content item is not an object")
		}

		if firstContent["type"] != "text" {
			t.Errorf("Expected content type 'text', got %v", firstContent["type"])
		}

		text, ok := firstContent["text"].(string)
		if !ok {
			t.Fatal("Content text is not a string")
		}

		// Verify it's valid JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(text), &jsonData); err != nil {
			t.Errorf("Content text is not valid JSON: %v", err)
		}
	})

	t.Run("ToolsCall_GetASOverview", func(t *testing.T) {
		req := mcp.NewRequest("tools/call", map[string]interface{}{
			"name": "getASOverview",
			"arguments": map[string]interface{}{
				"resource": "15169",
			},
		}, 4)

		response := sendMCPRequest(t, mcpURL, req)

		if response.Error != nil {
			t.Fatalf("Tools/call failed: %v", response.Error)
		}

		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Tools/call result is not an object")
		}

		content, ok := result["content"].([]interface{})
		if !ok {
			t.Fatal("Tools/call result missing content array")
		}

		if len(content) == 0 {
			t.Error("Expected at least one content item")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		req := mcp.NewRequest("ping", nil, 5)
		response := sendMCPRequest(t, mcpURL, req)

		if response.Error != nil {
			t.Fatalf("Ping failed: %v", response.Error)
		}

		// Ping should return an empty object
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Ping result is not an object")
		}

		if len(result) != 0 {
			t.Errorf("Expected empty object for ping result, got %v", result)
		}
	})

	t.Run("MethodNotFound", func(t *testing.T) {
		req := mcp.NewRequest("nonexistent", nil, 6)
		response := sendMCPRequest(t, mcpURL, req)

		if response.Error == nil {
			t.Fatal("Expected error for nonexistent method")
		}

		if response.Error.Code != mcp.MethodNotFound {
			t.Errorf("Expected MethodNotFound error code %d, got %d", mcp.MethodNotFound, response.Error.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)

		resp, err := http.Post(mcpURL, "application/json", bytes.NewBuffer(invalidJSON))
		if err != nil {
			t.Fatalf("Failed to send invalid JSON: %v", err)
		}
		defer resp.Body.Close()

		var response mcp.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Error == nil {
			t.Fatal("Expected error for invalid JSON")
		}

		if response.Error.Code != mcp.ParseError {
			t.Errorf("Expected ParseError code %d, got %d", mcp.ParseError, response.Error.Code)
		}
	})
}

func sendMCPRequest(t *testing.T, url string, req *mcp.Request) *mcp.Response {
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response mcp.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return &response
}

func TestMCPConcurrency(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// First initialize the server
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	_ = sendMCPRequest(t, mcpURL, initReq)

	// Send initialized notification
	notif := mcp.NewNotification("initialized", nil)
	reqBody, _ := json.Marshal(notif)
	http.Post(mcpURL, "application/json", bytes.NewBuffer(reqBody))

	// Test concurrent requests
	numRequests := 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			req := mcp.NewRequest("tools/call", map[string]interface{}{
				"name": "getNetworkInfo",
				"arguments": map[string]interface{}{
					"resource": fmt.Sprintf("8.8.8.%d", (id%254)+1),
				},
			}, id+100)

			response := sendMCPRequest(t, mcpURL, req)
			if response.Error != nil {
				results <- fmt.Errorf("request %d failed: %v", id, response.Error)
			} else {
				results <- nil
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}
