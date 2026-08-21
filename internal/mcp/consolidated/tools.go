package consolidated

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	baseParams map[string]interface{},
) (*Result, error) {
	detected, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	routes, err := RouteOperations(detected, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to route operations: %w", err)
	}

	return ct.executeAndAggregate(ctx, detected, operations, routes, depth, baseParams)
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
	if err := validateEnum(depth, "depth", DepthBasic, DepthDetailed, DepthComprehensive); err != nil {
		return nil, err
	}

	return ct.executeOperations(ctx, resource, toOperations(operationStrings), depth, nil)
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
	if err := validateEnum(timeframe, "timeframe", TimeframeCurrent); err != nil {
		return nil, err
	}
	if timeframe != TimeframeCurrent && !containsString(analysisTypes, AnalysisUpdates) {
		return nil, fmt.Errorf("historical timeframe requires the %q analysis type", AnalysisUpdates)
	}

	result, err := ct.executeOperations(ctx, resource, operations, DepthDetailed, map[string]interface{}{
		"timeframe": timeframe,
	})
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
	if err := validateEnum(format, "format", FormatSummary); err != nil {
		return nil, err
	}

	depth := DepthBasic
	result, err := ct.executeOperations(ctx, resource, operations, depth, nil)
	if err != nil {
		return nil, err
	}
	result.AddMetadata("format", format)
	return result, nil
}

func (ct *Tools) ValidateSecurity(ctx context.Context, params map[string]interface{}) (*Result, error) {
	resource, err := extractRequiredString(params, "resource")
	if err != nil {
		return nil, err
	}
	detectedResource, err := DetectResource(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource type: %w", err)
	}

	defaultChecks := []string{SecurityCheckRPKI, SecurityCheckAbuseContacts}
	if detectedResource.Type == ASN {
		defaultChecks = []string{SecurityCheckAbuseContacts, SecurityCheckBGPHijacking}
	}

	checkTypes, err := extractStringSliceWithDefault(
		params,
		"checks",
		defaultChecks,
	)
	if err != nil {
		return nil, err
	}
	if detectedResource.Type == ASN && containsString(checkTypes, SecurityCheckRPKI) {
		return nil, fmt.Errorf("RPKI validation requires an IP address or prefix; ASN-wide current validation is not implemented")
	}

	operations, err := mapStringsToOperations(checkTypes, "checks")
	if err != nil {
		return nil, err
	}

	baseParams := make(map[string]interface{})
	asnParam := extractOptionalString(params, "asn", "")
	if asnParam != "" {
		if !containsString(checkTypes, SecurityCheckRPKI) {
			return nil, fmt.Errorf("asn parameter requires the %q security check", SecurityCheckRPKI)
		}
		if detectedResource.Type == ASN {
			return nil, fmt.Errorf("asn parameter is only applicable when validating an IP address or prefix")
		}
		detectedASN, detectErr := DetectResource(asnParam)
		if detectErr != nil || detectedASN.Type != ASN {
			return nil, fmt.Errorf("asn parameter must be a valid ASN")
		}
		asnParam = detectedASN.Value
		baseParams["asn"] = asnParam
	}

	result, err := ct.executeOperations(ctx, resource, operations, DepthDetailed, baseParams)
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

	defaultRelationships := []string{RelationshipNeighbors}
	if _, exists := params["relationships"]; !exists {
		detected, detectErr := DetectResource(resource)
		if detectErr == nil && detected.Type == IPPrefix {
			defaultRelationships = []string{RelationshipRelatedNetworks}
		}
	}

	relationshipTypes, err := extractStringSliceWithDefault(params, "relationships", defaultRelationships)
	if err != nil {
		return nil, err
	}

	operations, err := mapStringsToOperations(relationshipTypes, "relationships")
	if err != nil {
		return nil, err
	}

	scope := extractOptionalString(params, "scope", ScopeDirect)
	if err := validateEnum(scope, "scope", ScopeDirect); err != nil {
		return nil, err
	}

	result, err := ct.executeOperations(ctx, resource, operations, DepthBasic, nil)
	if err != nil {
		return nil, err
	}

	result.AddMetadata("scope", scope)
	return result, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (ct *Tools) SearchByLocation(ctx context.Context, params map[string]interface{}) (*Result, error) {
	country, err := extractRequiredString(params, "country")
	if err != nil {
		return nil, err
	}

	typeParam := extractOptionalString(params, "type", LocationTypeASNs)

	if typeParam != LocationTypeASNs {
		return nil, fmt.Errorf("unsupported location type %q: only %q is implemented", typeParam, LocationTypeASNs)
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

	return ct.executeAndAggregate(ctx, detected, []Operation{OpOverview}, routes, DepthBasic, nil)
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
	baseParams map[string]interface{},
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
	attemptedEndpoints := make([]string, 0, len(sortedEndpoints))
	succeededEndpoints := make([]string, 0, len(sortedEndpoints))

	for i, endpoint := range sortedEndpoints {
		// Check for context cancellation before each endpoint call
		if ctx.Err() != nil {
			// Mark this and all remaining endpoints as cancelled
			for _, remaining := range sortedEndpoints[i:] {
				result.Errors[remaining] = "cancelled"
			}
			break
		}

		endpointParams := paramsForEndpoint(endpoint, baseParams)
		for key, value := range translateDepthToEndpointParams(endpoint, depth) {
			endpointParams[key] = value
		}

		resourceValue, prepareErr := prepareEndpointCall(
			endpoint,
			routes.Dependencies,
			result.Results,
			endpointParams,
			resource,
		)
		if prepareErr != nil {
			result.Errors[endpoint] = prepareErr.Error()
			continue
		}

		endpointResult, err := ct.executor.ExecuteEndpoint(ctx, endpoint, resourceValue, endpointParams)
		attemptedEndpoints = append(attemptedEndpoints, endpoint)
		if err != nil {
			result.Errors[endpoint] = err.Error()
			continue
		}

		result.Results[endpoint] = endpointResult
		succeededEndpoints = append(succeededEndpoints, endpoint)
	}

	complete := len(succeededEndpoints) == len(sortedEndpoints)
	result.AddMetadataMap(map[string]interface{}{
		"endpoints_called":    attemptedEndpoints,
		"endpoints_planned":   sortedEndpoints,
		"endpoints_attempted": attemptedEndpoints,
		"endpoints_succeeded": succeededEndpoints,
		"complete":            complete,
		"depth":               depth,
		"resource_type":       resource.Type.String(),
	})
	if len(succeededEndpoints) == 0 && len(sortedEndpoints) > 0 {
		return result, fmt.Errorf("all planned endpoint calls failed")
	}

	return result, nil
}

func copyParams(params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(params))
	for key, value := range params {
		result[key] = value
	}
	return result
}

func paramsForEndpoint(endpoint string, baseParams map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if endpoint == "getRPKIValidation" {
		if asn, ok := baseParams["asn"]; ok {
			result["asn"] = asn
		}
	}
	if endpoint == "getBGPUpdates" {
		timeframe, ok := baseParams["timeframe"]
		if !ok {
			timeframe = TimeframeCurrent
		}
		result["timeframe"] = timeframe
	}
	return result
}

func prepareEndpointCall(
	endpoint string,
	dependencies map[string][]string,
	results map[string]interface{},
	params map[string]interface{},
	resource *DetectedResource,
) (string, error) {
	resourceValue := resource.Value
	deps := dependencies[endpoint]

	for _, dep := range deps {
		depResult, exists := results[dep]
		if !exists {
			return "", fmt.Errorf("cannot execute %s: dependency %s did not succeed", endpoint, dep)
		}

		if endpoint == "getRPKIValidation" && dep == "getNetworkInfo" {
			prefix := resource.Value
			if resource.Type == IPAddress {
				prefix = extractPrefixFromNetworkInfo(depResult)
			}
			if prefix == "" {
				return "", fmt.Errorf("cannot validate RPKI: network information did not include a prefix")
			}
			params["prefix"] = prefix

			asn := getOptionalStringParam(params, "asn")
			if asn == "" {
				asns := extractASNsFromNetworkInfo(depResult)
				switch len(asns) {
				case 0:
					return "", fmt.Errorf("cannot validate RPKI: network information did not include an origin ASN")
				case 1:
					asn = asns[0]
				default:
					return "", fmt.Errorf("cannot validate RPKI for a multi-origin prefix without an explicit asn parameter")
				}
			}

			normalizedASN, err := normalizeASN(asn)
			if err != nil {
				return "", err
			}
			resourceValue = normalizedASN
			delete(params, "asn")
		}

		if endpoint == "getAddressSpaceHierarchy" && dep == "getNetworkInfo" && resource.Type == IPAddress {
			if prefix := extractPrefixFromNetworkInfo(depResult); prefix != "" {
				resourceValue = prefix
			}
		}

		if endpoint == "getPrefixRoutingConsistency" && dep == "getNetworkInfo" && resource.Type == IPAddress {
			prefix := extractPrefixFromNetworkInfo(depResult)
			if prefix == "" {
				return "", fmt.Errorf("cannot check prefix routing consistency: network information did not include a prefix")
			}
			resourceValue = prefix
		}
	}

	return resourceValue, nil
}

func normalizeASN(value string) (string, error) {
	if _, err := strconv.ParseUint(value, 10, 32); err == nil {
		value = "AS" + value
	}
	detected, err := DetectResource(value)
	if err != nil || detected.Type != ASN {
		return "", fmt.Errorf("invalid ASN %q for RPKI validation", value)
	}
	return detected.Value, nil
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
	for _, endpoint := range endpoints {
		if inDegree[endpoint] == 0 {
			queue = append(queue, endpoint)
		}
	}

	result := make([]string, 0, len(endpoints))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, endpoint := range endpoints {
			deps := dependencies[endpoint]
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

func extractPrefixFromNetworkInfo(result interface{}) string {

	if resultMap, ok := result.(map[string]interface{}); ok {
		if data, ok := resultMap["data"].(map[string]interface{}); ok {
			if prefix, ok := data["prefix"].(string); ok {
				return prefix
			}
		}
	}

	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Pointer {
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

func extractASNsFromNetworkInfo(result interface{}) []string {
	var values []interface{}
	if resultMap, ok := result.(map[string]interface{}); ok {
		if data, ok := resultMap["data"].(map[string]interface{}); ok {
			values, _ = data["asns"].([]interface{})
		}
	}

	if values == nil {
		resultValue := reflect.ValueOf(result)
		if resultValue.IsValid() && resultValue.Kind() == reflect.Pointer && !resultValue.IsNil() {
			resultValue = resultValue.Elem()
		}
		if resultValue.IsValid() && resultValue.Kind() == reflect.Struct {
			dataField := resultValue.FieldByName("Data")
			if dataField.IsValid() && dataField.Kind() == reflect.Struct {
				asnsField := dataField.FieldByName("ASNs")
				if asnsField.IsValid() && asnsField.Kind() == reflect.Slice {
					values = make([]interface{}, 0, asnsField.Len())
					for i := 0; i < asnsField.Len(); i++ {
						values = append(values, asnsField.Index(i).Interface())
					}
				}
			}
		}
	}

	asns := make([]string, 0, len(values))
	for _, value := range values {
		candidate := strings.TrimSpace(fmt.Sprint(value))
		if candidate == "" {
			continue
		}
		if normalized, err := normalizeASN(candidate); err == nil {
			asns = append(asns, normalized)
		}
	}
	return asns
}
