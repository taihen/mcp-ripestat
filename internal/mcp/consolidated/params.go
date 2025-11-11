package consolidated

import "fmt"

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

// convertToStringSlice converts interface{} parameter to []string.
func convertToStringSlice(param interface{}, paramName string) ([]string, error) {
	if param == nil {
		return nil, nil
	}

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

// extractStringSliceWithDefault extracts a string slice parameter with a default value.
func extractStringSliceWithDefault(params map[string]interface{}, key string, defaultValue []string) ([]string, error) {
	param, ok := params[key]
	if !ok {
		return defaultValue, nil
	}

	result, err := convertToStringSlice(param, key)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return defaultValue, nil
	}

	return result, nil
}

// toOperations converts a string slice to Operation slice.
func toOperations(strings []string) []Operation {
	operations := make([]Operation, len(strings))
	for i, s := range strings {
		operations[i] = Operation(s)
	}
	return operations
}
