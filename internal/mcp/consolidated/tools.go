package consolidated

import (
	"context"
	"fmt"
)

// ToolExecutor defines the interface for executing individual RIPEstat endpoints.
type ToolExecutor interface {
	ExecuteEndpoint(ctx context.Context, endpoint string, resource string, params map[string]interface{}) (interface{}, error)
}

// Tools provides the consolidated tool implementations.
type Tools struct {
	executor ToolExecutor
}

// NewTools creates a new consolidated tools instance.
func NewTools(executor ToolExecutor) *Tools {
	return &Tools{
		executor: executor,
	}
}

// executeOperations is a convenience wrapper for the detect-route-execute pattern.
func (ct *Tools) executeOperations(
	ctx context.Context,
	resource string,
	operations []Operation,
	depth string,
) (*Result, error) {
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	return ct.executeAndAggregate(ctx, detected, operations, routes, depth)
}

// InvestigateResource - Primary investigation tool with auto-detection and intelligent routing.
func (ct *Tools) InvestigateResource(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}

	operationStrings, err := extractStringSliceWithDefault(params, "operations", []string{"overview"})
	if err != nil {
		return nil, err
	}

	depth := extractOptionalString(params, "depth", DepthBasic)

	return ct.executeOperations(ctx, resource, toOperations(operationStrings), depth)
}

// AnalyzeRouting - BGP and routing analysis with timeframe support.
func (ct *Tools) AnalyzeRouting(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}

	analysisTypes, err := extractStringSliceWithDefault(params, "analysis", []string{AnalysisConsistency})
	if err != nil {
		return nil, err
	}

	operations, err := mapStringsToOperations(analysisTypes, "analysis")
	if err != nil {
		return nil, err
	}

	timeframe := extractOptionalString(params, "timeframe", TimeframeCurrent)

	result, err := ct.executeOperations(ctx, resource, operations, DepthDetailed)
	if err != nil {
		return nil, err
	}

	result.AddMetadata("timeframe", timeframe)
	return result, nil
}

// QueryRegistry - Registry and administrative data.
func (ct *Tools) QueryRegistry(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}

	dataTypes, err := extractStringSliceWithDefault(params, "data", []string{DataTypeWhois})
	if err != nil {
		return nil, err
	}

	operations, err := mapStringsToOperations(dataTypes, "data")
	if err != nil {
		return nil, err
	}

	format := extractOptionalString(params, "format", FormatSummary)

	depth := DepthBasic
	if format == FormatDetailed {
		depth = DepthDetailed
	}

	return ct.executeOperations(ctx, resource, operations, depth)
}

// ValidateSecurity - Security and compliance checks.
func (ct *Tools) ValidateSecurity(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}

	checkTypes, err := extractStringSliceWithDefault(
		params,
		"checks",
		[]string{SecurityCheckRPKI, SecurityCheckAbuseContacts},
	)
	if err != nil {
		return nil, err
	}

	operations, err := mapStringsToOperations(checkTypes, "checks")
	if err != nil {
		return nil, err
	}

	asnParam := extractOptionalString(params, "asn", "")

	result, err := ct.executeOperations(ctx, resource, operations, DepthDetailed)
	if err != nil {
		return nil, err
	}

	if asnParam != "" {
		result.AddMetadata("asn", asnParam)
	}

	return result, nil
}

// ExploreRelationships - Network topology and relationships.
func (ct *Tools) ExploreRelationships(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}

	relationshipTypes, err := extractStringSliceWithDefault(
		params,
		"relationships",
		[]string{RelationshipNeighbors},
	)
	if err != nil {
		return nil, err
	}

	operations, err := mapStringsToOperations(relationshipTypes, "relationships")
	if err != nil {
		return nil, err
	}

	scope := extractOptionalString(params, "scope", ScopeDirect)

	depth := DepthBasic
	if scope == ScopeExtended {
		depth = DepthDetailed
	}

	result, err := ct.executeOperations(ctx, resource, operations, depth)
	if err != nil {
		return nil, err
	}

	result.AddMetadata("scope", scope)
	return result, nil
}

// SearchByLocation - Geographic analysis.
func (ct *Tools) SearchByLocation(ctx context.Context, params map[string]interface{}) (*Result, error) {
	country, err := extractRequiredString(params, "country")
	if err != nil {
		return nil, err
	}

	typeParam := extractOptionalString(params, "type", LocationTypeASNs)

	// Validate type parameter
	switch typeParam {
	case LocationTypeASNs, LocationTypePrefixes, LocationTypeStatistics:
		// Valid type
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
	routes, err := RouteOperations(detected, []Operation{OpOverview})
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	return ct.executeAndAggregate(ctx, detected, []Operation{OpOverview}, routes, DepthBasic)
}

// executeAndAggregate executes all endpoints and aggregates the results.
func (ct *Tools) executeAndAggregate(
	ctx context.Context,
	resource *DetectedResource,
	operations []Operation,
	routes *RouteResult,
	depth string,
) (*Result, error) {
	result := &Result{
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
	result.AddMetadataMap(map[string]interface{}{
		"endpoints_called": routes.Endpoints,
		"depth":            depth,
		"resource_type":    resource.Type.String(),
	})

	return result, nil
}
