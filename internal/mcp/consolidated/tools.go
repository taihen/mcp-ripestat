package consolidated

import (
	"context"
	"fmt"
)

// ToolExecutor defines the interface for executing individual RIPEstat endpoints
type ToolExecutor interface {
	ExecuteEndpoint(ctx context.Context, endpoint string, resource string, params map[string]interface{}) (interface{}, error)
}

// ConsolidatedTools provides the consolidated tool implementations
type ConsolidatedTools struct {
	executor ToolExecutor
}

// NewConsolidatedTools creates a new consolidated tools instance
func NewConsolidatedTools(executor ToolExecutor) *ConsolidatedTools {
	return &ConsolidatedTools{
		executor: executor,
	}
}

// InvestigateResource - Primary investigation tool with auto-detection and intelligent routing
func (ct *ConsolidatedTools) InvestigateResource(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	// Extract parameters
	resource, ok := params["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	operationsParam, ok := params["operations"]
	if !ok {
		operationsParam = []string{"overview"} // Default to overview
	}

	// Convert operations to []Operation
	var operations []Operation
	switch v := operationsParam.(type) {
	case []string:
		for _, op := range v {
			operations = append(operations, Operation(op))
		}
	case []interface{}:
		for _, op := range v {
			if opStr, ok := op.(string); ok {
				operations = append(operations, Operation(opStr))
			}
		}
	default:
		return nil, fmt.Errorf("operations must be an array of strings")
	}

	depth := "basic" // Default depth
	if d, ok := params["depth"].(string); ok {
		depth = d
	}

	// Detect resource type
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	// Execute endpoints and aggregate results
	return ct.executeAndAggregate(ctx, detected, operations, routes, depth)
}

// AnalyzeRouting - BGP and routing analysis with timeframe support
func (ct *ConsolidatedTools) AnalyzeRouting(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	resource, ok := params["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	analysisParam, ok := params["analysis"]
	if !ok {
		analysisParam = []string{"consistency"} // Default analysis
	}

	// Convert analysis types to operations
	var operations []Operation
	switch v := analysisParam.(type) {
	case []string:
		for _, analysis := range v {
			switch analysis {
			case "consistency":
				operations = append(operations, OpConsistency)
			case "path-optimization":
				operations = append(operations, OpRouting)
			case "updates":
				operations = append(operations, OpUpdates)
			case "looking-glass":
				operations = append(operations, OpLookingGlass)
			default:
				return nil, fmt.Errorf("unsupported analysis type: %s", analysis)
			}
		}
	case []interface{}:
		for _, analysis := range v {
			if analysisStr, ok := analysis.(string); ok {
				switch analysisStr {
				case "consistency":
					operations = append(operations, OpConsistency)
				case "path-optimization":
					operations = append(operations, OpRouting)
				case "updates":
					operations = append(operations, OpUpdates)
				case "looking-glass":
					operations = append(operations, OpLookingGlass)
				default:
					return nil, fmt.Errorf("unsupported analysis type: %s", analysisStr)
				}
			}
		}
	default:
		return nil, fmt.Errorf("analysis must be an array of strings")
	}

	timeframe := "current" // Default timeframe
	if tf, ok := params["timeframe"].(string); ok {
		timeframe = tf
	}

	// Detect resource type
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	// Execute with timeframe context
	result, err := ct.executeAndAggregate(ctx, detected, operations, routes, "detailed")
	if err != nil {
		return nil, err
	}

	// Add timeframe metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["timeframe"] = timeframe

	return result, nil
}

// QueryRegistry - Registry and administrative data
func (ct *ConsolidatedTools) QueryRegistry(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	resource, ok := params["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	dataParam, ok := params["data"]
	if !ok {
		dataParam = []string{"whois"} // Default to whois
	}

	// Convert data types to operations
	var operations []Operation
	switch v := dataParam.(type) {
	case []string:
		for _, data := range v {
			switch data {
			case "whois":
				operations = append(operations, OpOverview)
			case "allocation-history":
				operations = append(operations, OpHistory)
			case "hierarchy":
				operations = append(operations, OpHierarchy)
			case "contacts":
				operations = append(operations, OpSecurity)
			default:
				return nil, fmt.Errorf("unsupported data type: %s", data)
			}
		}
	case []interface{}:
		for _, data := range v {
			if dataStr, ok := data.(string); ok {
				switch dataStr {
				case "whois":
					operations = append(operations, OpOverview)
				case "allocation-history":
					operations = append(operations, OpHistory)
				case "hierarchy":
					operations = append(operations, OpHierarchy)
				case "contacts":
					operations = append(operations, OpSecurity)
				default:
					return nil, fmt.Errorf("unsupported data type: %s", dataStr)
				}
			}
		}
	default:
		return nil, fmt.Errorf("data must be an array of strings")
	}

	format := "summary" // Default format
	if f, ok := params["format"].(string); ok {
		format = f
	}

	// Detect resource type
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	// Execute with format context
	depth := "basic"
	if format == "detailed" {
		depth = "detailed"
	}

	return ct.executeAndAggregate(ctx, detected, operations, routes, depth)
}

// ValidateSecurity - Security and compliance checks
func (ct *ConsolidatedTools) ValidateSecurity(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	resource, ok := params["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	checksParam, ok := params["checks"]
	if !ok {
		checksParam = []string{"rpki", "abuse-contacts"} // Default checks
	}

	// Convert checks to operations
	var operations []Operation
	switch v := checksParam.(type) {
	case []string:
		for _, check := range v {
			switch check {
			case "rpki", "abuse-contacts", "bgp-hijacking":
				operations = append(operations, OpSecurity)
			default:
				return nil, fmt.Errorf("unsupported security check: %s", check)
			}
		}
	case []interface{}:
		for _, check := range v {
			if checkStr, ok := check.(string); ok {
				switch checkStr {
				case "rpki", "abuse-contacts", "bgp-hijacking":
					operations = append(operations, OpSecurity)
				default:
					return nil, fmt.Errorf("unsupported security check: %s", checkStr)
				}
			}
		}
	default:
		return nil, fmt.Errorf("checks must be an array of strings")
	}

	// Optional ASN parameter for RPKI validation
	var asnParam string
	if asn, ok := params["asn"].(string); ok {
		asnParam = asn
	}

	// Detect resource type
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	// Execute security checks
	result, err := ct.executeAndAggregate(ctx, detected, operations, routes, "detailed")
	if err != nil {
		return nil, err
	}

	// Add ASN context if provided
	if asnParam != "" {
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["asn"] = asnParam
	}

	return result, nil
}

// ExploreRelationships - Network topology and relationships
func (ct *ConsolidatedTools) ExploreRelationships(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	resource, ok := params["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	relationshipsParam, ok := params["relationships"]
	if !ok {
		relationshipsParam = []string{"neighbors"} // Default relationships
	}

	// Convert relationship types to operations
	var operations []Operation
	switch v := relationshipsParam.(type) {
	case []string:
		for _, rel := range v {
			switch rel {
			case "neighbors":
				operations = append(operations, OpNeighbors)
			case "announced-prefixes":
				operations = append(operations, OpRouting)
			case "related-networks":
				operations = append(operations, OpRelationships)
			default:
				return nil, fmt.Errorf("unsupported relationship type: %s", rel)
			}
		}
	case []interface{}:
		for _, rel := range v {
			if relStr, ok := rel.(string); ok {
				switch relStr {
				case "neighbors":
					operations = append(operations, OpNeighbors)
				case "announced-prefixes":
					operations = append(operations, OpRouting)
				case "related-networks":
					operations = append(operations, OpRelationships)
				default:
					return nil, fmt.Errorf("unsupported relationship type: %s", relStr)
				}
			}
		}
	default:
		return nil, fmt.Errorf("relationships must be an array of strings")
	}

	scope := "direct" // Default scope
	if s, ok := params["scope"].(string); ok {
		scope = s
	}

	// Detect resource type
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	// Execute with scope context
	depth := "basic"
	if scope == "extended" {
		depth = "detailed"
	}

	result, err := ct.executeAndAggregate(ctx, detected, operations, routes, depth)
	if err != nil {
		return nil, err
	}

	// Add scope metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["scope"] = scope

	return result, nil
}

// SearchByLocation - Geographic analysis
func (ct *ConsolidatedTools) SearchByLocation(ctx context.Context, params map[string]interface{}) (*ConsolidatedResult, error) {
	country, ok := params["country"].(string)
	if !ok || country == "" {
		return nil, fmt.Errorf("country parameter is required")
	}

	typeParam, ok := params["type"].(string)
	if !ok {
		typeParam = "asns" // Default type
	}

	// Convert type to operations
	var operations []Operation
	switch typeParam {
	case "asns", "prefixes", "statistics":
		operations = append(operations, OpOverview)
	default:
		return nil, fmt.Errorf("unsupported type: %s", typeParam)
	}

	// Detect country resource
	detected, err := DetectResource(country)
	if err != nil {
		return nil, fmt.Errorf("failed to detect country: %w", err)
	}

	if detected.Type != Country {
		return nil, fmt.Errorf("invalid country code: %s", country)
	}

	// Route operations to endpoints
	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	return ct.executeAndAggregate(ctx, detected, operations, routes, "basic")
}

// executeAndAggregate executes all endpoints and aggregates the results
func (ct *ConsolidatedTools) executeAndAggregate(ctx context.Context, resource *DetectedResource, operations []Operation, routes *RouteResult, depth string) (*ConsolidatedResult, error) {
	result := &ConsolidatedResult{
		Resource:   resource,
		Operations: operations,
		Results:    make(map[string]interface{}),
		Errors:     make(map[string]string),
		Metadata:   make(map[string]interface{}),
	}

	// Execute each endpoint
	for _, endpoint := range routes.Endpoints {
		endpointResult, err := ct.executor.ExecuteEndpoint(ctx, endpoint, resource.Value, map[string]interface{}{
			"depth": depth,
		})
		if err != nil {
			result.Errors[endpoint] = err.Error()
			continue
		}

		result.Results[endpoint] = endpointResult
	}

	// Add routing metadata
	result.Metadata["endpoints_called"] = routes.Endpoints
	result.Metadata["depth"] = depth
	result.Metadata["resource_type"] = resource.Type.String()

	return result, nil
}