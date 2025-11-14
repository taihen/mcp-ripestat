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

	for {
		result.Dependencies = buildDependencies(result.Endpoints)
		added := addMissingDependencyEndpoints(&result.Endpoints, result.Dependencies, endpointSet)
		if !added {
			break
		}
	}

	return result, nil
}

func addMissingDependencyEndpoints(endpoints *[]string, dependencies map[string][]string, endpointSet map[string]bool) bool {
	requiredDeps := make(map[string]bool)

	for _, deps := range dependencies {
		for _, dep := range deps {
			if !endpointSet[dep] {
				requiredDeps[dep] = true
			}
		}
	}

	for dep := range requiredDeps {
		*endpoints = append(*endpoints, dep)
		endpointSet[dep] = true
	}

	return len(requiredDeps) > 0
}

func buildDependencies(endpoints []string) map[string][]string {
	dependencies := make(map[string][]string)

	endpointSet := make(map[string]bool)
	for _, endpoint := range endpoints {
		endpointSet[endpoint] = true
	}

	if endpointSet["getRPKIValidation"] {
		dependencies["getRPKIValidation"] = []string{"getNetworkInfo"}
	}

	if endpointSet["getBGPUpdates"] {
		dependencies["getBGPUpdates"] = []string{"getRoutingStatus"}
	}

	if endpointSet["getAddressSpaceHierarchy"] {
		dependencies["getAddressSpaceHierarchy"] = []string{"getNetworkInfo"}
	}

	if endpointSet["getRelatedPrefixes"] {
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
