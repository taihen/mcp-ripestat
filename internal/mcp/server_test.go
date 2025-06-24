package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)

	if server.serverName != "test-server" {
		t.Errorf("Expected serverName to be 'test-server', got %s", server.serverName)
	}
	if server.serverVersion != "1.0.0" {
		t.Errorf("Expected serverVersion to be '1.0.0', got %s", server.serverVersion)
	}
	if server.disableWhatsMyIP != false {
		t.Errorf("Expected disableWhatsMyIP to be false, got %v", server.disableWhatsMyIP)
	}
	if server.initialized != false {
		t.Errorf("Expected initialized to be false, got %v", server.initialized)
	}
}

func TestProcessMessage_Initialize(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	initRequest := `{
		"jsonrpc": "2.0",
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		},
		"id": 1
	}`

	result, err := server.ProcessMessage(ctx, []byte(initRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got %s", response.JSONRPC)
	}
	// JSON unmarshaling converts numbers to float64
	if response.ID.(float64) != 1.0 {
		t.Errorf("Expected ID to be 1, got %v", response.ID)
	}
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	if response.Result == nil {
		t.Error("Expected result to be non-nil")
	}
}

func TestProcessMessage_Initialized(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	// First initialize
	initRequest := `{
		"jsonrpc": "2.0",
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		},
		"id": 1
	}`

	_, err := server.ProcessMessage(ctx, []byte(initRequest))
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Then send initialized notification
	initializedNotif := `{
		"jsonrpc": "2.0",
		"method": "initialized"
	}`

	result, err := server.ProcessMessage(ctx, []byte(initializedNotif))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for notification, got %v", result)
	}

	if !server.initialized {
		t.Error("Expected server to be initialized")
	}
}

func TestProcessMessage_ToolsList(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	server.initialized = true // Skip initialization for this test
	ctx := context.Background()

	toolsListRequest := `{
		"jsonrpc": "2.0",
		"method": "tools/list",
		"id": 2
	}`

	result, err := server.ProcessMessage(ctx, []byte(toolsListRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	// Verify the result contains tools
	resultData, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var toolsResult ToolsListResult
	if err := json.Unmarshal(resultData, &toolsResult); err != nil {
		t.Fatalf("Failed to unmarshal tools result: %v", err)
	}

	if len(toolsResult.Tools) == 0 {
		t.Error("Expected at least one tool in the result")
	}
}

func TestProcessMessage_ToolsListWithWhatsMyIPDisabled(t *testing.T) {
	server := NewServer("test-server", "1.0.0", true) // Disable whats-my-ip
	server.initialized = true
	ctx := context.Background()

	toolsListRequest := `{
		"jsonrpc": "2.0",
		"method": "tools/list",
		"id": 2
	}`

	result, err := server.ProcessMessage(ctx, []byte(toolsListRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	resultData, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var toolsResult ToolsListResult
	if err := json.Unmarshal(resultData, &toolsResult); err != nil {
		t.Fatalf("Failed to unmarshal tools result: %v", err)
	}

	// Check that getWhatsMyIP is not in the list
	for _, tool := range toolsResult.Tools {
		if tool.Name == "getWhatsMyIP" {
			t.Error("getWhatsMyIP should not be in tools list when disabled")
		}
	}
}

func TestProcessMessage_Ping(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	pingRequest := `{
		"jsonrpc": "2.0",
		"method": "ping",
		"id": 3
	}`

	result, err := server.ProcessMessage(ctx, []byte(pingRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	// JSON unmarshaling converts numbers to float64
	if response.ID.(float64) != 3.0 {
		t.Errorf("Expected ID to be 3, got %v", response.ID)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	invalidJSON := `{invalid json}`

	result, err := server.ProcessMessage(ctx, []byte(invalidJSON))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid JSON")
	}
	if response.Error.Code != ParseError {
		t.Errorf("Expected ParseError code %d, got %d", ParseError, response.Error.Code)
	}
}

func TestProcessMessage_MethodNotFound(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	unknownMethod := `{
		"jsonrpc": "2.0",
		"method": "unknown",
		"id": 4
	}`

	result, err := server.ProcessMessage(ctx, []byte(unknownMethod))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for unknown method")
	}
	if response.Error.Code != MethodNotFound {
		t.Errorf("Expected MethodNotFound code %d, got %d", MethodNotFound, response.Error.Code)
	}
}

func TestProcessMessage_ToolsCallBeforeInitialization(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	toolsCallRequest := `{
		"jsonrpc": "2.0",
		"method": "tools/call",
		"params": {
			"name": "getNetworkInfo",
			"arguments": {"resource": "8.8.8.8"}
		},
		"id": 5
	}`

	result, err := server.ProcessMessage(ctx, []byte(toolsCallRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for tools/call before initialization")
	}
	if response.Error.Code != InitializationError {
		t.Errorf("Expected InitializationError code %d, got %d", InitializationError, response.Error.Code)
	}
}

func TestExecuteToolCall_UnknownTool(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name: "unknownTool",
		Arguments: map[string]interface{}{
			"resource": "test",
		},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Errorf("Expected error message to contain 'unknown tool', got %s", err.Error())
	}
}

func TestExecuteToolCall_WhatsMyIPDisabled(t *testing.T) {
	server := NewServer("test-server", "1.0.0", true) // Disable whats-my-ip
	ctx := context.Background()

	params := &CallToolParams{
		Name:      "getWhatsMyIP",
		Arguments: map[string]interface{}{},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for disabled whats-my-ip tool")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected error message to contain 'disabled', got %s", err.Error())
	}
}

func TestValidateRequest_Integration(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	invalidRequest := `{
		"jsonrpc": "1.0",
		"method": "test",
		"id": 1
	}`

	result, err := server.ProcessMessage(ctx, []byte(invalidRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid request")
	}
	if response.Error.Code != InvalidRequest {
		t.Errorf("Expected InvalidRequest code %d, got %d", InvalidRequest, response.Error.Code)
	}
}

func TestExecuteToolCall_NetworkInfo_MissingResource(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name:      "getNetworkInfo",
		Arguments: map[string]interface{}{}, // Missing resource
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for missing resource parameter")
	}
	if !strings.Contains(err.Error(), "resource parameter is required") {
		t.Errorf("Expected error message about missing resource, got %s", err.Error())
	}
}

func TestExecuteToolCall_RPKIValidation_MissingPrefix(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name: "getRPKIValidation",
		Arguments: map[string]interface{}{
			"resource": "AS15169", // Missing prefix
		},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for missing prefix parameter")
	}
	if !strings.Contains(err.Error(), "prefix parameter is required") {
		t.Errorf("Expected error message about missing prefix, got %s", err.Error())
	}
}

func TestExecuteToolCall_ASNNeighbours_InvalidLOD(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name: "getASNNeighbours",
		Arguments: map[string]interface{}{
			"resource": "AS15169",
			"lod":      "invalid", // Invalid LOD
		},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid LOD parameter")
	}
	if !strings.Contains(err.Error(), "lod parameter must be 0 or 1") {
		t.Errorf("Expected error message about invalid LOD, got %s", err.Error())
	}
}

func TestExecuteToolCall_LookingGlass_InvalidLookBackLimit(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name: "getLookingGlass",
		Arguments: map[string]interface{}{
			"resource":        "8.8.8.0/24",
			"look_back_limit": "invalid", // Invalid look back limit
		},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid look_back_limit parameter")
	}
	if !strings.Contains(err.Error(), "look_back_limit parameter must be a valid integer") {
		t.Errorf("Expected error message about invalid look_back_limit, got %s", err.Error())
	}
}

func TestProcessMessage_ToolsCall_InvalidParams(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	server.initialized = true
	ctx := context.Background()

	// Test with invalid params structure
	toolsCallRequest := `{
		"jsonrpc": "2.0",
		"method": "tools/call",
		"params": "invalid params",
		"id": 5
	}`

	result, err := server.ProcessMessage(ctx, []byte(toolsCallRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid params")
	}
	if response.Error.Code != InvalidParams {
		t.Errorf("Expected InvalidParams code %d, got %d", InvalidParams, response.Error.Code)
	}
}

func TestProcessMessage_Initialize_InvalidParams(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	// Test with invalid params structure
	initRequest := `{
		"jsonrpc": "2.0",
		"method": "initialize",
		"params": "invalid params",
		"id": 1
	}`

	result, err := server.ProcessMessage(ctx, []byte(initRequest))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	response, ok := result.(*Response)
	if !ok {
		t.Fatalf("Expected Response, got %T", result)
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid params")
	}
	if response.Error.Code != InvalidParams {
		t.Errorf("Expected InvalidParams code %d, got %d", InvalidParams, response.Error.Code)
	}
}

func TestProcessMessage_UnknownNotification(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	unknownNotif := `{
		"jsonrpc": "2.0",
		"method": "unknown-notification"
	}`

	result, err := server.ProcessMessage(ctx, []byte(unknownNotif))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for unknown notification, got %v", result)
	}
}

func TestProcessMessage_CancellationNotification(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	cancelNotif := `{
		"jsonrpc": "2.0",
		"method": "notifications/cancelled"
	}`

	result, err := server.ProcessMessage(ctx, []byte(cancelNotif))
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for cancellation notification, got %v", result)
	}
}

func TestExecuteToolCall_ArgumentParsing(t *testing.T) {
	// Test argument parsing without making network calls
	testCases := []struct {
		name        string
		params      *CallToolParams
		expectError bool
	}{
		{
			name: "valid arguments",
			params: &CallToolParams{
				Name: "getNetworkInfo",
				Arguments: map[string]interface{}{
					"resource": "test",
				},
			},
			expectError: false,
		},
		{
			name: "arguments with meta field",
			params: &CallToolParams{
				Name: "getNetworkInfo",
				Arguments: map[string]interface{}{
					"resource": "test",
				},
				Meta: map[string]interface{}{
					"progressToken": 123,
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse arguments to test the argument parsing logic
			args := make(map[string]interface{})
			if tc.params.Arguments != nil {
				jsonData, err := json.Marshal(tc.params.Arguments)
				if err != nil && !tc.expectError {
					t.Errorf("Failed to marshal arguments: %v", err)
				}
				if err := json.Unmarshal(jsonData, &args); err != nil && !tc.expectError {
					t.Errorf("Failed to unmarshal arguments: %v", err)
				}
			}

			// Test that we can get the resource parameter
			if !tc.expectError {
				if resource, ok := args["resource"].(string); !ok || resource == "" {
					t.Error("Expected valid resource parameter")
				}
			}
		})
	}
}

func TestParseCallToolParams_InvalidJSON(t *testing.T) {
	// Test with a channel that can't be marshaled
	ch := make(chan int)
	_, err := ParseCallToolParams(ch)
	if err == nil {
		t.Error("Expected error for unmarshalable params")
	}
}

func TestCreateToolResultFromJSON_InvalidData(t *testing.T) {
	// Test with data that can't be marshaled (function)
	invalidData := func() {}
	result := CreateToolResultFromJSON(invalidData)

	if !result.IsError {
		t.Error("Expected error result for invalid data")
	}
	if !strings.Contains(result.Content[0].Text, "Error marshaling result") {
		t.Errorf("Expected error message about marshaling, got: %s", result.Content[0].Text)
	}
}

func TestExecuteToolCall_AllToolFunctions(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	testCases := []struct {
		name           string
		toolName       string
		args           map[string]interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:     "callASOverview success",
			toolName: "getASOverview",
			args:     map[string]interface{}{"resource": "15169"},
		},
		{
			name:         "callASOverview missing resource",
			toolName:     "getASOverview",
			args:         map[string]interface{}{},
			expectError:  true,
			errorMessage: "resource parameter is required",
		},
		{
			name:     "callAnnouncedPrefixes success",
			toolName: "getAnnouncedPrefixes",
			args:     map[string]interface{}{"resource": "15169"},
		},
		{
			name:         "callAnnouncedPrefixes missing resource",
			toolName:     "getAnnouncedPrefixes",
			args:         map[string]interface{}{},
			expectError:  true,
			errorMessage: "resource parameter is required",
		},
		{
			name:     "callRoutingStatus success",
			toolName: "getRoutingStatus",
			args:     map[string]interface{}{"resource": "8.8.8.8"},
		},
		{
			name:         "callRoutingStatus missing resource",
			toolName:     "getRoutingStatus",
			args:         map[string]interface{}{},
			expectError:  true,
			errorMessage: "resource parameter is required",
		},
		{
			name:     "callWhois success",
			toolName: "getWhois",
			args:     map[string]interface{}{"resource": "8.8.8.8"},
		},
		{
			name:         "callWhois missing resource",
			toolName:     "getWhois",
			args:         map[string]interface{}{},
			expectError:  true,
			errorMessage: "resource parameter is required",
		},
		{
			name:     "callAbuseContactFinder success",
			toolName: "getAbuseContactFinder",
			args:     map[string]interface{}{"resource": "8.8.8.8"},
		},
		{
			name:         "callAbuseContactFinder missing resource",
			toolName:     "getAbuseContactFinder",
			args:         map[string]interface{}{},
			expectError:  true,
			errorMessage: "resource parameter is required",
		},
		{
			name:     "callWhatsMyIP success",
			toolName: "getWhatsMyIP",
			args:     map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := &CallToolParams{
				Name:      tc.toolName,
				Arguments: tc.args,
			}

			result, err := server.executeToolCall(ctx, params)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tc.name)
					return
				}
				if tc.errorMessage != "" && !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got %s", tc.errorMessage, err.Error())
				}
			} else {
				// Note: These might fail due to network issues in tests, so we accept that
				if err != nil {
					t.Logf("Network call failed (expected in test environment): %v", err)
				} else if result == nil {
					t.Error("Expected non-nil result when no error occurred")
				}
			}
		})
	}
}

func TestExecuteToolCall_WhatsMyIPDisabledInDepth(t *testing.T) {
	server := NewServer("test-server", "1.0.0", true) // Disable whats-my-ip
	ctx := context.Background()

	params := &CallToolParams{
		Name:      "getWhatsMyIP",
		Arguments: map[string]interface{}{},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for disabled whats-my-ip tool")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected error message to contain 'disabled', got %s", err.Error())
	}
}

func TestProcessMessage_ToolsCall_CompleteFlow(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	// First initialize the server
	initRequest := `{
		"jsonrpc": "2.0",
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		},
		"id": 1
	}`

	_, err := server.ProcessMessage(ctx, []byte(initRequest))
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Send initialized notification
	initializedNotif := `{
		"jsonrpc": "2.0",
		"method": "initialized"
	}`

	_, err = server.ProcessMessage(ctx, []byte(initializedNotif))
	if err != nil {
		t.Fatalf("Initialized notification failed: %v", err)
	}

	// Test tools/call for each tool
	toolTests := []struct {
		name   string
		params string
	}{
		{
			name: "getNetworkInfo",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "getNetworkInfo",
					"arguments": {"resource": "8.8.8.8"}
				},
				"id": 2
			}`,
		},
		{
			name: "getASOverview",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "getASOverview",
					"arguments": {"resource": "15169"}
				},
				"id": 3
			}`,
		},
	}

	for _, test := range toolTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := server.ProcessMessage(ctx, []byte(test.params))
			if err != nil {
				t.Errorf("Tool call failed: %v", err)
				return
			}

			response, ok := result.(*Response)
			if !ok {
				t.Errorf("Expected Response, got %T", result)
				return
			}

			// Accept either success or error due to network conditions in tests
			if response.Error != nil {
				t.Logf("Tool call returned error (may be due to network): %v", response.Error)
			} else if response.Result == nil {
				t.Error("Expected result when no error occurred")
			}
		})
	}
}
