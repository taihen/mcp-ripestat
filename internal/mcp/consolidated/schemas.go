package consolidated

var ConsolidatedToolDescriptions = map[string]string{
	"investigateResource":  "Comprehensive investigation of IP addresses, prefixes, or ASNs with intelligent routing to relevant endpoints based on resource type and requested operations.",
	"analyzeRouting":       "BGP and routing analysis with timeframe support for consistency checks, path optimization, updates monitoring, and looking glass data.",
	"queryRegistry":        "Registry and administrative data retrieval including WHOIS information, allocation history, address space hierarchy, and contact information.",
	"validateSecurity":     "Security and compliance validation including RPKI validation, abuse contact discovery, and BGP hijacking detection.",
	"exploreRelationships": "Network topology and relationship exploration including AS neighbors, announced prefixes, and related network discovery.",
	"searchByLocation":     "Geographic analysis and location-based resource discovery for ASNs, prefixes, and statistics by country.",
}

var ConsolidatedSchemas = map[string]map[string]interface{}{
	"investigateResource": {
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
	"analyzeRouting": {
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
	"queryRegistry": {
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
	"validateSecurity": {
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
	"exploreRelationships": {
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
	"searchByLocation": {
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
}
