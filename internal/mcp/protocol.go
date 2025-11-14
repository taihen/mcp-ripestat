package mcp

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/taihen/mcp-ripestat/internal/mcp/consolidated"
)

// MCP Protocol Version.
const ProtocolVersion = "2025-06-18"

// Initialize request parameters.
type InitializeParams struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ClientInfo      ClientInfo  `json:"clientInfo"`
}

// ClientInfo represents information about the client.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Initialize result.
type InitializeResult struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ServerInfo      ServerInfo  `json:"serverInfo"`
}

// Server capabilities.
type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	Roots     *RootsCapability     `json:"roots,omitempty"`
	Transport *TransportCapability `json:"transport,omitempty"`
}

// ToolsCapability represents tools capability.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts capability.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability.
type LoggingCapability struct{}

// RootsCapability represents roots capability.
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// TransportCapability represents transport capability.
type TransportCapability struct {
	HTTP *HTTPTransportCapability `json:"http,omitempty"`
}

// HTTPTransportCapability represents HTTP transport capability.
type HTTPTransportCapability struct {
	Streamable bool     `json:"streamable"`
	Methods    []string `json:"methods"`
}

// Tool represents a tool that can be called.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// ToolsListResult represents the result of listing tools.
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams represents parameters for calling a tool.
type CallToolParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
	Meta      interface{} `json:"_meta,omitempty"`
}

// ToolResult represents the result of calling a tool.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content returned by a tool.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CreateInitializeResult creates an initialize result for the server.
func CreateInitializeResult(serverName, serverVersion string) *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: &Capabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
			Prompts:   &PromptsCapability{ListChanged: false},
			Logging:   &LoggingCapability{},
			Roots:     &RootsCapability{ListChanged: false},
			Transport: &TransportCapability{
				HTTP: &HTTPTransportCapability{
					Streamable: true,
					Methods:    []string{"POST", "GET"},
				},
			},
		},
		ServerInfo: ServerInfo{
			Name:    serverName,
			Version: serverVersion,
		},
	}
}

// CreateLegacyInitializeResult creates a simplified initialize result for older protocol versions.
func CreateLegacyInitializeResult(serverName, serverVersion string) *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2025-03-26", // Use older protocol version
		Capabilities: &Capabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
			Prompts:   &PromptsCapability{ListChanged: false},
			Logging:   &LoggingCapability{},
			Roots:     &RootsCapability{ListChanged: false},
			// No transport capabilities for legacy clients to avoid confusion
		},
		ServerInfo: ServerInfo{
			Name:    serverName,
			Version: serverVersion,
		},
	}
}

// CreateToolsList creates a list of available tools.
func CreateToolsList() *ToolsListResult {
	tools := []Tool{}

	// Collect consolidated tool names and sort them for deterministic ordering
	toolNames := make([]string, 0, len(consolidated.ConsolidatedSchemas))
	for toolName := range consolidated.ConsolidatedSchemas {
		toolNames = append(toolNames, toolName)
	}
	sort.Strings(toolNames)

	// Add consolidated tools from the centralized definitions in sorted order
	for _, toolName := range toolNames {
		schema := consolidated.ConsolidatedSchemas[toolName]
		description, exists := consolidated.ConsolidatedToolDescriptions[toolName]
		if !exists {
			// Fallback if description is missing (shouldn't happen, but be defensive)
			description = "Consolidated tool for RIPEstat operations"
		}
		tools = append(tools, Tool{
			Name:        toolName,
			Description: description,
			InputSchema: schema,
		})
	}

	// Add non-consolidated tools (e.g., getWhatsMyIP)
	tools = append(tools, Tool{
		Name:        "getWhatsMyIP",
		Description: "Get the caller's public IP address. Respects X-Forwarded-For headers when behind a proxy.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	return &ToolsListResult{Tools: tools}
}

// ParseCallToolParams parses tool call parameters from JSON.
func ParseCallToolParams(params interface{}) (*CallToolParams, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var callParams CallToolParams
	if err := json.Unmarshal(jsonData, &callParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal call tool params: %w", err)
	}

	return &callParams, nil
}

// CreateToolResult creates a tool result with text content.
func CreateToolResult(text string, isError bool) *ToolResult {
	return &ToolResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: isError,
	}
}

// CreateToolResultFromJSON creates a tool result from JSON data.
func CreateToolResultFromJSON(data interface{}) *ToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error marshaling result: %v", err), true)
	}

	return CreateToolResult(string(jsonData), false)
}
