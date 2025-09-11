package consolidated

// ConsolidatedSchemas contains the JSON schemas for all consolidated tools
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