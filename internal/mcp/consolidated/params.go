package consolidated

import "fmt"

func extractRequiredString(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("%s parameter is required", key)
	}
	return value, nil
}

func extractOptionalString(params map[string]interface{}, key, defaultValue string) string {
	if value, ok := params[key].(string); ok && value != "" {
		return value
	}
	return defaultValue
}

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

func toOperations(strings []string) []Operation {
	operations := make([]Operation, len(strings))
	for i, s := range strings {
		operations[i] = Operation(s)
	}
	return operations
}

func validateEnum(value, paramName string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("unsupported %s value: %s", paramName, value)
}
