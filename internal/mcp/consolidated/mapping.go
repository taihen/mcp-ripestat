package consolidated

import "fmt"

// paramToOperationMap defines mappings from parameter values to operations.
var paramToOperationMap = map[string]map[string]Operation{
	"analysis": {
		AnalysisConsistency:      OpConsistency,
		AnalysisPathOptimization: OpRouting,
		AnalysisUpdates:          OpUpdates,
		AnalysisLookingGlass:     OpLookingGlass,
	},
	"data": {
		DataTypeWhois:             OpOverview,
		DataTypeAllocationHistory: OpHistory,
		DataTypeHierarchy:         OpHierarchy,
		DataTypeContacts:          OpSecurity,
	},
	"checks": {
		SecurityCheckRPKI:          OpSecurity,
		SecurityCheckAbuseContacts: OpSecurity,
		SecurityCheckBGPHijacking:  OpSecurity,
	},
	"relationships": {
		RelationshipNeighbors:         OpNeighbors,
		RelationshipAnnouncedPrefixes: OpRouting,
		RelationshipRelatedNetworks:   OpRelationships,
	},
}

// mapStringsToOperations converts string values to Operations using a mapping.
func mapStringsToOperations(values []string, mappingKey string) ([]Operation, error) {
	mapping, exists := paramToOperationMap[mappingKey]
	if !exists {
		return nil, fmt.Errorf("unknown mapping key: %s", mappingKey)
	}

	var operations []Operation
	seen := make(map[Operation]bool)

	for _, value := range values {
		op, exists := mapping[value]
		if !exists {
			return nil, fmt.Errorf("unsupported %s value: %s", mappingKey, value)
		}
		// Deduplicate operations (e.g., multiple security checks map to same operation)
		if !seen[op] {
			operations = append(operations, op)
			seen[op] = true
		}
	}

	return operations, nil
}
