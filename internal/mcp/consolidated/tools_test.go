package consolidated

import (
	"context"
	"reflect"
	"testing"
)

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name         string
		endpoints    []string
		dependencies map[string][]string
		want         []string
		wantErr      bool
		verifyDeps   bool
	}{
		{
			name:         "no dependencies",
			endpoints:    []string{"getWhois", "getASOverview"},
			dependencies: map[string][]string{},
			want:         []string{"getWhois", "getASOverview"},
			wantErr:      false,
		},
		{
			name:      "single dependency",
			endpoints: []string{"getRPKIValidation", "getNetworkInfo"},
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			want:    []string{"getNetworkInfo", "getRPKIValidation"},
			wantErr: false,
		},
		{
			name:      "multiple dependencies",
			endpoints: []string{"getRPKIValidation", "getBGPUpdates", "getNetworkInfo", "getRoutingStatus"},
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
				"getBGPUpdates":     {"getRoutingStatus"},
			},
			want:       nil,
			wantErr:    false,
			verifyDeps: true,
		},
		{
			name:      "transitive dependencies",
			endpoints: []string{"getBGPUpdates", "getRoutingStatus", "getNetworkInfo"},
			dependencies: map[string][]string{
				"getBGPUpdates":    {"getRoutingStatus"},
				"getRoutingStatus": {"getNetworkInfo"},
			},
			want:    []string{"getNetworkInfo", "getRoutingStatus", "getBGPUpdates"},
			wantErr: false,
		},
		{
			name:      "circular dependency",
			endpoints: []string{"endpoint1", "endpoint2"},
			dependencies: map[string][]string{
				"endpoint1": {"endpoint2"},
				"endpoint2": {"endpoint1"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "dependency not in endpoints",
			endpoints: []string{"getRPKIValidation"},
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			want:    []string{"getRPKIValidation"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := topologicalSort(tt.endpoints, tt.dependencies)
			if (err != nil) != tt.wantErr {
				t.Errorf("topologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tt.verifyDeps {
				if len(got) != len(tt.endpoints) {
					t.Errorf("topologicalSort() length = %d, want %d", len(got), len(tt.endpoints))
					return
				}

				position := make(map[string]int)
				for i, ep := range got {
					position[ep] = i
				}

				for dependent, deps := range tt.dependencies {
					depPos, depExists := position[dependent]
					if !depExists {
						continue
					}

					for _, dep := range deps {
						depDepPos, depDepExists := position[dep]
						if !depDepExists {
							continue
						}

						if depPos <= depDepPos {
							t.Errorf("topologicalSort() dependency violation: %s (pos %d) should come after %s (pos %d)", dependent, depPos, dep, depDepPos)
						}
					}
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("topologicalSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPrefixFromNetworkInfo(t *testing.T) {
	tests := []struct {
		name      string
		result    interface{}
		want      string
		wantEmpty bool
	}{
		{
			name: "map format with data.prefix",
			result: map[string]interface{}{
				"data": map[string]interface{}{
					"prefix": "8.8.8.0/24",
				},
			},
			want:      "8.8.8.0/24",
			wantEmpty: false,
		},
		{
			name: "map format without prefix",
			result: map[string]interface{}{
				"data": map[string]interface{}{
					"asns": []interface{}{},
				},
			},
			want:      "",
			wantEmpty: true,
		},
		{
			name:      "nil result",
			result:    nil,
			want:      "",
			wantEmpty: true,
		},
		{
			name:      "non-map result",
			result:    "string",
			want:      "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPrefixFromNetworkInfo(tt.result)
			if (got == "") != tt.wantEmpty {
				t.Errorf("extractPrefixFromNetworkInfo() = %v, wantEmpty %v", got, tt.wantEmpty)
			}
			if got != tt.want {
				t.Errorf("extractPrefixFromNetworkInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDependencyData(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		dependencies map[string][]string
		results      map[string]interface{}
		resource     *DetectedResource
		wantParams   map[string]interface{}
		wantOverride string
	}{
		{
			name:     "getRPKIValidation extracts prefix",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			wantParams: map[string]interface{}{
				"prefix": "8.8.8.0/24",
			},
			wantOverride: "",
		},
		{
			name:     "getAddressSpaceHierarchy extracts prefix for IP address",
			endpoint: "getAddressSpaceHierarchy",
			dependencies: map[string][]string{
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			wantParams:   map[string]interface{}{},
			wantOverride: "8.8.8.0/24",
		},
		{
			name:     "getAddressSpaceHierarchy with IPPrefix does not override",
			endpoint: "getAddressSpaceHierarchy",
			dependencies: map[string][]string{
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPPrefix,
				Value: "8.8.8.0/24",
			},
			wantParams:   map[string]interface{}{},
			wantOverride: "",
		},
		{
			name:         "no dependencies",
			endpoint:     "getWhois",
			dependencies: map[string][]string{},
			results:      map[string]interface{}{},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			wantParams:   map[string]interface{}{},
			wantOverride: "",
		},
		{
			name:     "dependency result not available",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			wantParams:   map[string]interface{}{},
			wantOverride: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]interface{})
			override := extractDependencyData(tt.endpoint, tt.dependencies, tt.results, params, tt.resource)

			if override != tt.wantOverride {
				t.Errorf("extractDependencyData() override = %v, want %v", override, tt.wantOverride)
			}

			if !reflect.DeepEqual(params, tt.wantParams) {
				t.Errorf("extractDependencyData() params = %v, want %v", params, tt.wantParams)
			}
		})
	}
}

type mockExecutor struct {
	results map[string]interface{}
	errors  map[string]error
}

func (m *mockExecutor) ExecuteEndpoint(_ context.Context, endpoint string, _ string, _ map[string]interface{}) (interface{}, error) {
	if err, ok := m.errors[endpoint]; ok {
		return nil, err
	}
	if result, ok := m.results[endpoint]; ok {
		return result, nil
	}
	return nil, nil
}

func testExecuteAndAggregateHelper(t *testing.T, endpoint string, operation Operation, expectedEndpoints []string) {
	t.Helper()
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getNetworkInfo": map[string]interface{}{
				"data": map[string]interface{}{
					"prefix": "8.8.8.0/24",
				},
			},
			endpoint: map[string]interface{}{
				"status": "ok",
			},
		},
		errors: map[string]error{},
	}

	tools := NewTools(executor)
	resource := &DetectedResource{
		Type:  IPAddress,
		Value: "8.8.8.8",
	}

	routes := &RouteResult{
		Endpoints: []string{endpoint, "getNetworkInfo"},
		Dependencies: map[string][]string{
			endpoint: {"getNetworkInfo"},
		},
	}

	ctx := context.Background()
	result, err := tools.executeAndAggregate(ctx, resource, []Operation{operation}, routes, DepthBasic)
	if err != nil {
		t.Fatalf("executeAndAggregate() error = %v", err)
	}

	for _, expectedEndpoint := range expectedEndpoints {
		if _, ok := result.Results[expectedEndpoint]; !ok {
			t.Errorf("executeAndAggregate() missing %s result", expectedEndpoint)
		}
	}
}

func TestExecuteAndAggregate_DependencyOrder(t *testing.T) {
	testExecuteAndAggregateHelper(t, "getRPKIValidation", OpSecurity, []string{"getNetworkInfo", "getRPKIValidation"})
}

func TestExecuteAndAggregate_ResourceOverride(t *testing.T) {
	testExecuteAndAggregateHelper(t, "getAddressSpaceHierarchy", OpHierarchy, []string{"getNetworkInfo", "getAddressSpaceHierarchy"})
}

func TestNewTools(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{},
		errors:  map[string]error{},
	}
	tools := NewTools(executor)
	if tools == nil {
		t.Error("NewTools() returned nil")
	}
	if tools.executor != executor {
		t.Error("NewTools() executor mismatch")
	}
}

func TestTools_InvestigateResource(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getNetworkInfo": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params with resource",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
			},
			expectError: false,
		},
		{
			name: "valid params with operations",
			params: map[string]interface{}{
				"resource":   "8.8.8.8",
				"operations": []string{"overview", "security"},
			},
			expectError: false,
		},
		{
			name: "valid params with depth",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
				"depth":    DepthDetailed,
			},
			expectError: false,
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"operations": []string{"overview"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tools.InvestigateResource(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("InvestigateResource() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestTools_AnalyzeRouting(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getPrefixRoutingConsistency": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
			},
			expectError: false,
		},
		{
			name: "valid params with analysis",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
				"analysis": []string{"consistency"},
			},
			expectError: false,
		},
		{
			name: "valid params with timeframe",
			params: map[string]interface{}{
				"resource":  "8.8.8.0/24",
				"timeframe": Timeframe1Week,
			},
			expectError: false,
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"analysis": []string{"consistency"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tools.AnalyzeRouting(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("AnalyzeRouting() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result != nil {
				if timeframe, ok := result.Metadata["timeframe"]; ok && tt.params["timeframe"] != nil {
					if timeframe != tt.params["timeframe"] {
						t.Errorf("AnalyzeRouting() timeframe = %v, want %v", timeframe, tt.params["timeframe"])
					}
				}
			}
		})
	}
}

func TestTools_QueryRegistry(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getWhois": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
			},
			expectError: false,
		},
		{
			name: "valid params with data",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
				"data":     []string{"whois"},
			},
			expectError: false,
		},
		{
			name: "valid params with format summary",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
				"format":   FormatSummary,
			},
			expectError: false,
		},
		{
			name: "valid params with format detailed",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
				"format":   FormatDetailed,
			},
			expectError: false,
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"data": []string{"whois"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tools.QueryRegistry(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("QueryRegistry() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestTools_ValidateSecurity(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getRPKIValidation": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
			},
			expectError: false,
		},
		{
			name: "valid params with checks",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
				"checks":   []string{"rpki"},
			},
			expectError: false,
		},
		{
			name: "valid params with asn",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
				"asn":      "AS15169",
			},
			expectError: false,
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"checks": []string{"rpki"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tools.ValidateSecurity(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateSecurity() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result != nil {
				if asn, ok := result.Metadata["asn"]; ok && tt.params["asn"] != nil {
					if asn != tt.params["asn"] {
						t.Errorf("ValidateSecurity() asn = %v, want %v", asn, tt.params["asn"])
					}
				}
			}
		})
	}
}

func TestTools_ExploreRelationships(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getASNNeighbours": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"resource": "AS15169",
			},
			expectError: false,
		},
		{
			name: "valid params with relationships",
			params: map[string]interface{}{
				"resource":      "AS15169",
				"relationships": []string{"neighbors"},
			},
			expectError: false,
		},
		{
			name: "valid params with scope direct",
			params: map[string]interface{}{
				"resource": "AS15169",
				"scope":    ScopeDirect,
			},
			expectError: false,
		},
		{
			name: "valid params with scope extended",
			params: map[string]interface{}{
				"resource": "AS15169",
				"scope":    ScopeExtended,
			},
			expectError: false,
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"relationships": []string{"neighbors"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tools.ExploreRelationships(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("ExploreRelationships() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && result != nil {
				if scope, ok := result.Metadata["scope"]; ok && tt.params["scope"] != nil {
					if scope != tt.params["scope"] {
						t.Errorf("ExploreRelationships() scope = %v, want %v", scope, tt.params["scope"])
					}
				}
			}
		})
	}
}

func TestTools_SearchByLocation(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getCountryASNs": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"country": "US",
			},
			expectError: false,
		},
		{
			name: "valid params with type",
			params: map[string]interface{}{
				"country": "US",
				"type":    LocationTypeASNs,
			},
			expectError: false,
		},
		{
			name: "invalid type",
			params: map[string]interface{}{
				"country": "US",
				"type":    "invalid",
			},
			expectError: true,
		},
		{
			name: "missing country",
			params: map[string]interface{}{
				"type": LocationTypeASNs,
			},
			expectError: true,
		},
		{
			name: "invalid country code",
			params: map[string]interface{}{
				"country": "INVALID",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tools.SearchByLocation(ctx, tt.params)
			if (err != nil) != tt.expectError {
				t.Errorf("SearchByLocation() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestTools_ExecuteIndividualEndpoint(t *testing.T) {
	executor := &mockExecutor{
		results: map[string]interface{}{
			"getNetworkInfo": map[string]interface{}{"status": "ok"},
		},
		errors: map[string]error{},
	}
	tools := NewTools(executor)
	ctx := context.Background()

	tests := []struct {
		name        string
		endpoint    string
		args        map[string]interface{}
		expectError bool
	}{
		{
			name:     "valid endpoint with resource",
			endpoint: "getNetworkInfo",
			args: map[string]interface{}{
				"resource": "8.8.8.8",
			},
			expectError: false,
		},
		{
			name:     "valid endpoint with additional params",
			endpoint: "getNetworkInfo",
			args: map[string]interface{}{
				"resource": "8.8.8.8",
				"extra":    "value",
			},
			expectError: false,
		},
		{
			name:     "missing resource",
			endpoint: "getNetworkInfo",
			args: map[string]interface{}{
				"extra": "value",
			},
			expectError: true,
		},
		{
			name:     "empty resource",
			endpoint: "getNetworkInfo",
			args: map[string]interface{}{
				"resource": "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tools.ExecuteIndividualEndpoint(ctx, tt.endpoint, tt.args)
			if (err != nil) != tt.expectError {
				t.Errorf("ExecuteIndividualEndpoint() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestTranslateDepthToEndpointParams(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		depth    string
		wantLOD  int
	}{
		{
			name:     "getASNNeighbours with basic depth",
			endpoint: "getASNNeighbours",
			depth:    DepthBasic,
			wantLOD:  0,
		},
		{
			name:     "getASNNeighbours with detailed depth",
			endpoint: "getASNNeighbours",
			depth:    DepthDetailed,
			wantLOD:  1,
		},
		{
			name:     "getASNNeighbours with comprehensive depth",
			endpoint: "getASNNeighbours",
			depth:    DepthComprehensive,
			wantLOD:  1,
		},
		{
			name:     "getCountryASNs with basic depth",
			endpoint: "getCountryASNs",
			depth:    DepthBasic,
			wantLOD:  0,
		},
		{
			name:     "getCountryASNs with detailed depth",
			endpoint: "getCountryASNs",
			depth:    DepthDetailed,
			wantLOD:  1,
		},
		{
			name:     "other endpoint with any depth",
			endpoint: "getNetworkInfo",
			depth:    DepthDetailed,
			wantLOD:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := translateDepthToEndpointParams(tt.endpoint, tt.depth)
			if lodEndpoints[tt.endpoint] {
				if lod, ok := params["lod"].(int); ok {
					if lod != tt.wantLOD {
						t.Errorf("translateDepthToEndpointParams() lod = %v, want %v", lod, tt.wantLOD)
					}
				} else {
					t.Error("translateDepthToEndpointParams() missing lod parameter")
				}
			}
		})
	}
}

var lodEndpoints = map[string]bool{
	"getASNNeighbours": true,
	"getCountryASNs":   true,
}
