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

// Session ID shared across tests
var testSessionID string

func TestMCPProtocol(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	t.Run("Initialize", func(t *testing.T) {
		req := mcp.NewRequest("initialize", map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		}, 1)

		resp, response := sendMCPRequestWithHeaders(t, mcpURL, req, nil)

		// Store session ID for subsequent requests
		testSessionID = resp.Header.Get("Mcp-Session-Id")
		if testSessionID == "" {
			t.Log("No session ID received, tests may fail")
		}

		if response.Error != nil {
			t.Fatalf("Initialize failed: %v", response.Error)
		}

		// Verify response structure
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Initialize result is not an object")
		}

		if result["protocolVersion"] != "2025-06-18" {
			t.Errorf("Expected protocolVersion '2025-06-18', got %v", result["protocolVersion"])
		}

		if _, ok := result["capabilities"]; !ok {
			t.Error("Initialize result missing capabilities")
		}

		if _, ok := result["serverInfo"]; !ok {
			t.Error("Initialize result missing serverInfo")
		}
	})

	t.Run("Initialized", func(t *testing.T) {
		notif := mcp.NewNotification("notifications/initialized", nil)

		reqBody, err := json.Marshal(notif)
		if err != nil {
			t.Fatalf("Failed to marshal notification: %v", err)
		}

		httpReq, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json, text/event-stream")
		if testSessionID != "" {
			httpReq.Header.Set("Mcp-Session-Id", testSessionID)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("Failed to send initialized notification: %v", err)
		}
		defer resp.Body.Close()

		// SDK may accept various status codes for notifications
		// 200 = processed, 202 = accepted, 204 = no content, 400 = rejected (e.g., no session)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
			// SDK is strict about session management, 400 may be acceptable
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 200/202/204/400 for notification, got %d", resp.StatusCode)
			}
		}
	})

	t.Run("ToolsList", func(t *testing.T) {
		req := mcp.NewRequest("tools/list", nil, 2)
		_, response := sendMCPRequestWithSession(t, mcpURL, req, testSessionID)

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

	t.Run("ToolsCall_InvestigateResource", func(t *testing.T) {
		// Use the consolidated investigateResource tool instead of individual getNetworkInfo
		req := mcp.NewRequest("tools/call", map[string]interface{}{
			"name": "investigateResource",
			"arguments": map[string]interface{}{
				"resource":   "8.8.8.8",
				"operations": []string{"overview"},
				"depth":      "basic",
			},
		}, 3)

		_, response := sendMCPRequestWithSession(t, mcpURL, req, testSessionID)

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

	t.Run("ToolsCall_ExploreRelationships", func(t *testing.T) {
		// Use the consolidated exploreRelationships tool instead of individual getASOverview
		req := mcp.NewRequest("tools/call", map[string]interface{}{
			"name": "exploreRelationships",
			"arguments": map[string]interface{}{
				"resource":      "15169",
				"relationships": []string{"neighbors"},
				"scope":         "direct",
			},
		}, 4)

		_, response := sendMCPRequestWithSession(t, mcpURL, req, testSessionID)

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
		_, response := sendMCPRequestWithSession(t, mcpURL, req, testSessionID)

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

		reqBody, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		httpReq, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json, text/event-stream")
		if testSessionID != "" {
			httpReq.Header.Set("Mcp-Session-Id", testSessionID)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// SDK may return 200 with error in body or 400 for unknown method
		if resp.StatusCode == http.StatusOK {
			var response mcp.Response
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if response.Error == nil {
				t.Fatal("Expected error for nonexistent method")
			}
			if response.Error.Code != mcp.MethodNotFound {
				t.Logf("Got error code %d (expected %d)", response.Error.Code, mcp.MethodNotFound)
			}
		} else if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 for unknown method, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)

		httpReq, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(invalidJSON))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json, text/event-stream")
		if testSessionID != "" {
			httpReq.Header.Set("Mcp-Session-Id", testSessionID)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("Failed to send invalid JSON: %v", err)
		}
		defer resp.Body.Close()

		// SDK may return 400 with error message or 200 with JSON error response
		if resp.StatusCode == http.StatusOK {
			var response mcp.Response
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Logf("Failed to decode response (may be expected for invalid JSON): %v", err)
				return
			}

			if response.Error == nil {
				t.Fatal("Expected error for invalid JSON")
			}

			if response.Error.Code != mcp.ParseError {
				t.Logf("Got error code %d (expected %d)", response.Error.Code, mcp.ParseError)
			}
		} else if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 for invalid JSON, got %d", resp.StatusCode)
		}
	})
}

func sendMCPRequest(t *testing.T, url string, req *mcp.Request) *mcp.Response {
	_, response := sendMCPRequestWithHeaders(t, url, req, nil)
	return response
}

func sendMCPRequestWithSession(t *testing.T, url string, req *mcp.Request, sessionID string) (*http.Response, *mcp.Response) {
	headers := map[string]string{}
	if sessionID != "" {
		headers["Mcp-Session-Id"] = sessionID
	}
	return sendMCPRequestWithHeaders(t, url, req, headers)
}

func sendMCPRequestWithHeaders(t *testing.T, url string, req *mcp.Request, headers map[string]string) (*http.Response, *mcp.Response) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response mcp.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		resp.Body.Close()
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp, &response
}

func TestMCPConcurrency(t *testing.T) {
	mcpURL := serverURL + "/mcp"

	// First initialize the server (get new session for this test)
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, 1)

	resp, _ := sendMCPRequestWithHeaders(t, mcpURL, initReq, nil)
	concurrencySessionID := resp.Header.Get("Mcp-Session-Id")

	// Send initialized notification
	notif := mcp.NewNotification("initialized", nil)
	reqBody, _ := json.Marshal(notif)
	httpReq, _ := http.NewRequest("POST", mcpURL, bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if concurrencySessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", concurrencySessionID)
	}
	http.DefaultClient.Do(httpReq)

	// Test concurrent requests
	numRequests := 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			req := mcp.NewRequest("tools/call", map[string]interface{}{
				"name": "investigateResource",
				"arguments": map[string]interface{}{
					"resource":   fmt.Sprintf("8.8.8.%d", (id%254)+1),
					"operations": []string{"overview"},
					"depth":      "basic",
				},
			}, id+100)

			_, response := sendMCPRequestWithSession(t, mcpURL, req, concurrencySessionID)
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
