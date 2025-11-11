# Code Cleanup Opportunities - Sprint 34

## Executive Summary

The sprint-34 consolidated tools implementation (1,197 lines) contains significant code duplication that can be reduced by **~40%** through refactoring. The main issues are:

- **5 instances** of identical type-switch parameter conversion logic
- **5 instances** of identical resource extraction pattern
- **6 instances** of identical error handling patterns
- **3 instances** of metadata initialization checks
- Multiple mapping functions with similar structure

## Priority 1: Extract Parameter Conversion Helper Functions

### Issue: Duplicated Type Switch Logic (5x)
**Lines affected**: 40-53, 90-125, 173-208, 250-273, 324-355 in `tools.go`

Every tool method has nearly identical code for converting `interface{}` parameters to string slices:

```go
switch v := operationsParam.(type) {
case []string:
    for _, op := range v {
        operations = append(operations, Operation(op))
    }
case []interface{}:
    for _, op := range v {
        if opStr, ok := op.(string); ok {
            operations = append(operations, Operation(opStr))
        }
    }
default:
    return nil, fmt.Errorf("operations must be an array of strings")
}
```

**Recommendation**: Create helper function:

```go
// convertToStringSlice converts interface{} parameter to []string.
func convertToStringSlice(param interface{}, paramName string) ([]string, error) {
    switch v := param.(type) {
    case []string:
        return v, nil
    case []interface{}:
        result := make([]string, 0, len(v))
        for _, item := range v {
            if str, ok := item.(string); ok {
                result = append(result, str)
            } else {
                return nil, fmt.Errorf("%s must contain only strings", paramName)
            }
        }
        return result, nil
    default:
        return nil, fmt.Errorf("%s must be an array of strings", paramName)
    }
}
```

**Impact**: Reduces ~200 lines of duplicated code

---

## Priority 2: Extract Resource Extraction Pattern

### Issue: Repeated Resource Validation (5x)
**Lines affected**: 28-30, 78-80, 161-163, 238-240, 312-314 in `tools.go`

Every method starts with:

```go
resource, ok := params["resource"].(string)
if !ok || resource == "" {
    return nil, fmt.Errorf("resource parameter is required")
}
```

**Recommendation**: Create helper function:

```go
// extractRequiredString extracts and validates a required string parameter.
func extractRequiredString(params map[string]interface{}, key string) (string, error) {
    value, ok := params[key].(string)
    if !ok || value == "" {
        return "", fmt.Errorf("%s parameter is required", key)
    }
    return value, nil
}

// extractOptionalString extracts an optional string parameter with default.
func extractOptionalString(params map[string]interface{}, key, defaultValue string) string {
    if value, ok := params[key].(string); ok && value != "" {
        return value
    }
    return defaultValue
}
```

**Impact**: Reduces ~50 lines, improves consistency

---

## Priority 3: Create Mapping Registry

### Issue: Duplicated Mapping Logic
**Lines affected**: 93-120, 176-203, 253-268, 327-350 in `tools.go`

Four different methods have nearly identical mapping logic with nested switches:

```go
switch v := analysisParam.(type) {
case []string:
    for _, analysis := range v {
        switch analysis {
        case "consistency":
            operations = append(operations, OpConsistency)
        // ... more cases
        }
    }
case []interface{}:
    // Duplicate of above logic
}
```

**Recommendation**: Create a mapping-based approach:

```go
// paramToOperationMap defines mappings from parameter values to operations.
var paramToOperationMap = map[string]map[string]Operation{
    "analysis": {
        "consistency":       OpConsistency,
        "path-optimization": OpRouting,
        "updates":           OpUpdates,
        "looking-glass":     OpLookingGlass,
    },
    "data": {
        "whois":              OpOverview,
        "allocation-history": OpHistory,
        "hierarchy":          OpHierarchy,
        "contacts":           OpSecurity,
    },
    "checks": {
        "rpki":           OpSecurity,
        "abuse-contacts": OpSecurity,
        "bgp-hijacking":  OpSecurity,
    },
    "relationships": {
        "neighbors":           OpNeighbors,
        "announced-prefixes":  OpRouting,
        "related-networks":    OpRelationships,
    },
}

// mapStringsToOperations converts string values to Operations using a mapping.
func mapStringsToOperations(values []string, mappingKey string) ([]Operation, error) {
    mapping, exists := paramToOperationMap[mappingKey]
    if !exists {
        return nil, fmt.Errorf("unknown mapping key: %s", mappingKey)
    }

    var operations []Operation
    for _, value := range values {
        op, exists := mapping[value]
        if !exists {
            return nil, fmt.Errorf("unsupported %s type: %s", mappingKey, value)
        }
        operations = append(operations, op)
    }
    return operations, nil
}
```

**Impact**: Reduces ~250 lines of nested switch statements

---

## Priority 4: Extract Common Execution Pattern

### Issue: Repeated Detect-Route-Execute Pattern (6x)
**Lines affected**: Throughout all tool methods

Every method follows the same pattern:

```go
// 1. Detect resource type
detected, err := DetectResource(resource)
if err != nil {
    return nil, fmt.Errorf("failed to detect resource type: %w", err)
}

// 2. Route operations to endpoints
routes, err := RouteOperations(detected, operations)
if err != nil {
    return nil, fmt.Errorf("failed to route operations: %w", err)
}

// 3. Execute and aggregate
return ct.executeAndAggregate(ctx, detected, operations, routes, depth)
```

**Recommendation**: Create wrapper function:

```go
// executeOperations is a convenience wrapper for the detect-route-execute pattern.
func (ct *Tools) executeOperations(
    ctx context.Context,
    resource string,
    operations []Operation,
    depth string,
) (*Result, error) {
    detected, err := DetectResource(resource)
    if err != nil {
        return nil, fmt.Errorf("failed to detect resource type: %w", err)
    }

    routes, err := RouteOperations(detected, operations)
    if err != nil {
        return nil, fmt.Errorf("failed to route operations: %w", err)
    }

    return ct.executeAndAggregate(ctx, detected, operations, routes, depth)
}
```

**Impact**: Reduces ~90 lines

---

## Priority 5: Simplify Metadata Handling

### Issue: Repeated Metadata Nil Check (3x)
**Lines affected**: 151-154, 301-304, 386-389 in `tools.go`

```go
if result.Metadata == nil {
    result.Metadata = make(map[string]interface{})
}
result.Metadata["key"] = value
```

**Recommendation**: Create helper method:

```go
// addMetadata safely adds metadata to a result, initializing the map if needed.
func (r *Result) addMetadata(key string, value interface{}) {
    if r.Metadata == nil {
        r.Metadata = make(map[string]interface{})
    }
    r.Metadata[key] = value
}

// addMetadataMap adds multiple metadata entries at once.
func (r *Result) addMetadataMap(data map[string]interface{}) {
    if r.Metadata == nil {
        r.Metadata = make(map[string]interface{})
    }
    for k, v := range data {
        r.Metadata[k] = v
    }
}
```

**Impact**: Cleaner, more maintainable code

---

## Priority 6: Define Constants for Magic Strings

### Issue: Magic Strings Throughout Code

Numerous hard-coded strings like:
- "consistency", "path-optimization", "updates"
- "whois", "allocation-history", "hierarchy"
- "rpki", "abuse-contacts", "bgp-hijacking"
- "basic", "detailed", "comprehensive"

**Recommendation**: Define constants:

```go
// Depth levels for investigation detail.
const (
    DepthBasic        = "basic"
    DepthDetailed     = "detailed"
    DepthComprehensive = "comprehensive"
)

// Analysis types for routing analysis.
const (
    AnalysisConsistency      = "consistency"
    AnalysisPathOptimization = "path-optimization"
    AnalysisUpdates          = "updates"
    AnalysisLookingGlass     = "looking-glass"
)

// Registry data types.
const (
    DataTypeWhois            = "whois"
    DataTypeAllocationHistory = "allocation-history"
    DataTypeHierarchy        = "hierarchy"
    DataTypeContacts         = "contacts"
)

// Security check types.
const (
    SecurityCheckRPKI          = "rpki"
    SecurityCheckAbuseContacts = "abuse-contacts"
    SecurityCheckBGPHijacking  = "bgp-hijacking"
)
```

**Impact**: Better maintainability, refactoring safety

---

## Priority 7: Add Unit Tests for Helper Functions

### Issue: No Tests for Internal Logic

Currently, the consolidated package has no dedicated unit tests for the tool methods or parameter conversion logic.

**Recommendation**: Add tests:

```go
// tools_test.go
package consolidated

func TestConvertToStringSlice(t *testing.T) {
    tests := []struct {
        name      string
        input     interface{}
        paramName string
        want      []string
        wantErr   bool
    }{
        {
            name:      "string slice",
            input:     []string{"a", "b", "c"},
            paramName: "test",
            want:      []string{"a", "b", "c"},
            wantErr:   false,
        },
        {
            name:      "interface slice with strings",
            input:     []interface{}{"a", "b", "c"},
            paramName: "test",
            want:      []string{"a", "b", "c"},
            wantErr:   false,
        },
        {
            name:      "invalid type",
            input:     "not a slice",
            paramName: "test",
            want:      nil,
            wantErr:   true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := convertToStringSlice(tt.input, tt.paramName)
            if (err != nil) != tt.wantErr {
                t.Errorf("convertToStringSlice() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("convertToStringSlice() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Impact**: Prevents regressions, documents expected behavior

---

## Priority 8: Simplify Tool Methods

### Issue: Long, Complex Methods

After extracting helpers, each tool method can be simplified significantly.

**Example - Before** (~80 lines):
```go
func (ct *Tools) AnalyzeRouting(ctx context.Context, params map[string]interface{}) (*Result, error) {
    resource, ok := params["resource"].(string)
    if !ok || resource == "" {
        return nil, fmt.Errorf("resource parameter is required")
    }

    analysisParam, ok := params["analysis"]
    if !ok {
        analysisParam = []string{"consistency"}
    }

    var operations []Operation
    switch v := analysisParam.(type) {
    case []string:
        for _, analysis := range v {
            switch analysis {
            case "consistency":
                operations = append(operations, OpConsistency)
            // ... many more cases
            }
        }
    case []interface{}:
        // ... duplicate logic
    }

    // ... more code
}
```

**Example - After** (~20 lines):
```go
func (ct *Tools) AnalyzeRouting(ctx context.Context, params map[string]interface{}) (*Result, error) {
    resource, err := extractRequiredString(params, "resource")
    if err != nil {
        return nil, err
    }

    analysisTypes, err := extractStringSliceWithDefault(
        params, "analysis", []string{AnalysisConsistency},
    )
    if err != nil {
        return nil, err
    }

    operations, err := mapStringsToOperations(analysisTypes, "analysis")
    if err != nil {
        return nil, err
    }

    timeframe := extractOptionalString(params, "timeframe", "current")

    result, err := ct.executeOperations(ctx, resource, operations, DepthDetailed)
    if err != nil {
        return nil, err
    }

    result.addMetadata("timeframe", timeframe)
    return result, nil
}
```

**Impact**: 75% reduction in method length, much easier to read

---

## Impact Summary

| Priority | Refactoring | Lines Saved | Effort |
|----------|-------------|-------------|--------|
| 1 | Parameter conversion helpers | ~200 | Medium |
| 2 | Resource extraction | ~50 | Low |
| 3 | Mapping registry | ~250 | Medium |
| 4 | Common execution pattern | ~90 | Low |
| 5 | Metadata handling | ~30 | Low |
| 6 | Constants | ~0 | Low |
| 7 | Unit tests | -100 (add) | High |
| 8 | Simplify methods | ~150 | Low |
| **Total** | **Net reduction** | **~670 lines** | **~3-4 hours** |

## Additional Observations

### Positive Aspects
1. ✅ Clear separation of concerns with detector, router, executor
2. ✅ Good use of interfaces (ToolExecutor)
3. ✅ Consistent naming conventions
4. ✅ Comprehensive error handling
5. ✅ Good documentation comments

### Minor Issues
1. `executeAndAggregate` continues on error instead of short-circuiting - consider adding option for fail-fast
2. No validation that operations are supported for detected resource type before routing
3. Dependencies in RouteResult are built but never used
4. Consider adding context cancellation checks in loops

## Recommended Implementation Order

1. **Phase 1** (Low risk): Add helper functions without changing existing code
2. **Phase 2** (Medium risk): Add unit tests for helpers
3. **Phase 3** (High risk): Refactor tool methods to use helpers
4. **Phase 4** (Low risk): Add constants and update references
5. **Phase 5** (Medium risk): Add integration tests

## Files to Create

```
internal/mcp/consolidated/
  ├── params.go          # Parameter extraction helpers
  ├── params_test.go     # Tests for parameter helpers
  ├── mapping.go         # Operation mapping logic
  ├── mapping_test.go    # Tests for mapping
  └── constants.go       # String constants
```

This refactoring will make the code:
- ✅ 40% shorter
- ✅ More maintainable
- ✅ Easier to test
- ✅ More consistent
- ✅ Less error-prone
