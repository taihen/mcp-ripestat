//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestASPathLengthE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Prepare the MCP request for AS Path Length
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "test-as-path-length",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "getASPathLength",
			"arguments": map[string]interface{}{
				"resource": "AS3333",
			},
		},
	}

	// Send the request
	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(serverURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify the response structure
	if response["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got %v", response["jsonrpc"])
	}

	if response["id"] != "test-as-path-length" {
		t.Errorf("Expected id 'test-as-path-length', got %v", response["id"])
	}

	// Check if there's an error field
	if errField, exists := response["error"]; exists {
		t.Fatalf("Received error response: %v", errField)
	}

	// Verify the result structure
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object, got %T", response["result"])
	}

	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatalf("Expected content to be an array, got %T", result["content"])
	}

	if len(content) == 0 {
		t.Fatal("Expected at least one content item")
	}

	// Verify the first content item
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

	// Parse the JSON within the text field to verify AS Path Length data structure
	var asPathLengthData map[string]interface{}
	if err := json.Unmarshal([]byte(text), &asPathLengthData); err != nil {
		t.Fatalf("Failed to parse AS Path Length data: %v", err)
	}

	// Verify the AS Path Length response structure
	data, ok := asPathLengthData["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be an object, got %T", asPathLengthData["data"])
	}

	// Check for stats array
	stats, ok := data["stats"].([]interface{})
	if !ok {
		t.Fatalf("Expected stats to be an array, got %T", data["stats"])
	}

	if len(stats) == 0 {
		t.Fatal("Expected at least one stat entry")
	}

	// Verify the first stat entry structure
	firstStat, ok := stats[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first stat to be an object, got %T", stats[0])
	}

	// Verify required fields in stat entry
	requiredFields := []string{"number", "count", "location", "stripped", "unstripped"}
	for _, field := range requiredFields {
		if _, exists := firstStat[field]; !exists {
			t.Errorf("Expected field '%s' in stat entry", field)
		}
	}

	// Verify stripped and unstripped path stats structure
	for _, pathType := range []string{"stripped", "unstripped"} {
		pathStats, ok := firstStat[pathType].(map[string]interface{})
		if !ok {
			t.Errorf("Expected %s to be an object, got %T", pathType, firstStat[pathType])
			continue
		}

		pathStatsFields := []string{"sum", "min", "max", "avg"}
		for _, field := range pathStatsFields {
			if _, exists := pathStats[field]; !exists {
				t.Errorf("Expected field '%s' in %s path stats", field, pathType)
			}
		}
	}

	// Verify resource in data
	if resource, ok := data["resource"].(string); !ok || resource != "3333" {
		t.Errorf("Expected resource '3333', got %v", data["resource"])
	}

	// Verify query_time exists
	if _, exists := data["query_time"]; !exists {
		t.Error("Expected query_time field in data")
	}

	// Verify status and status_code
	if status, ok := asPathLengthData["status"].(string); !ok || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", asPathLengthData["status"])
	}

	if statusCode, ok := asPathLengthData["status_code"].(float64); !ok || statusCode != 200 {
		t.Errorf("Expected status_code 200, got %v", asPathLengthData["status_code"])
	}

	fmt.Printf("AS Path Length E2E test passed successfully with %d stat entries\n", len(stats))
}
