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
			Name:        "investigateResource",
			Description: "Comprehensive investigation of IP addresses, prefixes, or ASNs with intelligent routing to relevant endpoints based on resource type and requested operations.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "IP address, prefix, ASN, or country code to investigate",
					},
					"operations": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"overview", "routing", "security", "history", "neighbors", "relationships", "hierarchy"},
						},
						"description": "Operations to perform on the resource",
						"default":     []string{"overview"},
					},
					"depth": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"basic", "detailed", "comprehensive"},
						"description": "Level of detail for the investigation",
						"default":     "basic",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "analyzeRouting",
			Description: "BGP and routing analysis with timeframe support for consistency checks, path optimization, updates monitoring, and looking glass data.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "IP address, prefix, or ASN to analyze",
					},
					"analysis": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"consistency", "path-optimization", "updates", "looking-glass"},
						},
						"description": "Types of routing analysis to perform",
						"default":     []string{"consistency"},
					},
					"timeframe": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"current", "1d", "1w", "1m"},
						"description": "Timeframe for analysis",
						"default":     "current",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "queryRegistry",
			Description: "Registry and administrative data retrieval including WHOIS information, allocation history, address space hierarchy, and contact information.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "IP address, prefix, or ASN to query",
					},
					"data": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"whois", "allocation-history", "hierarchy", "contacts"},
						},
						"description": "Types of registry data to retrieve",
						"default":     []string{"whois"},
					},
					"format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"summary", "detailed"},
						"description": "Format of the returned data",
						"default":     "summary",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "validateSecurity",
			Description: "Security and compliance validation including RPKI validation, abuse contact discovery, and BGP hijacking detection.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "IP address, prefix, or ASN to validate",
					},
					"checks": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"rpki", "abuse-contacts", "bgp-hijacking"},
						},
						"description": "Security checks to perform",
						"default":     []string{"rpki", "abuse-contacts"},
					},
					"asn": map[string]interface{}{
						"type":        "string",
						"description": "ASN for RPKI validation (optional)",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "exploreRelationships",
			Description: "Network topology and relationship exploration including AS neighbors, announced prefixes, and related network discovery.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "ASN or prefix to explore relationships for",
					},
					"relationships": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"neighbors", "announced-prefixes", "related-networks"},
						},
						"description": "Types of relationships to explore",
						"default":     []string{"neighbors"},
					},
					"scope": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"direct", "extended"},
						"description": "Scope of relationship exploration",
						"default":     "direct",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "searchByLocation",
			Description: "Geographic analysis and location-based resource discovery for ASNs, prefixes, and statistics by country.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"country": map[string]interface{}{
						"type":        "string",
						"pattern":     "^[A-Z]{2}$",
						"description": "Two-letter ISO country code",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"asns", "prefixes", "statistics"},
						"description": "Type of location-based data to retrieve",
						"default":     "asns",
					},
				},
				"required":             []string{"country"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "getBGPState",
			Description: "Get the state of BGP routes for a resource at a certain point in time, as observed by all the RIS collectors.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "The IP address, prefix, AS, or comma-separated list of resources to query.",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"description": "Optional timestamp for historical BGP data (ISO format or Unix timestamp).",
					},
					"rrcs": map[string]interface{}{
						"type":        "string",
						"description": "Optional specific Route Collectors to query.",
					},
					"unix_timestamps": map[string]interface{}{
						"type":        "boolean",
						"description": "Format timestamps as Unix time (default: false).",
					},
				},
				"required":             []string{"resource"},
				"additionalProperties": false,
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
