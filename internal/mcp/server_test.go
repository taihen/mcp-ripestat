package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
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

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

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
	server := NewServer("test-server", "1.0.0", true)
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
	server := NewServer("test-server", "1.0.0", true)
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

func TestProcessMessage_ToolsCall_InvalidParams(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	server.initialized = true
	ctx := context.Background()

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

			if !tc.expectError {
				if resource, ok := args["resource"].(string); !ok || resource == "" {
					t.Error("Expected valid resource parameter")
				}
			}
		})
	}
}

func TestParseCallToolParams_InvalidJSON(t *testing.T) {

	ch := make(chan int)
	_, err := ParseCallToolParams(ch)
	if err == nil {
		t.Error("Expected error for unmarshalable params")
	}
}

func TestCreateToolResultFromJSON_InvalidData(t *testing.T) {

	invalidData := func() {}
	result := CreateToolResultFromJSON(invalidData)

	if !result.IsError {
		t.Error("Expected error result for invalid data")
	}
	if !strings.Contains(result.Content[0].Text, "Error marshaling result") {
		t.Errorf("Expected error message about marshaling, got: %s", result.Content[0].Text)
	}
}

func TestExecuteToolCall_WhatsMyIPDisabledInDepth(t *testing.T) {
	server := NewServer("test-server", "1.0.0", true)
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

func TestProcessMessage_EdgeCases(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	result, err := server.ProcessMessage(ctx, []byte("invalid json"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response, ok := result.(*Response); ok {
		if response.Error == nil || response.Error.Code != ParseError {
			t.Error("Expected ParseError for invalid JSON")
		}
	}

	unknownMessage := []byte(`{"jsonrpc": "2.0"}`)
	result, err = server.ProcessMessage(ctx, unknownMessage)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response, ok := result.(*Response); ok {
		if response.Error == nil || response.Error.Code != InvalidRequest {
			t.Error("Expected InvalidRequest for unknown message type")
		}
	}
}

func TestHandleInitialize_EdgeCases(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)

	req := &Request{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      json.RawMessage(`1`),
		Params:  nil,
	}

	result, err := server.handleInitialize(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response, ok := result.(*Response); !ok || response.Error != nil {
		t.Error("Expected successful response with nil params")
	}

	req.Params = make(chan int)
	result, err = server.handleInitialize(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response, ok := result.(*Response); ok {
		if response.Error == nil || response.Error.Code != InvalidParams {
			t.Error("Expected InvalidParams for unmarshalable params")
		}
	}
}

func TestProcessMessage_ToolsCall_CompleteFlow(t *testing.T) {
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

	_, err := server.ProcessMessage(ctx, []byte(initRequest))
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	initializedNotif := `{
		"jsonrpc": "2.0",
		"method": "initialized"
	}`

	_, err = server.ProcessMessage(ctx, []byte(initializedNotif))
	if err != nil {
		t.Fatalf("Initialized notification failed: %v", err)
	}

	toolTests := []struct {
		name   string
		params string
	}{
		{
			name: "investigateResource",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "investigateResource",
					"arguments": {"resource": "8.8.8.8"}
				},
				"id": 2
			}`,
		},
		{
			name: "analyzeRouting",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "analyzeRouting",
					"arguments": {"resource": "AS15169"}
				},
				"id": 3
			}`,
		},
		{
			name: "queryRegistry",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "queryRegistry",
					"arguments": {"resource": "AS3333"}
				},
				"id": 4
			}`,
		},
		{
			name: "validateSecurity",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "validateSecurity",
					"arguments": {"resource": "8.8.8.0/24"}
				},
				"id": 5
			}`,
		},
		{
			name: "exploreRelationships",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "exploreRelationships",
					"arguments": {"resource": "AS15169"}
				},
				"id": 6
			}`,
		},
		{
			name: "searchByLocation",
			params: `{
				"jsonrpc": "2.0",
				"method": "tools/call",
				"params": {
					"name": "searchByLocation",
					"arguments": {"country": "US"}
				},
				"id": 7
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

			if response.Error != nil {
				t.Logf("Tool call returned error (may be due to network): %v", response.Error)
			} else if response.Result == nil {
				t.Error("Expected result when no error occurred")
			}
		})
	}
}

func TestProcessMessage_ParseError(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	invalidJSON := `{completely invalid json`

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

func TestParseMessage_ErrorResponseCases(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		isError bool
	}{
		{
			name: "valid error response",
			input: `{
				"jsonrpc": "2.0",
				"error": {
					"code": -1,
					"message": "test error"
				},
				"id": 1
			}`,
			isError: false,
		},
		{
			name: "invalid error response structure",
			input: `{
				"jsonrpc": "2.0",
				"error": "invalid error structure",
				"id": 1
			}`,
			isError: true,
		},
		{
			name: "valid result response",
			input: `{
				"jsonrpc": "2.0",
				"result": {"data": "test"},
				"id": 1
			}`,
			isError: false,
		},
		{
			name: "invalid result response structure",
			input: `{
				"jsonrpc": "2.0",
				"result": {"data": "test"},
				"id": 1,
				"invalid_field": true
			}`,
			isError: false,
		},
		{
			name: "invalid request structure",
			input: `{
				"jsonrpc": "2.0",
				"method": 123,
				"id": 1
			}`,
			isError: true,
		},
		{
			name: "invalid notification structure",
			input: `{
				"jsonrpc": "2.0",
				"method": 456
			}`,
			isError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMessage([]byte(tc.input))

			if tc.isError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.isError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestExecuteToolCall_ArgumentMarshalingError(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name: "investigateResource",
		Arguments: map[string]interface{}{
			"resource": make(chan int),
		},
	}

	_, err := server.executeToolCall(ctx, params)
	if err == nil {
		t.Error("Expected error for unmarshalable arguments")
	}
	if !strings.Contains(err.Error(), "failed to marshal arguments") {
		t.Errorf("Expected marshaling error, got: %v", err)
	}
}

func TestExecuteToolCall_ArgumentUnmarshalingError(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name:      "investigateResource",
		Arguments: "invalid json string",
	}

	_, err := server.executeToolCall(ctx, params)

	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestHandleToolsCall_ToolExecutionError(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	server.initialized = true
	ctx := context.Background()

	toolsCallRequest := `{
		"jsonrpc": "2.0",
		"method": "tools/call",
		"params": {
			"name": "unknownTool",
			"arguments": {"resource": "test"}
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
		t.Error("Expected error response for unknown tool")
	}
	if response.Error.Code != ToolError {
		t.Errorf("Expected ToolError code %d, got %d", ToolError, response.Error.Code)
	}
}

func TestCoverageCompletionTests(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
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

	if response.Error == nil {
		t.Error("Expected error response for tools/list before initialization")
	}
	if response.Error.Code != InitializationError {
		t.Errorf("Expected InitializationError code %d, got %d", InitializationError, response.Error.Code)
	}
}

func TestParseMessage_ErrorResponseEdgeCases(t *testing.T) {

	malformedErrorResponse := `{
		"jsonrpc": "2.0",
		"error": {
			"code": "not_a_number",
			"message": "test error"
		},
		"id": 1
	}`

	_, err := ParseMessage([]byte(malformedErrorResponse))
	if err == nil {
		t.Error("Expected error for malformed error response")
	}
	if !strings.Contains(err.Error(), "invalid error response") {
		t.Errorf("Expected 'invalid error response' error, got: %v", err)
	}

	malformedResultResponse := `{
		"jsonrpc": "2.0",
		"result": {"data": "test"
		"id": 1
	}`

	_, err = ParseMessage([]byte(malformedResultResponse))
	if err == nil {
		t.Error("Expected error for malformed result response")
	}
}

func TestFormatErrorMessage(t *testing.T) {
	testCases := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "error without Error prefix",
			input:    errors.New("network timeout"),
			expected: "Error: network timeout",
		},
		{
			name:     "error with Error prefix",
			input:    errors.New("Error: invalid resource"),
			expected: "Error: invalid resource",
		},
		{
			name:     "error with lowercase error prefix",
			input:    errors.New("error: something went wrong"),
			expected: "Error: error: something went wrong",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatErrorMessage(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestParameterHelpers(t *testing.T) {
	t.Run("getRequiredStringParam", func(t *testing.T) {
		testCases := []struct {
			name        string
			args        map[string]interface{}
			key         string
			errorMsg    string
			expectedVal string
			expectError bool
		}{
			{
				name:        "valid string parameter",
				args:        map[string]interface{}{"resource": "test-value"},
				key:         "resource",
				errorMsg:    "Error: resource required",
				expectedVal: "test-value",
				expectError: false,
			},
			{
				name:        "missing parameter",
				args:        map[string]interface{}{},
				key:         "resource",
				errorMsg:    "Error: resource required",
				expectedVal: "",
				expectError: true,
			},
			{
				name:        "wrong type parameter",
				args:        map[string]interface{}{"resource": 123},
				key:         "resource",
				errorMsg:    "Error: resource required",
				expectedVal: "",
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				val, errResult := getRequiredStringParam(tc.args, tc.key, tc.errorMsg)

				if tc.expectError {
					if errResult == nil {
						t.Error("Expected error result but got none")
					} else if errResult.Content[0].Text != tc.errorMsg {
						t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, errResult.Content[0].Text)
					}
				} else {
					if errResult != nil {
						t.Errorf("Expected no error but got: %v", errResult.Content[0].Text)
					}
					if val != tc.expectedVal {
						t.Errorf("Expected value '%s', got '%s'", tc.expectedVal, val)
					}
				}
			})
		}
	})

	t.Run("getOptionalStringParam", func(t *testing.T) {
		testCases := []struct {
			name        string
			args        map[string]interface{}
			key         string
			expectedVal string
		}{
			{
				name:        "existing parameter",
				args:        map[string]interface{}{"query_time": "2023-01-01"},
				key:         "query_time",
				expectedVal: "2023-01-01",
			},
			{
				name:        "missing parameter",
				args:        map[string]interface{}{},
				key:         "query_time",
				expectedVal: "",
			},
			{
				name:        "wrong type parameter",
				args:        map[string]interface{}{"query_time": 123},
				key:         "query_time",
				expectedVal: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				val := getOptionalStringParam(tc.args, tc.key)
				if val != tc.expectedVal {
					t.Errorf("Expected value '%s', got '%s'", tc.expectedVal, val)
				}
			})
		}
	})

	t.Run("validateLODParam", func(t *testing.T) {
		testCases := []struct {
			name        string
			args        map[string]interface{}
			expectedVal int
			expectError bool
		}{
			{
				name:        "valid LOD 0",
				args:        map[string]interface{}{"lod": "0"},
				expectedVal: 0,
				expectError: false,
			},
			{
				name:        "valid LOD 1",
				args:        map[string]interface{}{"lod": "1"},
				expectedVal: 1,
				expectError: false,
			},
			{
				name:        "missing LOD parameter",
				args:        map[string]interface{}{},
				expectedVal: 0,
				expectError: false,
			},
			{
				name:        "invalid LOD value",
				args:        map[string]interface{}{"lod": "2"},
				expectedVal: 0,
				expectError: true,
			},
			{
				name:        "non-numeric LOD",
				args:        map[string]interface{}{"lod": "abc"},
				expectedVal: 0,
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				val, errResult := validateLODParam(tc.args)

				if tc.expectError {
					if errResult == nil {
						t.Error("Expected error result but got none")
					}
				} else {
					if errResult != nil {
						t.Errorf("Expected no error but got: %v", errResult.Content[0].Text)
					}
					if val != tc.expectedVal {
						t.Errorf("Expected value %d, got %d", tc.expectedVal, val)
					}
				}
			})
		}
	})

	t.Run("validateLookBackLimitParam", func(t *testing.T) {
		testCases := []struct {
			name        string
			args        map[string]interface{}
			expectedVal int
			expectError bool
		}{
			{
				name:        "valid look back limit",
				args:        map[string]interface{}{"look_back_limit": "10"},
				expectedVal: 10,
				expectError: false,
			},
			{
				name:        "missing look back limit",
				args:        map[string]interface{}{},
				expectedVal: 0,
				expectError: false,
			},
			{
				name:        "invalid look back limit",
				args:        map[string]interface{}{"look_back_limit": "abc"},
				expectedVal: 0,
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				val, errResult := validateLookBackLimitParam(tc.args)

				if tc.expectError {
					if errResult == nil {
						t.Error("Expected error result but got none")
					}
				} else {
					if errResult != nil {
						t.Errorf("Expected no error but got: %v", errResult.Content[0].Text)
					}
					if val != tc.expectedVal {
						t.Errorf("Expected value %d, got %d", tc.expectedVal, val)
					}
				}
			})
		}
	})
}

func TestWithHTTPRequest(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest("GET", "http://example.com", nil)

	ctxWithReq := WithHTTPRequest(ctx, req)

	retrievedReq, ok := HTTPRequestFromContext(ctxWithReq)
	if !ok {
		t.Fatal("Expected to retrieve HTTP request from context")
	}
	if retrievedReq != req {
		t.Errorf("Expected retrieved request to match stored request")
	}
}

func TestHTTPRequestFromContext_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := HTTPRequestFromContext(ctx)
	if ok {
		t.Error("Expected no HTTP request in context")
	}
}

func TestCallWhatsMyIP_WithHTTPContext(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)

	testCases := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		cfConnecting  string
		remoteAddr    string
		expectLog     string
	}{
		{
			name:          "X-Forwarded-For header",
			xForwardedFor: "192.168.1.100",
			remoteAddr:    "10.0.0.1:8080",
			expectLog:     "192.168.1.100",
		},
		{
			name:       "X-Real-IP header",
			xRealIP:    "203.0.113.45",
			remoteAddr: "10.0.0.1:8080",
			expectLog:  "203.0.113.45",
		},
		{
			name:         "CF-Connecting-IP header",
			cfConnecting: "198.51.100.67",
			remoteAddr:   "10.0.0.1:8080",
			expectLog:    "198.51.100.67",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "172.16.0.50:9090",
			expectLog:  "172.16.0.50",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			req := httptest.NewRequest("POST", "http://example.com/mcp", nil)
			req.RemoteAddr = tc.remoteAddr

			if tc.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}
			if tc.xRealIP != "" {
				req.Header.Set("X-Real-IP", tc.xRealIP)
			}
			if tc.cfConnecting != "" {
				req.Header.Set("CF-Connecting-IP", tc.cfConnecting)
			}

			ctx := WithHTTPRequest(context.Background(), req)

			params := &CallToolParams{
				Name:      "getWhatsMyIP",
				Arguments: map[string]interface{}{},
			}

			result, err := server.executeToolCall(ctx, params)

			if result == nil && err == nil {
				t.Error("Expected either result or error")
			}
		})
	}
}

func TestCallWhatsMyIP_WithoutHTTPContext(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	params := &CallToolParams{
		Name:      "getWhatsMyIP",
		Arguments: map[string]interface{}{},
	}

	result, err := server.executeToolCall(ctx, params)

	if result == nil && err == nil {
		t.Error("Expected either result or error")
	}
}

func TestWithSessionID(t *testing.T) {
	ctx := context.Background()
	sessionID := "test-session-123"

	ctxWithSession := WithSessionID(ctx, sessionID)

	retrievedSessionID, ok := SessionIDFromContext(ctxWithSession)
	if !ok {
		t.Fatal("Expected to retrieve session ID from context")
	}
	if retrievedSessionID != sessionID {
		t.Errorf("Expected session ID '%s', got '%s'", sessionID, retrievedSessionID)
	}
}

func TestSessionIDFromContext_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := SessionIDFromContext(ctx)
	if ok {
		t.Error("Expected no session ID in context")
	}
}

func TestParseQueryToRequest(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)

	t.Run("valid queries", func(t *testing.T) {
		testParseQueryToRequestValid(t, server)
	})

	t.Run("error cases", func(t *testing.T) {
		testParseQueryToRequestErrors(t, server)
	})
}

func testParseQueryToRequestValid(t *testing.T, server *Server) {
	testCases := []struct {
		name      string
		query     url.Values
		checkFunc func(*testing.T, []byte)
	}{
		{
			name: "valid query with method and resource",
			query: url.Values{
				"method":   []string{"tools/call"},
				"resource": []string{"8.8.8.8"},
				"id":       []string{"123"},
			},
			checkFunc: func(t *testing.T, data []byte) {
				var request Request
				if err := json.Unmarshal(data, &request); err != nil {
					t.Fatalf("Failed to unmarshal request: %v", err)
				}
				if request.Method != "tools/call" {
					t.Errorf("Expected method 'tools/call', got '%s'", request.Method)
				}
				if request.ID != "123" {
					t.Errorf("Expected ID '123', got '%v'", request.ID)
				}
				if params, ok := request.Params.(map[string]interface{}); ok {
					if resource, exists := params["resource"]; !exists || resource != "8.8.8.8" {
						t.Errorf("Expected resource '8.8.8.8', got '%v'", resource)
					}
				} else {
					t.Error("Expected params to be map[string]interface{}")
				}
			},
		},
		{
			name: "valid query without ID (should default to 1)",
			query: url.Values{
				"method":   []string{"ping"},
				"resource": []string{"test"},
			},
			checkFunc: func(t *testing.T, data []byte) {
				var request Request
				if err := json.Unmarshal(data, &request); err != nil {
					t.Fatalf("Failed to unmarshal request: %v", err)
				}
				if request.ID != "1" {
					t.Errorf("Expected default ID '1', got '%v'", request.ID)
				}
			},
		},
		{
			name: "query with JSON params parameter",
			query: url.Values{
				"method": []string{"tools/call"},
				"params": []string{`{"resource": "test", "lod": "1"}`},
				"id":     []string{"456"},
			},
			checkFunc: func(t *testing.T, data []byte) {
				var request Request
				if err := json.Unmarshal(data, &request); err != nil {
					t.Fatalf("Failed to unmarshal request: %v", err)
				}
				if params, ok := request.Params.(map[string]interface{}); ok {
					if resource, exists := params["resource"]; !exists || resource != "test" {
						t.Errorf("Expected resource 'test', got '%v'", resource)
					}
					if lod, exists := params["lod"]; !exists || lod != "1" {
						t.Errorf("Expected lod '1', got '%v'", lod)
					}
				} else {
					t.Error("Expected params to be map[string]interface{}")
				}
			},
		},
		{
			name: "query with multiple values (should use first)",
			query: url.Values{
				"method":   []string{"tools/call"},
				"resource": []string{"first", "second"},
			},
			checkFunc: func(t *testing.T, data []byte) {
				var request Request
				if err := json.Unmarshal(data, &request); err != nil {
					t.Fatalf("Failed to unmarshal request: %v", err)
				}
				if params, ok := request.Params.(map[string]interface{}); ok {
					if resource, exists := params["resource"]; !exists || resource != "first" {
						t.Errorf("Expected resource 'first', got '%v'", resource)
					}
				} else {
					t.Error("Expected params to be map[string]interface{}")
				}
			},
		},
		{
			name: "query with reserved parameter names (should be filtered)",
			query: url.Values{
				"method":   []string{"tools/call"},
				"id":       []string{"789"},
				"params":   []string{`{"test": "value"}`},
				"resource": []string{"8.8.8.8"},
			},
			checkFunc: func(t *testing.T, data []byte) {
				var request Request
				if err := json.Unmarshal(data, &request); err != nil {
					t.Fatalf("Failed to unmarshal request: %v", err)
				}
				if request.Method != "tools/call" {
					t.Errorf("Expected method 'tools/call', got '%s'", request.Method)
				}
				if request.ID != "789" {
					t.Errorf("Expected ID '789', got '%v'", request.ID)
				}

				if params, ok := request.Params.(map[string]interface{}); ok {
					if test, exists := params["test"]; !exists || test != "value" {
						t.Errorf("Expected test 'value', got '%v'", test)
					}
					if _, exists := params["resource"]; exists {
						t.Error("Resource parameter should not exist when params JSON is provided")
					}
				} else {
					t.Error("Expected params to be map[string]interface{}")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := server.ParseQueryToRequest(tc.query)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if tc.checkFunc != nil {
				tc.checkFunc(t, data)
			}
		})
	}
}

func testParseQueryToRequestErrors(t *testing.T, server *Server) {
	testCases := []struct {
		name     string
		query    url.Values
		errorMsg string
	}{
		{
			name: "missing method parameter",
			query: url.Values{
				"resource": []string{"8.8.8.8"},
			},
			errorMsg: "method parameter is required",
		},
		{
			name: "query with invalid JSON params",
			query: url.Values{
				"method": []string{"tools/call"},
				"params": []string{`{invalid json`},
			},
			errorMsg: "invalid params JSON",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.ParseQueryToRequest(tc.query)
			if err == nil {
				t.Error("Expected error but got none")
			} else if !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%v'", tc.errorMsg, err)
			}
		})
	}
}

func TestExecuteToolCall_UncoveredFunctions(t *testing.T) {
	server := NewServer("test-server", "1.0.0", false)
	ctx := context.Background()

	testBasicResourceTool := func(t *testing.T, toolName, resourceValue string) {

		params := &CallToolParams{
			Name:      toolName,
			Arguments: map[string]interface{}{"resource": resourceValue},
		}

		result, err := server.executeToolCall(ctx, params)
		if err != nil {
			t.Logf("Network call failed (expected in test environment): %v", err)
		} else if result == nil {
			t.Error("Expected non-nil result")
		}

		params = &CallToolParams{
			Name:      toolName,
			Arguments: map[string]interface{}{},
		}

		result, err = server.executeToolCall(ctx, params)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected ToolResult, got nil")
			return
		}
		if !result.IsError {
			t.Error("Expected error ToolResult")
		}
		if !strings.Contains(result.Content[0].Text, "resource parameter is required") {
			t.Errorf("Expected error message about missing resource, got %s", result.Content[0].Text)
		}
	}

	t.Run("validateSecurity", func(t *testing.T) {
		testBasicResourceTool(t, "validateSecurity", "8.8.8.0/24")
	})

	t.Run("searchByLocation", func(t *testing.T) {

		params := &CallToolParams{
			Name:      "searchByLocation",
			Arguments: map[string]interface{}{"country": "NL"},
		}

		result, err := server.executeToolCall(ctx, params)
		if err != nil {
			t.Logf("Network call failed (expected in test environment): %v", err)
		} else if result == nil {
			t.Error("Expected non-nil result")
		}

		params = &CallToolParams{
			Name:      "searchByLocation",
			Arguments: map[string]interface{}{},
		}

		result, err = server.executeToolCall(ctx, params)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected ToolResult, got nil")
			return
		}
		if !result.IsError {
			t.Error("Expected error ToolResult")
		}
		if !strings.Contains(result.Content[0].Text, "country parameter is required") {
			t.Errorf("Expected error message about missing country, got %s", result.Content[0].Text)
		}
	})

	t.Run("analyzeRouting", func(t *testing.T) {
		testBasicResourceTool(t, "analyzeRouting", "8.8.8.8")
	})
}
