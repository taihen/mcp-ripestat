package consolidated

import (
	"context"
	"fmt"
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

func TestExtractNetworkInfoFromTypedResponse(t *testing.T) {
	type data struct {
		Prefix string
		ASNs   []interface{}
	}
	type response struct {
		Data data
	}

	fixture := &response{Data: data{Prefix: "8.8.8.0/24", ASNs: []interface{}{15169}}}
	if got := extractPrefixFromNetworkInfo(fixture); got != "8.8.8.0/24" {
		t.Errorf("extractPrefixFromNetworkInfo() = %q", got)
	}
	if got := extractASNsFromNetworkInfo(fixture); !reflect.DeepEqual(got, []string{"AS15169"}) {
		t.Errorf("extractASNsFromNetworkInfo() = %v", got)
	}
}

func TestPrepareEndpointCall(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		dependencies map[string][]string
		results      map[string]interface{}
		resource     *DetectedResource
		params       map[string]interface{}
		wantParams   map[string]interface{}
		wantResource string
		wantErr      bool
	}{
		{
			name:     "RPKI derives prefix and single origin ASN",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
						"asns":   []interface{}{15169},
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			params: map[string]interface{}{},
			wantParams: map[string]interface{}{
				"prefix": "8.8.8.0/24",
			},
			wantResource: "AS15169",
		},
		{
			name:     "RPKI uses explicit ASN",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
						"asns":   []interface{}{15169, 3356},
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			params:       map[string]interface{}{"asn": "AS3356"},
			wantParams:   map[string]interface{}{"prefix": "8.8.8.0/24"},
			wantResource: "AS3356",
		},
		{
			name:     "RPKI rejects ambiguous origin",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
						"asns":   []interface{}{15169, 3356},
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPPrefix,
				Value: "8.8.8.0/24",
			},
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name:     "RPKI preserves requested more-specific prefix",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
						"asns":   []interface{}{15169},
					},
				},
			},
			resource: &DetectedResource{
				Type:  IPPrefix,
				Value: "8.8.8.0/25",
			},
			params:       map[string]interface{}{"asn": "AS15169"},
			wantParams:   map[string]interface{}{"prefix": "8.8.8.0/25"},
			wantResource: "AS15169",
		},
		{
			name:     "hierarchy converts IP to prefix",
			endpoint: "getAddressSpaceHierarchy",
			dependencies: map[string][]string{
				"getAddressSpaceHierarchy": {"getNetworkInfo"},
			},
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{"prefix": "8.8.8.0/24"},
				},
			},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			params:       map[string]interface{}{},
			wantParams:   map[string]interface{}{},
			wantResource: "8.8.8.0/24",
		},
		{
			name:     "failed dependency blocks call",
			endpoint: "getRPKIValidation",
			dependencies: map[string][]string{
				"getRPKIValidation": {"getNetworkInfo"},
			},
			results: map[string]interface{}{},
			resource: &DetectedResource{
				Type:  IPAddress,
				Value: "8.8.8.8",
			},
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResource, err := prepareEndpointCall(tt.endpoint, tt.dependencies, tt.results, tt.params, tt.resource)
			if (err != nil) != tt.wantErr {
				t.Fatalf("prepareEndpointCall() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotResource != tt.wantResource {
				t.Errorf("prepareEndpointCall() resource = %v, want %v", gotResource, tt.wantResource)
			}
			if !reflect.DeepEqual(tt.params, tt.wantParams) {
				t.Errorf("prepareEndpointCall() params = %v, want %v", tt.params, tt.wantParams)
			}
		})
	}
}

type mockExecutor struct {
	results map[string]interface{}
	errors  map[string]error
	calls   []executorCall
}

type executorCall struct {
	endpoint string
	resource string
	params   map[string]interface{}
}

func (m *mockExecutor) ExecuteEndpoint(_ context.Context, endpoint, resource string, params map[string]interface{}) (interface{}, error) {
	m.calls = append(m.calls, executorCall{endpoint: endpoint, resource: resource, params: copyParams(params)})
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
					"asns":   []interface{}{15169},
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
	result, err := tools.executeAndAggregate(ctx, resource, []Operation{operation}, routes, DepthBasic, nil)
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
		return
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
				"analysis":  []string{AnalysisUpdates},
				"timeframe": TimeframeCurrent,
			},
			expectError: false,
		},
		{
			name: "historical timeframe without temporal analysis",
			params: map[string]interface{}{
				"resource":  "8.8.8.0/24",
				"analysis":  []string{AnalysisConsistency},
				"timeframe": Timeframe1Day,
			},
			expectError: true,
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
			name: "unimplemented detailed format",
			params: map[string]interface{}{
				"resource": "8.8.8.8",
				"format":   FormatDetailed,
			},
			expectError: true,
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
		{
			name: "asn without RPKI check",
			params: map[string]interface{}{
				"resource": "8.8.8.0/24",
				"checks":   []string{SecurityCheckAbuseContacts},
				"asn":      "AS15169",
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

func TestValidateSecurity_ExecutionContract(t *testing.T) {
	newExecutor := func() *mockExecutor {
		return &mockExecutor{
			results: map[string]interface{}{
				"getNetworkInfo": map[string]interface{}{
					"data": map[string]interface{}{
						"prefix": "8.8.8.0/24",
						"asns":   []interface{}{15169, 3356},
					},
				},
				"getRPKIValidation":           map[string]interface{}{"status": "valid"},
				"getRPKIHistory":              map[string]interface{}{"status": "ok"},
				"getAbuseContactFinder":       map[string]interface{}{"contacts": []string{"abuse@example.net"}},
				"getRoutingStatus":            map[string]interface{}{"status": "ok"},
				"getPrefixRoutingConsistency": map[string]interface{}{"status": "ok"},
				"getBGPUpdates":               map[string]interface{}{"status": "ok"},
			},
			errors: map[string]error{},
		}
	}

	t.Run("RPKI forwards ASN and derived prefix without unrelated checks", func(t *testing.T) {
		executor := newExecutor()
		tools := NewTools(executor)
		result, err := tools.ValidateSecurity(context.Background(), map[string]interface{}{
			"resource": "8.8.8.0/24",
			"checks":   []string{SecurityCheckRPKI},
			"asn":      "AS15169",
		})
		if err != nil {
			t.Fatalf("ValidateSecurity() error = %v", err)
		}
		if len(result.Errors) != 0 {
			t.Fatalf("ValidateSecurity() endpoint errors = %v", result.Errors)
		}

		assertCalledEndpoints(t, executor.calls, "getNetworkInfo", "getRPKIValidation", "getRPKIHistory")
		call := requireCall(t, executor.calls, "getRPKIValidation")
		if call.resource != "AS15169" {
			t.Errorf("RPKI resource = %q, want AS15169", call.resource)
		}
		if got := call.params["prefix"]; got != "8.8.8.0/24" {
			t.Errorf("RPKI prefix = %v, want 8.8.8.0/24", got)
		}
		if _, ok := call.params["asn"]; ok {
			t.Error("RPKI ASN must be carried as the endpoint resource, not duplicated in params")
		}
	})

	t.Run("abuse contacts does not trigger RPKI or hijacking endpoints", func(t *testing.T) {
		executor := newExecutor()
		tools := NewTools(executor)
		_, err := tools.ValidateSecurity(context.Background(), map[string]interface{}{
			"resource": "8.8.8.0/24",
			"checks":   []string{SecurityCheckAbuseContacts},
		})
		if err != nil {
			t.Fatalf("ValidateSecurity() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getAbuseContactFinder")
	})

	t.Run("BGP hijacking uses only routing anomaly signals", func(t *testing.T) {
		executor := newExecutor()
		tools := NewTools(executor)
		_, err := tools.ValidateSecurity(context.Background(), map[string]interface{}{
			"resource": "8.8.8.0/24",
			"checks":   []string{SecurityCheckBGPHijacking},
		})
		if err != nil {
			t.Fatalf("ValidateSecurity() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getRoutingStatus", "getPrefixRoutingConsistency", "getBGPUpdates")
	})

	t.Run("ASN RPKI is rejected instead of returning history", func(t *testing.T) {
		executor := newExecutor()
		tools := NewTools(executor)
		_, err := tools.ValidateSecurity(context.Background(), map[string]interface{}{
			"resource": "AS15169",
			"checks":   []string{SecurityCheckRPKI},
		})
		if err == nil {
			t.Fatal("ValidateSecurity() expected ASN RPKI to be rejected")
		}
		if len(executor.calls) != 0 {
			t.Errorf("ValidateSecurity() made calls before rejecting ASN RPKI: %v", executor.calls)
		}
	})

	t.Run("ASN defaults avoid unsupported RPKI", func(t *testing.T) {
		executor := newExecutor()
		executor.results["getASRoutingConsistency"] = map[string]interface{}{"status": "ok"}
		tools := NewTools(executor)
		_, err := tools.ValidateSecurity(context.Background(), map[string]interface{}{
			"resource": "AS15169",
		})
		if err != nil {
			t.Fatalf("ValidateSecurity() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getAbuseContactFinder", "getASRoutingConsistency", "getBGPUpdates")
	})
}

func TestSpecializedTools_ExecutionContracts(t *testing.T) {
	t.Run("registry whois invokes only whois", func(t *testing.T) {
		executor := &mockExecutor{results: map[string]interface{}{"getWhois": map[string]interface{}{"status": "ok"}}, errors: map[string]error{}}
		tools := NewTools(executor)
		_, err := tools.QueryRegistry(context.Background(), map[string]interface{}{
			"resource": "8.8.8.8",
			"data":     []string{DataTypeWhois},
		})
		if err != nil {
			t.Fatalf("QueryRegistry() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getWhois")
	})

	t.Run("routing updates forwards timeframe", func(t *testing.T) {
		executor := &mockExecutor{results: map[string]interface{}{"getBGPUpdates": map[string]interface{}{"status": "ok"}}, errors: map[string]error{}}
		tools := NewTools(executor)
		_, err := tools.AnalyzeRouting(context.Background(), map[string]interface{}{
			"resource":  "8.8.8.0/24",
			"analysis":  []string{AnalysisUpdates},
			"timeframe": TimeframeCurrent,
		})
		if err != nil {
			t.Fatalf("AnalyzeRouting() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getBGPUpdates")
		call := requireCall(t, executor.calls, "getBGPUpdates")
		if got := call.params["timeframe"]; got != TimeframeCurrent {
			t.Errorf("timeframe = %v, want %s", got, TimeframeCurrent)
		}
	})

	t.Run("announced prefixes does not expand to all routing analysis", func(t *testing.T) {
		executor := &mockExecutor{results: map[string]interface{}{"getAnnouncedPrefixes": map[string]interface{}{"status": "ok"}}, errors: map[string]error{}}
		tools := NewTools(executor)
		_, err := tools.ExploreRelationships(context.Background(), map[string]interface{}{
			"resource":      "AS15169",
			"relationships": []string{RelationshipAnnouncedPrefixes},
		})
		if err != nil {
			t.Fatalf("ExploreRelationships() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getAnnouncedPrefixes")
	})

	t.Run("prefix relationship default uses related networks", func(t *testing.T) {
		executor := &mockExecutor{results: map[string]interface{}{"getRelatedPrefixes": map[string]interface{}{"status": "ok"}}, errors: map[string]error{}}
		tools := NewTools(executor)
		_, err := tools.ExploreRelationships(context.Background(), map[string]interface{}{
			"resource": "8.8.8.0/24",
		})
		if err != nil {
			t.Fatalf("ExploreRelationships() error = %v", err)
		}
		assertCalledEndpoints(t, executor.calls, "getRelatedPrefixes")
	})
}

func TestExecuteAndAggregate_CompletenessMetadata(t *testing.T) {
	resource := &DetectedResource{Type: IPPrefix, Value: "8.8.8.0/24", Validated: true}
	routes := &RouteResult{
		Endpoints:    []string{"getWhois", "getRoutingStatus"},
		Dependencies: map[string][]string{},
	}

	t.Run("partial result is explicitly incomplete", func(t *testing.T) {
		executor := &mockExecutor{
			results: map[string]interface{}{"getWhois": map[string]interface{}{"status": "ok"}},
			errors:  map[string]error{"getRoutingStatus": fmt.Errorf("upstream failed")},
		}
		result, err := NewTools(executor).executeAndAggregate(
			context.Background(), resource, []Operation{OpOverview}, routes, DepthBasic, nil,
		)
		if err != nil {
			t.Fatalf("executeAndAggregate() error = %v", err)
		}
		if complete, ok := result.Metadata["complete"].(bool); !ok || complete {
			t.Errorf("complete = %v, want false", result.Metadata["complete"])
		}
		if len(result.Errors) != 1 {
			t.Errorf("errors = %v, want one endpoint error", result.Errors)
		}
		if got := result.Metadata["endpoints_succeeded"].([]string); !reflect.DeepEqual(got, []string{"getWhois"}) {
			t.Errorf("endpoints_succeeded = %v", got)
		}
	})

	t.Run("total failure returns top-level error", func(t *testing.T) {
		executor := &mockExecutor{
			results: map[string]interface{}{},
			errors: map[string]error{
				"getWhois":         fmt.Errorf("upstream failed"),
				"getRoutingStatus": fmt.Errorf("upstream failed"),
			},
		}
		result, err := NewTools(executor).executeAndAggregate(
			context.Background(), resource, []Operation{OpOverview}, routes, DepthBasic, nil,
		)
		if err == nil {
			t.Fatal("executeAndAggregate() expected total failure error")
		}
		if result == nil || result.Metadata["complete"] != false {
			t.Errorf("result metadata = %v, want complete=false", result)
		}
	})
}

func assertCalledEndpoints(t *testing.T, calls []executorCall, expected ...string) {
	t.Helper()
	got := make(map[string]int, len(calls))
	for _, call := range calls {
		got[call.endpoint]++
	}
	if len(got) != len(expected) {
		t.Fatalf("called endpoints = %v, want exactly %v", got, expected)
	}
	for _, endpoint := range expected {
		if got[endpoint] != 1 {
			t.Errorf("endpoint %s called %d times, want once", endpoint, got[endpoint])
		}
	}
}

func requireCall(t *testing.T, calls []executorCall, endpoint string) executorCall {
	t.Helper()
	for _, call := range calls {
		if call.endpoint == endpoint {
			return call
		}
	}
	t.Fatalf("endpoint %s was not called", endpoint)
	return executorCall{}
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
			name: "unimplemented extended scope",
			params: map[string]interface{}{
				"resource": "AS15169",
				"scope":    ScopeExtended,
			},
			expectError: true,
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
			name: "unimplemented prefixes type",
			params: map[string]interface{}{
				"country": "US",
				"type":    LocationTypePrefixes,
			},
			expectError: true,
		},
		{
			name: "unimplemented statistics type",
			params: map[string]interface{}{
				"country": "US",
				"type":    LocationTypeStatistics,
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
