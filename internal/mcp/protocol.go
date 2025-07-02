package mcp

import (
	"encoding/json"
	"fmt"
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
	tools := []Tool{
		{
			Name:        "getNetworkInfo",
			Description: "Get network information for an IP address or prefix.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address or prefix to query.",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "getASOverview",
			Description: "Get an overview of an Autonomous System (AS).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The AS number to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getAnnouncedPrefixes",
			Description: "Get a list of prefixes announced by an Autonomous System (AS).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The AS number to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getRoutingStatus",
			Description: "Get the routing status for an IP prefix.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getRoutingHistory",
			Description: "Get routing history information for an IP address, prefix, or ASN.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address, prefix, or ASN to query for routing history.",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "Start time for the query in ISO8601 format (e.g., '2024-01-01T00:00:00Z'). If omitted, uses default historical range.",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "End time for the query in ISO8601 format (e.g., '2024-12-31T23:59:59Z'). If omitted, uses current time.",
					},
					"max_results": map[string]interface{}{
						"type":        "string",
						"description": "Maximum number of routing events to return. Helps limit response size for large datasets.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getWhois",
			Description: "Get whois information for an IP address, prefix, or ASN.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address, prefix, or ASN to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getAbuseContactFinder",
			Description: "Get abuse contact information for an IP address or prefix.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address or prefix to query for abuse contacts.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getRPKIValidation",
			Description: "Get RPKI validation status for a resource (ASN) and prefix combination.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The ASN to validate against the prefix.",
					},
					"prefix": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to validate.",
					},
				},
				"required": []string{"resource", "prefix"},
			},
		},
		{
			Name:        "getASNNeighbours",
			Description: "Get ASN neighbours for an Autonomous System. Left neighbours are downstream providers, right neighbours are upstream providers.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The AS number to query for neighbours.",
					},
					"lod": map[string]interface{}{
						"type":        "string",
						"description": "Level of detail: 0 (basic) or 1 (detailed with power, v4_peers, v6_peers). Default is 0.",
					},
					"query_time": map[string]interface{}{
						"type":        "string",
						"description": "Query time in ISO8601 format for historical data. If omitted, uses latest snapshot.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getLookingGlass",
			Description: "Get looking glass information for an IP prefix, showing BGP routing data from RIPE NCC's Route Reflection Collectors (RRCs).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to query for looking glass information.",
					},
					"look_back_limit": map[string]interface{}{
						"type":        "string",
						"description": "Time limit in seconds to look back for BGP data. Maximum is 172800 seconds (48 hours). Default is 0.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getCountryASNs",
			Description: "Get Autonomous System Numbers (ASNs) for a given country code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "Two-letter ISO country code (e.g., 'nl', 'us', 'de').",
					},
					"lod": map[string]interface{}{
						"type":        "string",
						"description": "Level of detail: 0 (basic stats) or 1 (includes lists of routed/non-routed ASNs). Default is 0.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getRPKIHistory",
			Description: "Get RPKI history information for an IP prefix, showing the historical RPKI validation status.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to query for RPKI history.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getBGPlay",
			Description: "Get BGP play data for an IP address or prefix, showing BGP routing events and timeline.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address or prefix to query for BGP play data.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getPrefixRoutingConsistency",
			Description: "Get prefix routing consistency information for an IP prefix, showing BGP routing consistency data.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to query for routing consistency.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getPrefixOverview",
			Description: "Get prefix overview information for an IP prefix.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP prefix to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getAddressSpaceHierarchy",
			Description: "Get address space hierarchy information for an IP address or prefix.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address or prefix to query.",
					},
				},
				"required": []string{"resource"},
			},
		},
		{
			Name:        "getWhatsMyIP",
			Description: "Get the caller's public IP address. Respects X-Forwarded-For headers when behind a proxy.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

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
