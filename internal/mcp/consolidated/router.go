package consolidated

import (
	"fmt"
)

// routingMatrix defines which endpoints to call for each resource type and operation combination.
var routingMatrix = map[ResourceType]map[Operation][]string{
	IPAddress: {
		OpOverview:     {"getNetworkInfo", "getWhois"},
		OpSecurity:     {"getAbuseContactFinder", "getRPKIValidation"},
		OpRouting:      {"getRoutingStatus", "getBGPUpdates"},
		OpHistory:      {"getRoutingHistory", "getAllocationHistory"},
		OpHierarchy:    {"getAddressSpaceHierarchy"},
		OpUpdates:      {"getBGPUpdates"},
		OpLookingGlass: {"getLookingGlass", "getBGPState"},
	},
	IPPrefix: {
		OpOverview:      {"getPrefixOverview", "getNetworkInfo", "getWhois"},
		OpSecurity:      {"getAbuseContactFinder", "getRPKIValidation", "getRPKIHistory"},
		OpRouting:       {"getRoutingStatus", "getPrefixRoutingConsistency"},
		OpHistory:       {"getRoutingHistory", "getAllocationHistory"},
		OpConsistency:   {"getPrefixRoutingConsistency"},
		OpRelationships: {"getRelatedPrefixes"},
		OpHierarchy:     {"getAddressSpaceHierarchy"},
		OpUpdates:       {"getBGPUpdates"},
		OpLookingGlass:  {"getLookingGlass", "getBGPState"},
	},
	ASN: {
		OpOverview:      {"getASOverview", "getWhois"},
		OpRouting:       {"getAnnouncedPrefixes", "getASRoutingConsistency", "getASPathLength"},
		OpNeighbors:     {"getASNNeighbours"},
		OpConsistency:   {"getASRoutingConsistency"},
		OpRelationships: {"getASNNeighbours", "getAnnouncedPrefixes"},
		OpHistory:       {"getRoutingHistory"},
	},
	Country: {
		OpOverview: {"getCountryASNs"},
	},
}

// operationPriority defines the execution order for operations (lower number = higher priority).
var operationPriority = map[Operation]int{
	OpOverview:      1,
	OpSecurity:      2,
	OpRouting:       3,
	OpNeighbors:     4,
	OpRelationships: 5,
	OpHistory:       6,
	OpConsistency:   7,
	OpUpdates:       8,
	OpLookingGlass:  9,
	OpHierarchy:     10,
}

// RouteOperations determines which endpoints to call for given operations on a resource.
func RouteOperations(resource *DetectedResource, operations []Operation) (*RouteResult, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource cannot be nil")
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("operations cannot be empty")
	}

	resourceRoutes, exists := routingMatrix[resource.Type]
	if !exists {
		return nil, fmt.Errorf("no routing available for resource type: %s", resource.Type.String())
	}

	result := &RouteResult{
		Endpoints:    []string{},
		Order:        []int{},
		Dependencies: make(map[string][]string),
	}

	endpointSet := make(map[string]bool)
	operationOrder := make(map[Operation]int)

	// Collect all endpoints for the requested operations
	for i, operation := range operations {
		endpoints, exists := resourceRoutes[operation]
		if !exists {
			return nil, fmt.Errorf("operation '%s' not supported for resource type '%s'", operation, resource.Type.String())
		}

		operationOrder[operation] = operationPriority[operation]

		for _, endpoint := range endpoints {
			if !endpointSet[endpoint] {
				result.Endpoints = append(result.Endpoints, endpoint)
				result.Order = append(result.Order, i)
				endpointSet[endpoint] = true
			}
		}
	}

	// Set up dependencies for certain endpoint combinations
	result.Dependencies = buildDependencies(result.Endpoints, resource.Type)

	return result, nil
}

// buildDependencies sets up execution dependencies between endpoints.
func buildDependencies(endpoints []string, _ ResourceType) map[string][]string {
	dependencies := make(map[string][]string)

	// Some endpoints should run after others for optimal data aggregation
	endpointSet := make(map[string]bool)
	for _, endpoint := range endpoints {
		endpointSet[endpoint] = true
	}

	// RPKI validation should come after network info for context
	if endpointSet["getRPKIValidation"] && endpointSet["getNetworkInfo"] {
		dependencies["getRPKIValidation"] = []string{"getNetworkInfo"}
	}

	// BGP updates should come after routing status for context
	if endpointSet["getBGPUpdates"] && endpointSet["getRoutingStatus"] {
		dependencies["getBGPUpdates"] = []string{"getRoutingStatus"}
	}

	// Address space hierarchy should come after network info
	if endpointSet["getAddressSpaceHierarchy"] && endpointSet["getNetworkInfo"] {
		dependencies["getAddressSpaceHierarchy"] = []string{"getNetworkInfo"}
	}

	// Related prefixes should come after prefix overview
	if endpointSet["getRelatedPrefixes"] && endpointSet["getPrefixOverview"] {
		dependencies["getRelatedPrefixes"] = []string{"getPrefixOverview"}
	}

	return dependencies
}

// GetSupportedOperations returns the operations supported for a given resource type.
func GetSupportedOperations(resourceType ResourceType) []Operation {
	resourceRoutes, exists := routingMatrix[resourceType]
	if !exists {
		return []Operation{}
	}

	operations := make([]Operation, 0, len(resourceRoutes))
	for operation := range resourceRoutes {
		operations = append(operations, operation)
	}

	return operations
}

// ValidateOperations checks if all operations are supported for the given resource type.
func ValidateOperations(resourceType ResourceType, operations []Operation) error {
	supportedOps := GetSupportedOperations(resourceType)
	supportedSet := make(map[Operation]bool)
	for _, op := range supportedOps {
		supportedSet[op] = true
	}

	for _, operation := range operations {
		if !supportedSet[operation] {
			return fmt.Errorf("operation '%s' not supported for resource type '%s'", operation, resourceType.String())
		}
	}

	return nil
}

// GetEndpointsForOperation returns the specific endpoints that will be called for an operation.
func GetEndpointsForOperation(resourceType ResourceType, operation Operation) ([]string, error) {
	resourceRoutes, exists := routingMatrix[resourceType]
	if !exists {
		return nil, fmt.Errorf("no routing available for resource type: %s", resourceType.String())
	}

	endpoints, exists := resourceRoutes[operation]
	if !exists {
		return nil, fmt.Errorf("operation '%s' not supported for resource type '%s'", operation, resourceType.String())
	}

	return endpoints, nil
}
