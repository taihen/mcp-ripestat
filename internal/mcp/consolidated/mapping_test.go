package consolidated

import (
	"reflect"
	"testing"
)

func TestMapStringsToOperations(t *testing.T) {
	tests := []struct {
		name       string
		values     []string
		mappingKey string
		want       []Operation
		wantErr    bool
		errSubstr  string
	}{
		{
			name:       "analysis - consistency",
			values:     []string{AnalysisConsistency},
			mappingKey: "analysis",
			want:       []Operation{OpConsistency},
			wantErr:    false,
		},
		{
			name:       "analysis - multiple types",
			values:     []string{AnalysisConsistency, AnalysisPathOptimization, AnalysisUpdates},
			mappingKey: "analysis",
			want:       []Operation{OpConsistency, OpRouting, OpUpdates},
			wantErr:    false,
		},
		{
			name:       "data - whois",
			values:     []string{DataTypeWhois},
			mappingKey: "data",
			want:       []Operation{OpOverview},
			wantErr:    false,
		},
		{
			name:       "data - all types",
			values:     []string{DataTypeWhois, DataTypeAllocationHistory, DataTypeHierarchy, DataTypeContacts},
			mappingKey: "data",
			want:       []Operation{OpOverview, OpHistory, OpHierarchy, OpSecurity},
			wantErr:    false,
		},
		{
			name:       "checks - single check",
			values:     []string{SecurityCheckRPKI},
			mappingKey: "checks",
			want:       []Operation{OpSecurity},
			wantErr:    false,
		},
		{
			name:       "checks - deduplication",
			values:     []string{SecurityCheckRPKI, SecurityCheckAbuseContacts, SecurityCheckBGPHijacking},
			mappingKey: "checks",
			want:       []Operation{OpSecurity}, // All map to same operation, should deduplicate
			wantErr:    false,
		},
		{
			name:       "relationships - all types",
			values:     []string{RelationshipNeighbors, RelationshipAnnouncedPrefixes, RelationshipRelatedNetworks},
			mappingKey: "relationships",
			want:       []Operation{OpNeighbors, OpRouting, OpRelationships},
			wantErr:    false,
		},
		{
			name:       "invalid mapping key",
			values:     []string{"value"},
			mappingKey: "invalid",
			want:       nil,
			wantErr:    true,
			errSubstr:  "unknown mapping key",
		},
		{
			name:       "invalid value for analysis",
			values:     []string{"invalid-analysis"},
			mappingKey: "analysis",
			want:       nil,
			wantErr:    true,
			errSubstr:  "unsupported",
		},
		{
			name:       "empty values",
			values:     []string{},
			mappingKey: "analysis",
			want:       []Operation{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapStringsToOperations(tt.values, tt.mappingKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapStringsToOperations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Handle nil vs empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapStringsToOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParamToOperationMapCompleteness(t *testing.T) {
	// Verify all constants are in the mapping
	t.Run("analysis constants", func(t *testing.T) {
		mapping := paramToOperationMap["analysis"]
		constants := []string{
			AnalysisConsistency,
			AnalysisPathOptimization,
			AnalysisUpdates,
			AnalysisLookingGlass,
		}
		for _, c := range constants {
			if _, exists := mapping[c]; !exists {
				t.Errorf("analysis constant %q not in mapping", c)
			}
		}
	})

	t.Run("data constants", func(t *testing.T) {
		mapping := paramToOperationMap["data"]
		constants := []string{
			DataTypeWhois,
			DataTypeAllocationHistory,
			DataTypeHierarchy,
			DataTypeContacts,
		}
		for _, c := range constants {
			if _, exists := mapping[c]; !exists {
				t.Errorf("data constant %q not in mapping", c)
			}
		}
	})

	t.Run("checks constants", func(t *testing.T) {
		mapping := paramToOperationMap["checks"]
		constants := []string{
			SecurityCheckRPKI,
			SecurityCheckAbuseContacts,
			SecurityCheckBGPHijacking,
		}
		for _, c := range constants {
			if _, exists := mapping[c]; !exists {
				t.Errorf("checks constant %q not in mapping", c)
			}
		}
	})

	t.Run("relationships constants", func(t *testing.T) {
		mapping := paramToOperationMap["relationships"]
		constants := []string{
			RelationshipNeighbors,
			RelationshipAnnouncedPrefixes,
			RelationshipRelatedNetworks,
		}
		for _, c := range constants {
			if _, exists := mapping[c]; !exists {
				t.Errorf("relationships constant %q not in mapping", c)
			}
		}
	})
}
