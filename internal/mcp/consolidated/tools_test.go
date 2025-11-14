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
		verifyDeps   bool // If true, verify dependency order instead of exact match
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
			want:       nil, // Order can vary, we'll verify dependencies
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
						continue // Dependency not in endpoints, skip
					}

					for _, dep := range deps {
						depDepPos, depDepExists := position[dep]
						if !depDepExists {
							continue // Dependency not in endpoints, skip
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
