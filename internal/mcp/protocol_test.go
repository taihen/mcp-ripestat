package mcp

import (
	"encoding/json"
	"testing"
)

func TestCreateInitializeResult(t *testing.T) {
	result := CreateInitializeResult("test-server", "1.0.0")

	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("Expected ProtocolVersion to be %s, got %s", ProtocolVersion, result.ProtocolVersion)
	}
	if result.ServerInfo.Name != "test-server" {
		t.Errorf("Expected ServerInfo.Name to be 'test-server', got %s", result.ServerInfo.Name)
	}
	if result.ServerInfo.Version != "1.0.0" {
		t.Errorf("Expected ServerInfo.Version to be '1.0.0', got %s", result.ServerInfo.Version)
	}
	if result.Capabilities == nil {
		t.Error("Expected Capabilities to be non-nil")
	}

	// Test new transport capabilities
	caps, ok := result.Capabilities.(*Capabilities)
	if !ok {
		t.Fatal("Expected Capabilities to be *Capabilities type")
	}

	if caps.Transport == nil {
		t.Error("Expected Transport capability to be non-nil")
	} else {
		if caps.Transport.HTTP == nil {
			t.Error("Expected HTTP transport capability to be non-nil")
		} else {
			if !caps.Transport.HTTP.Streamable {
				t.Error("Expected HTTP transport to be streamable")
			}
			expectedMethods := []string{"POST", "GET"}
			if len(caps.Transport.HTTP.Methods) != len(expectedMethods) {
				t.Errorf("Expected %d HTTP methods, got %d", len(expectedMethods), len(caps.Transport.HTTP.Methods))
			}
			for i, method := range caps.Transport.HTTP.Methods {
				if method != expectedMethods[i] {
					t.Errorf("Expected method %s at index %d, got %s", expectedMethods[i], i, method)
				}
			}
		}
	}
}

func TestCreateToolsList(t *testing.T) {
	tools := CreateToolsList()

	if len(tools.Tools) == 0 {
		t.Error("Expected tools list to have at least one tool")
	}

	// Check that all expected tools are present
	expectedTools := []string{
		"getNetworkInfo",
		"getASOverview",
		"getAnnouncedPrefixes",
		"getRoutingStatus",
		"getWhois",
		"getAbuseContactFinder",
		"getRPKIValidation",
		"getASNNeighbours",
		"getLookingGlass",
		"getWhatsMyIP",
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		toolNames[tool.Name] = true

		// Check that each tool has required fields
		if tool.Name == "" {
			t.Error("Tool name should not be empty")
		}
		if tool.Description == "" {
			t.Error("Tool description should not be empty")
		}
		if tool.InputSchema == nil {
			t.Error("Tool input schema should not be nil")
		}
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool %s not found in tools list", expectedTool)
		}
	}
}

func TestParseCallToolParams(t *testing.T) {
	tests := []struct {
		name    string
		params  interface{}
		want    *CallToolParams
		wantErr bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"name": "getNetworkInfo",
				"arguments": map[string]interface{}{
					"resource": "8.8.8.8",
				},
			},
			want: &CallToolParams{
				Name: "getNetworkInfo",
				Arguments: map[string]interface{}{
					"resource": "8.8.8.8",
				},
			},
			wantErr: false,
		},
		{
			name: "params without arguments",
			params: map[string]interface{}{
				"name": "getWhatsMyIP",
			},
			want: &CallToolParams{
				Name: "getWhatsMyIP",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCallToolParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCallToolParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.want.Name {
					t.Errorf("ParseCallToolParams() Name = %v, want %v", got.Name, tt.want.Name)
				}
			}
		})
	}
}

func TestCreateToolResult(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		isError bool
	}{
		{
			name:    "success result",
			text:    "Success message",
			isError: false,
		},
		{
			name:    "error result",
			text:    "Error message",
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateToolResult(tt.text, tt.isError)

			if result.IsError != tt.isError {
				t.Errorf("CreateToolResult() IsError = %v, want %v", result.IsError, tt.isError)
			}
			if len(result.Content) != 1 {
				t.Errorf("CreateToolResult() Content length = %d, want 1", len(result.Content))
			}
			if result.Content[0].Type != "text" {
				t.Errorf("CreateToolResult() Content[0].Type = %s, want 'text'", result.Content[0].Type)
			}
			if result.Content[0].Text != tt.text {
				t.Errorf("CreateToolResult() Content[0].Text = %s, want %s", result.Content[0].Text, tt.text)
			}
		})
	}
}

func TestCreateToolResultFromJSON(t *testing.T) {
	testData := map[string]interface{}{
		"key":    "value",
		"number": 42,
		"nested": map[string]interface{}{
			"inner": "data",
		},
	}

	result := CreateToolResultFromJSON(testData)

	if result.IsError {
		t.Error("CreateToolResultFromJSON() should not create error result for valid data")
	}
	if len(result.Content) != 1 {
		t.Errorf("CreateToolResultFromJSON() Content length = %d, want 1", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("CreateToolResultFromJSON() Content[0].Type = %s, want 'text'", result.Content[0].Type)
	}

	// Verify the JSON is valid
	var parsed interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &parsed); err != nil {
		t.Errorf("CreateToolResultFromJSON() produced invalid JSON: %v", err)
	}
}

func TestProtocolVersion(t *testing.T) {
	if ProtocolVersion != "2025-06-18" {
		t.Errorf("ProtocolVersion = %s, want '2025-06-18'", ProtocolVersion)
	}
}

func TestCapabilitiesStructure(t *testing.T) {
	caps := &Capabilities{
		Tools:     &ToolsCapability{},
		Resources: &ResourcesCapability{},
		Prompts:   &PromptsCapability{},
		Logging:   &LoggingCapability{},
	}

	// Test that capabilities can be marshaled to JSON
	data, err := json.Marshal(caps)
	if err != nil {
		t.Errorf("Failed to marshal Capabilities: %v", err)
	}

	// Test that capabilities can be unmarshaled from JSON
	var unmarshaled Capabilities
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal Capabilities: %v", err)
	}
}

func TestToolInputSchema(t *testing.T) {
	tools := CreateToolsList()

	for _, tool := range tools.Tools {
		// Verify each tool has a valid input schema
		schemaData, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Errorf("Tool %s has invalid input schema: %v", tool.Name, err)
			continue
		}

		// Verify the schema is a valid object
		var schema map[string]interface{}
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			t.Errorf("Tool %s input schema is not a valid object: %v", tool.Name, err)
			continue
		}

		// Check that it has a type field
		if schemaType, ok := schema["type"].(string); !ok || schemaType != "object" {
			t.Errorf("Tool %s input schema type should be 'object', got %v", tool.Name, schema["type"])
		}

		// Check that it has properties
		if _, ok := schema["properties"]; !ok {
			t.Errorf("Tool %s input schema should have 'properties' field", tool.Name)
		}
	}
}
