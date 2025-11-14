package consolidated

import (
	"fmt"
)

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

	for i, operation := range operations {
		endpoints, exists := resourceRoutes[operation]
		if !exists {
			return nil, fmt.Errorf("operation '%s' not supported for resource type '%s'", operation, resource.Type.String())
		}

		for _, endpoint := range endpoints {
			if !endpointSet[endpoint] {
				result.Endpoints = append(result.Endpoints, endpoint)
				result.Order = append(result.Order, i)
				endpointSet[endpoint] = true
			}
		}
	}

	result.Dependencies = buildDependencies(result.Endpoints, resource.Type)

	return result, nil
}

func buildDependencies(endpoints []string, _ ResourceType) map[string][]string {
	dependencies := make(map[string][]string)

	endpointSet := make(map[string]bool)
	for _, endpoint := range endpoints {
		endpointSet[endpoint] = true
	}

	if endpointSet["getRPKIValidation"] && endpointSet["getNetworkInfo"] {
		dependencies["getRPKIValidation"] = []string{"getNetworkInfo"}
	}

	if endpointSet["getBGPUpdates"] && endpointSet["getRoutingStatus"] {
		dependencies["getBGPUpdates"] = []string{"getRoutingStatus"}
	}

	if endpointSet["getAddressSpaceHierarchy"] && endpointSet["getNetworkInfo"] {
		dependencies["getAddressSpaceHierarchy"] = []string{"getNetworkInfo"}
	}

	if endpointSet["getRelatedPrefixes"] && endpointSet["getPrefixOverview"] {
		dependencies["getRelatedPrefixes"] = []string{"getPrefixOverview"}
	}

	return dependencies
}

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
