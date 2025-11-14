package consolidated

import "fmt"

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
		if !seen[op] {
			operations = append(operations, op)
			seen[op] = true
		}
	}

	return operations, nil
}
