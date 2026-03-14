//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/taihen/mcp-ripestat/internal/mcp"
)

func TestASPathLengthE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mcpURL := serverURL + "/mcp"

	// Initialize a dedicated session for this test.
	initReq := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}, "aspath-init")
	initResp, initResult := sendMCPRequestWithHeaders(t, mcpURL, initReq, nil)
	if initResult.Error != nil {
		t.Fatalf("Initialize failed: %v", initResult.Error)
	}
	sessionID := initResp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("Expected session ID from initialize response")
	}

	// AS path length is now reached via consolidated analyzeRouting tool.
	req := mcp.NewRequest("tools/call", map[string]interface{}{
		"name": "analyzeRouting",
		"arguments": map[string]interface{}{
			"resource": "AS3333",
			"analysis": []string{"path-optimization"},
		},
	}, "test-as-path-length")
	_, response := sendMCPRequestWithSession(t, mcpURL, req, sessionID)

	// Check if there's an error field
	if response.Error != nil {
		t.Fatalf("Received error response: %v", response.Error)
	}

	// Verify the result structure
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object, got %T", response.Result)
	}

	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatalf("Expected content to be an array, got %T", result["content"])
	}
	if len(content) == 0 {
		t.Fatal("Expected at least one content item")
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first content to be an object, got %T", content[0])
	}
	if firstContent["type"] != "text" {
		t.Errorf("Expected content type 'text', got %v", firstContent["type"])
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		t.Fatalf("Expected text to be a string, got %T", firstContent["text"])
	}

	// Consolidated tool payload should contain endpoint-level results.
	var consolidated map[string]interface{}
	if err := json.Unmarshal([]byte(text), &consolidated); err != nil {
		t.Fatalf("Failed to parse consolidated response: %v", err)
	}
	results, ok := consolidated["results"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected consolidated results map, got %T", consolidated["results"])
	}
	if _, exists := results["getASPathLength"]; !exists {
		t.Fatalf("Expected getASPathLength result in consolidated payload")
	}

	// Also ensure no endpoint-level errors are present for this path.
	if errs, exists := consolidated["errors"].(map[string]interface{}); exists && len(errs) > 0 {
		t.Fatalf("Expected no endpoint errors, got %v", errs)
	}

	fmt.Printf("AS Path Length E2E test passed successfully with consolidated routing analysis\n")
}
