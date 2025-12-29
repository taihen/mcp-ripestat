package consolidated

import (
	"context"
	"fmt"
	"reflect"
)

type ToolExecutor interface {
	ExecuteEndpoint(ctx context.Context, endpoint string, resource string, params map[string]interface{}) (interface{}, error)
}

type Tools struct {
	executor ToolExecutor
}

func NewTools(executor ToolExecutor) *Tools {
	return &Tools{
		executor: executor,
	}
}

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

func (ct *Tools) SearchByLocation(ctx context.Context, params map[string]interface{}) (*Result, error) {
	country, err := extractRequiredString(params, "country")
	if err != nil {
		return nil, err
	}

	typeParam := extractOptionalString(params, "type", LocationTypeASNs)

	switch typeParam {
	case LocationTypeASNs, LocationTypePrefixes, LocationTypeStatistics:
	default:
		return nil, fmt.Errorf("unsupported type: %s", typeParam)
	}

	detected, err := DetectResource(country)
	if err != nil {
		return nil, fmt.Errorf("failed to detect country: %w", err)
	}

	if detected.Type != Country {
		return nil, fmt.Errorf("invalid country code: %s", country)
	}

	routes, err := RouteOperations(detected, []Operation{OpOverview})
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	return ct.executeAndAggregate(ctx, detected, []Operation{OpOverview}, routes, DepthBasic)
}

func (ct *Tools) ExecuteIndividualEndpoint(ctx context.Context, endpoint string, args map[string]interface{}) (interface{}, error) {
	resource, ok := args["resource"].(string)
	if !ok || resource == "" {
		return nil, fmt.Errorf("resource parameter is required")
	}

	params := make(map[string]interface{})
	for key, value := range args {
		if key != "resource" {
			params[key] = value
		}
	}

	return ct.executor.ExecuteEndpoint(ctx, endpoint, resource, params)
}

func translateDepthToEndpointParams(endpoint string, depth string) map[string]interface{} {
	params := make(map[string]interface{})

	lodEndpoints := map[string]bool{
		"getASNNeighbours": true,
		"getCountryASNs":   true,
	}

	if lodEndpoints[endpoint] {
		var lod int
		switch depth {
		case DepthBasic:
			lod = 0
		case DepthDetailed, DepthComprehensive:
			lod = 1
		default:
			lod = 0
		}
		params["lod"] = lod
	}

	return params
}

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

	sortedEndpoints, err := topologicalSort(routes.Endpoints, routes.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to sort endpoints by dependencies: %w", err)
	}

	for _, endpoint := range sortedEndpoints {
		// Check for context cancellation before each endpoint call
		if ctx.Err() != nil {
			result.Errors[endpoint] = ctx.Err().Error()
			// Mark remaining endpoints as cancelled
			for _, remaining := range sortedEndpoints {
				if _, exists := result.Results[remaining]; !exists {
					if _, hasError := result.Errors[remaining]; !hasError {
						result.Errors[remaining] = "cancelled"
					}
				}
			}
			break
		}

		endpointParams := translateDepthToEndpointParams(endpoint, depth)

		resourceOverride := extractDependencyData(endpoint, routes.Dependencies, result.Results, endpointParams, resource)

		resourceValue := resource.Value
		if resourceOverride != "" {
			resourceValue = resourceOverride
		}

		endpointResult, err := ct.executor.ExecuteEndpoint(ctx, endpoint, resourceValue, endpointParams)
		if err != nil {
			result.Errors[endpoint] = err.Error()
			continue
		}

		result.Results[endpoint] = endpointResult
	}

	result.AddMetadataMap(map[string]interface{}{
		"endpoints_called": routes.Endpoints,
		"depth":            depth,
		"resource_type":    resource.Type.String(),
	})

	return result, nil
}

func topologicalSort(endpoints []string, dependencies map[string][]string) ([]string, error) {

	inDegree := make(map[string]int)
	for _, endpoint := range endpoints {
		inDegree[endpoint] = 0
	}

	for endpoint, deps := range dependencies {
		if _, exists := inDegree[endpoint]; !exists {
			continue
		}
		for _, dep := range deps {
			if _, exists := inDegree[dep]; exists {
				inDegree[endpoint]++
			}
		}
	}

	queue := make([]string, 0)
	for endpoint, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, endpoint)
		}
	}

	result := make([]string, 0, len(endpoints))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for endpoint, deps := range dependencies {
			for _, dep := range deps {
				if dep == current {
					inDegree[endpoint]--
					if inDegree[endpoint] == 0 {
						queue = append(queue, endpoint)
					}
				}
			}
		}
	}

	if len(result) != len(endpoints) {
		return nil, fmt.Errorf("circular dependency detected in endpoints")
	}

	return result, nil
}

func extractDependencyData(endpoint string, dependencies map[string][]string, results map[string]interface{}, params map[string]interface{}, resource *DetectedResource) string {
	deps, hasDeps := dependencies[endpoint]
	if !hasDeps {
		return ""
	}

	var resourceOverride string

	for _, dep := range deps {
		depResult, exists := results[dep]
		if !exists {
			continue
		}

		if endpoint == "getRPKIValidation" && dep == "getNetworkInfo" {
			if prefix := extractPrefixFromNetworkInfo(depResult); prefix != "" {
				params["prefix"] = prefix
			}
		}

		if endpoint == "getAddressSpaceHierarchy" && dep == "getNetworkInfo" && resource.Type == IPAddress {
			if prefix := extractPrefixFromNetworkInfo(depResult); prefix != "" {
				resourceOverride = prefix
			}
		}

	}

	return resourceOverride
}

func extractPrefixFromNetworkInfo(result interface{}) string {

	if resultMap, ok := result.(map[string]interface{}); ok {
		if data, ok := resultMap["data"].(map[string]interface{}); ok {
			if prefix, ok := data["prefix"].(string); ok {
				return prefix
			}
		}
	}

	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	if resultValue.Kind() == reflect.Struct {

		dataField := resultValue.FieldByName("Data")
		if dataField.IsValid() && dataField.Kind() == reflect.Struct {
			prefixField := dataField.FieldByName("Prefix")
			if prefixField.IsValid() && prefixField.Kind() == reflect.String {
				return prefixField.String()
			}
		}
	}

	return ""
}
