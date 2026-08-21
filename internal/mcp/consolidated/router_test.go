package consolidated

import (
	"reflect"
	"testing"
)

func TestRouteOperations_IncludesDependencyEndpoints(t *testing.T) {
	tests := []struct {
		name          string
		resource      *DetectedResource
		operations    []Operation
		wantEndpoints []string
		wantDeps      map[string][]string
		shouldContain []string
	}{
		{
			name: "OpSecurity for IPAddress includes getNetworkInfo dependency",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations:    []Operation{OpSecurity},
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
			operations:    []Operation{OpRouting},
			shouldContain: []string{"getRoutingStatus", "getBGPUpdates"},
			wantDeps:      map[string][]string{},
		},
		{
			name: "OpHierarchy for IPAddress includes getNetworkInfo dependency",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations:    []Operation{OpHierarchy},
			shouldContain: []string{"getAddressSpaceHierarchy", "getNetworkInfo"},
			wantDeps: map[string][]string{
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
		},
		{
			name: "OpRelationships for IPPrefix routes directly",
			resource: &DetectedResource{
				Type:      IPPrefix,
				Value:     "8.8.8.0/24",
				Validated: true,
			},
			operations:    []Operation{OpRelationships},
			shouldContain: []string{"getRelatedPrefixes"},
			wantDeps:      map[string][]string{},
		},
		{
			name: "Multiple operations with dependencies",
			resource: &DetectedResource{
				Type:      IPAddress,
				Value:     "8.8.8.8",
				Validated: true,
			},
			operations:    []Operation{OpSecurity, OpHierarchy},
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

func TestRouteOperations_NilResource(t *testing.T) {
	_, err := RouteOperations(nil, []Operation{OpOverview})
	if err == nil {
		t.Error("RouteOperations() expected error for nil resource")
	}
}

func TestRouteOperations_EmptyOperations(t *testing.T) {
	resource := &DetectedResource{
		Type:      IPAddress,
		Value:     "8.8.8.8",
		Validated: true,
	}
	_, err := RouteOperations(resource, []Operation{})
	if err == nil {
		t.Error("RouteOperations() expected error for empty operations")
	}
}

func TestRouteOperations_UnsupportedResourceType(t *testing.T) {
	resource := &DetectedResource{
		Type:      Invalid,
		Value:     "invalid",
		Validated: false,
	}
	_, err := RouteOperations(resource, []Operation{OpOverview})
	if err == nil {
		t.Error("RouteOperations() expected error for unsupported resource type")
	}
}

func TestRouteOperations_UnsupportedOperation(t *testing.T) {
	resource := &DetectedResource{
		Type:      IPAddress,
		Value:     "8.8.8.8",
		Validated: true,
	}
	_, err := RouteOperations(resource, []Operation{OpNeighbors})
	if err == nil {
		t.Error("RouteOperations() expected error for unsupported operation")
	}
}

func TestAddMissingDependencyEndpoints(t *testing.T) {
	endpoints := []string{"getRPKIValidation"}
	endpointSet := map[string]bool{"getRPKIValidation": true}
	dependencies := map[string][]string{
		"getRPKIValidation": {"getNetworkInfo"},
	}

	added := addMissingDependencyEndpoints(&endpoints, dependencies, endpointSet)
	if !added {
		t.Error("addMissingDependencyEndpoints() expected to add missing dependency")
	}

	endpointSetCheck := make(map[string]bool)
	for _, ep := range endpoints {
		endpointSetCheck[ep] = true
	}

	if !endpointSetCheck["getNetworkInfo"] {
		t.Error("addMissingDependencyEndpoints() missing getNetworkInfo endpoint")
	}
	if !endpointSetCheck["getRPKIValidation"] {
		t.Error("addMissingDependencyEndpoints() missing getRPKIValidation endpoint")
	}
}

func TestAddMissingDependencyEndpoints_NoMissingDeps(t *testing.T) {
	endpoints := []string{"getRPKIValidation", "getNetworkInfo"}
	endpointSet := map[string]bool{"getRPKIValidation": true, "getNetworkInfo": true}
	dependencies := map[string][]string{
		"getRPKIValidation": {"getNetworkInfo"},
	}

	added := addMissingDependencyEndpoints(&endpoints, dependencies, endpointSet)
	if added {
		t.Error("addMissingDependencyEndpoints() expected no additions")
	}
}

func TestBuildDependencies(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []string
		wantDeps  map[string][]string
	}{
		{
			name:      "getRPKIValidation dependency",
			endpoints: []string{"getRPKIValidation"},
			wantDeps: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
		},
		{
			name:      "getBGPUpdates has no data dependency",
			endpoints: []string{"getBGPUpdates"},
			wantDeps:  map[string][]string{},
		},
		{
			name:      "getAddressSpaceHierarchy requires an explicitly planned dependency",
			endpoints: []string{"getAddressSpaceHierarchy"},
			wantDeps:  map[string][]string{},
		},
		{
			name:      "getRelatedPrefixes has no data dependency",
			endpoints: []string{"getRelatedPrefixes"},
			wantDeps:  map[string][]string{},
		},
		{
			name:      "multiple dependencies",
			endpoints: []string{"getRPKIValidation", "getBGPUpdates"},
			wantDeps: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
		},
		{
			name:      "no dependencies",
			endpoints: []string{"getWhois", "getASOverview"},
			wantDeps:  map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDependencies(tt.endpoints)
			if !reflect.DeepEqual(got, tt.wantDeps) {
				t.Errorf("buildDependencies() = %v, want %v", got, tt.wantDeps)
			}
		})
	}
}

func TestGetSupportedOperations(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		wantEmpty    bool
	}{
		{
			name:         "IPAddress",
			resourceType: IPAddress,
			wantEmpty:    false,
		},
		{
			name:         "IPPrefix",
			resourceType: IPPrefix,
			wantEmpty:    false,
		},
		{
			name:         "ASN",
			resourceType: ASN,
			wantEmpty:    false,
		},
		{
			name:         "Country",
			resourceType: Country,
			wantEmpty:    false,
		},
		{
			name:         "Invalid",
			resourceType: Invalid,
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSupportedOperations(tt.resourceType)
			if (len(got) == 0) != tt.wantEmpty {
				t.Errorf("GetSupportedOperations() = %v, wantEmpty %v", got, tt.wantEmpty)
			}
		})
	}
}

func TestValidateOperations(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		operations   []Operation
		wantErr      bool
	}{
		{
			name:         "valid operations for IPAddress",
			resourceType: IPAddress,
			operations:   []Operation{OpOverview, OpSecurity},
			wantErr:      false,
		},
		{
			name:         "invalid operation for IPAddress",
			resourceType: IPAddress,
			operations:   []Operation{OpNeighbors},
			wantErr:      true,
		},
		{
			name:         "valid operations for ASN",
			resourceType: ASN,
			operations:   []Operation{OpOverview, OpNeighbors},
			wantErr:      false,
		},
		{
			name:         "invalid operation for ASN",
			resourceType: ASN,
			operations:   []Operation{OpSecurity},
			wantErr:      true,
		},
		{
			name:         "invalid resource type",
			resourceType: Invalid,
			operations:   []Operation{OpOverview},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOperations(tt.resourceType, tt.operations)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOperations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEndpointsForOperation(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		operation    Operation
		wantErr      bool
		wantEmpty    bool
	}{
		{
			name:         "valid operation for IPAddress",
			resourceType: IPAddress,
			operation:    OpOverview,
			wantErr:      false,
			wantEmpty:    false,
		},
		{
			name:         "invalid operation for IPAddress",
			resourceType: IPAddress,
			operation:    OpNeighbors,
			wantErr:      true,
			wantEmpty:    true,
		},
		{
			name:         "valid operation for ASN",
			resourceType: ASN,
			operation:    OpNeighbors,
			wantErr:      false,
			wantEmpty:    false,
		},
		{
			name:         "invalid resource type",
			resourceType: Invalid,
			operation:    OpOverview,
			wantErr:      true,
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEndpointsForOperation(tt.resourceType, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpointsForOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(got) == 0) != tt.wantEmpty {
				t.Errorf("GetEndpointsForOperation() = %v, wantEmpty %v", got, tt.wantEmpty)
			}
		})
	}
}
