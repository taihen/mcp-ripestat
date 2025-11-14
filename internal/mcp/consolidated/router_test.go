package consolidated

import (
	"reflect"
	"testing"
)

func TestRouteOperations_IncludesDependencyEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		resource       *DetectedResource
		operations     []Operation
		wantEndpoints  []string
		wantDeps       map[string][]string
		shouldContain  []string // Endpoints that should be in the result
	}{
		{
			name: "OpSecurity for IPAddress includes getNetworkInfo dependency",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations: []Operation{OpSecurity},
			shouldContain: []string{"getAbuseContactFinder", "getRPKIValidation", "getNetworkInfo"},
			wantDeps: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
		},
		{
			name: "OpRouting for IPAddress includes getRoutingStatus dependency",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations: []Operation{OpRouting},
			shouldContain: []string{"getRoutingStatus", "getBGPUpdates"},
			wantDeps: map[string][]string{
				"getBGPUpdates": {"getRoutingStatus"},
			},
		},
		{
			name: "OpHierarchy for IPAddress includes getNetworkInfo dependency",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations: []Operation{OpHierarchy},
			shouldContain: []string{"getAddressSpaceHierarchy", "getNetworkInfo"},
			wantDeps: map[string][]string{
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
		},
		{
			name: "OpRelationships for IPPrefix includes getPrefixOverview dependency",
			resource: &DetectedResource{
				Type:      IPPrefix,
				Value:     "8.8.8.0/24",
				Validated: true,
			},
			operations: []Operation{OpRelationships},
			shouldContain: []string{"getRelatedPrefixes", "getPrefixOverview"},
			wantDeps: map[string][]string{
				"getRelatedPrefixes": {"getPrefixOverview"},
			},
		},
		{
			name: "Multiple operations with dependencies",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations: []Operation{OpSecurity, OpHierarchy},
			shouldContain: []string{"getAbuseContactFinder", "getRPKIValidation", "getNetworkInfo", "getAddressSpaceHierarchy"},
			wantDeps: map[string][]string{
				"getRPKIValidation":        {"getNetworkInfo"},
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RouteOperations(tt.resource, tt.operations)
			if err != nil {
				t.Fatalf("RouteOperations() error = %v", err)
			}

			endpointSet := make(map[string]bool)
			for _, endpoint := range result.Endpoints {
				endpointSet[endpoint] = true
			}

			for _, endpoint := range tt.shouldContain {
				if !endpointSet[endpoint] {
					t.Errorf("RouteOperations() missing required endpoint: %s. Got endpoints: %v", endpoint, result.Endpoints)
				}
			}

			if !reflect.DeepEqual(result.Dependencies, tt.wantDeps) {
				t.Errorf("RouteOperations() Dependencies = %v, want %v", result.Dependencies, tt.wantDeps)
			}
		})
	}
}

func TestRouteOperations_NoDuplicateDependencyEndpoints(t *testing.T) {
	resource := &DetectedResource{
		Type:      IPAddress,
		Value:     "8.8.8.8",
		Validated: true,
	}

	result, err := RouteOperations(resource, []Operation{OpOverview, OpSecurity})
	if err != nil {
		t.Fatalf("RouteOperations() error = %v", err)
	}

	count := 0
	for _, endpoint := range result.Endpoints {
		if endpoint == "getNetworkInfo" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("RouteOperations() getNetworkInfo appears %d times, want 1. Endpoints: %v", count, result.Endpoints)
	}
}

