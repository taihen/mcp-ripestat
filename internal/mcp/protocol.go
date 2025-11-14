package mcp

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/taihen/mcp-ripestat/internal/mcp/consolidated"
)


const ProtocolVersion = "2025-06-18"


type InitializeParams struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ClientInfo      ClientInfo  `json:"clientInfo"`
}


type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}


type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}


type InitializeResult struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ServerInfo      ServerInfo  `json:"serverInfo"`
}


type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	Roots     *RootsCapability     `json:"roots,omitempty"`
	Transport *TransportCapability `json:"transport,omitempty"`
}


type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}


type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}


type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}


type LoggingCapability struct{}


type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}


type TransportCapability struct {
	HTTP *HTTPTransportCapability `json:"http,omitempty"`
}


type HTTPTransportCapability struct {
	Streamable bool     `json:"streamable"`
	Methods    []string `json:"methods"`
}


type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}


type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}


type CallToolParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
	Meta      interface{} `json:"_meta,omitempty"`
}


type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}


type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}


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


func CreateLegacyInitializeResult(serverName, serverVersion string) *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2025-03-26",
		Capabilities: &Capabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
			Prompts:   &PromptsCapability{ListChanged: false},
			Logging:   &LoggingCapability{},
			Roots:     &RootsCapability{ListChanged: false},

		},
		ServerInfo: ServerInfo{
			Name:    serverName,
			Version: serverVersion,
		},
	}
}


func CreateToolsList() *ToolsListResult {
	tools := []Tool{}


	toolNames := make([]string, 0, len(consolidated.ConsolidatedSchemas))
	for toolName := range consolidated.ConsolidatedSchemas {
		toolNames = append(toolNames, toolName)
	}
	sort.Strings(toolNames)


	for _, toolName := range toolNames {
		schema := consolidated.ConsolidatedSchemas[toolName]
		description, exists := consolidated.ConsolidatedToolDescriptions[toolName]
		if !exists {

			description = "Consolidated tool for RIPEstat operations"
		}
		tools = append(tools, Tool{
			Name:        toolName,
			Description: description,
			InputSchema: schema,
		})
	}


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


func CreateToolResultFromJSON(data interface{}) *ToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return CreateToolResult(fmt.Sprintf("Error marshaling result: %v", err), true)
	}

	return CreateToolResult(string(jsonData), false)
}
