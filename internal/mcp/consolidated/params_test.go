package consolidated

import (
	"reflect"
	"testing"
)

func TestExtractRequiredString(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		key       string
		want      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid string",
			params:  map[string]interface{}{"resource": "AS3333"},
			key:     "resource",
			want:    "AS3333",
			wantErr: false,
		},
		{
			name:      "missing key",
			params:    map[string]interface{}{},
			key:       "resource",
			want:      "",
			wantErr:   true,
			errSubstr: "required",
		},
		{
			name:      "empty string",
			params:    map[string]interface{}{"resource": ""},
			key:       "resource",
			want:      "",
			wantErr:   true,
			errSubstr: "required",
		},
		{
			name:      "wrong type",
			params:    map[string]interface{}{"resource": 123},
			key:       "resource",
			want:      "",
			wantErr:   true,
			errSubstr: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractRequiredString(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractRequiredString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractRequiredString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractOptionalString(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "value present",
			params:       map[string]interface{}{"depth": "detailed"},
			key:          "depth",
			defaultValue: "basic",
			want:         "detailed",
		},
		{
			name:         "value missing",
			params:       map[string]interface{}{},
			key:          "depth",
			defaultValue: "basic",
			want:         "basic",
		},
		{
			name:         "empty string uses default",
			params:       map[string]interface{}{"depth": ""},
			key:          "depth",
			defaultValue: "basic",
			want:         "basic",
		},
		{
			name:         "wrong type uses default",
			params:       map[string]interface{}{"depth": 123},
			key:          "depth",
			defaultValue: "basic",
			want:         "basic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractOptionalString(tt.params, tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("extractOptionalString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		param     interface{}
		paramName string
		want      []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "nil input",
			param:     nil,
			paramName: "test",
			want:      nil,
			wantErr:   false,
		},
		{
			name:      "string slice",
			param:     []string{"a", "b", "c"},
			paramName: "test",
			want:      []string{"a", "b", "c"},
			wantErr:   false,
		},
		{
			name:      "interface slice with strings",
			param:     []interface{}{"a", "b", "c"},
			paramName: "test",
			want:      []string{"a", "b", "c"},
			wantErr:   false,
		},
		{
			name:      "empty interface slice",
			param:     []interface{}{},
			paramName: "test",
			want:      []string{},
			wantErr:   false,
		},
		{
			name:      "interface slice with mixed types",
			param:     []interface{}{"a", 123, "c"},
			paramName: "test",
			want:      nil,
			wantErr:   true,
			errSubstr: "contain only strings",
		},
		{
			name:      "invalid type - string",
			param:     "not a slice",
			paramName: "test",
			want:      nil,
			wantErr:   true,
			errSubstr: "array of strings",
		},
		{
			name:      "invalid type - int",
			param:     123,
			paramName: "test",
			want:      nil,
			wantErr:   true,
			errSubstr: "array of strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToStringSlice(tt.param, tt.paramName)
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

func TestExtractStringSliceWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue []string
		want         []string
		wantErr      bool
	}{
		{
			name:         "key missing uses default",
			params:       map[string]interface{}{},
			key:          "operations",
			defaultValue: []string{"overview"},
			want:         []string{"overview"},
			wantErr:      false,
		},
		{
			name:         "string slice present",
			params:       map[string]interface{}{"operations": []string{"routing", "security"}},
			key:          "operations",
			defaultValue: []string{"overview"},
			want:         []string{"routing", "security"},
			wantErr:      false,
		},
		{
			name:         "interface slice present",
			params:       map[string]interface{}{"operations": []interface{}{"routing", "security"}},
			key:          "operations",
			defaultValue: []string{"overview"},
			want:         []string{"routing", "security"},
			wantErr:      false,
		},
		{
			name:         "empty slice uses default",
			params:       map[string]interface{}{"operations": []string{}},
			key:          "operations",
			defaultValue: []string{"overview"},
			want:         []string{"overview"},
			wantErr:      false,
		},
		{
			name:         "invalid type returns error",
			params:       map[string]interface{}{"operations": "not a slice"},
			key:          "operations",
			defaultValue: []string{"overview"},
			want:         nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractStringSliceWithDefault(tt.params, tt.key, tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractStringSliceWithDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractStringSliceWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToOperations(t *testing.T) {
	tests := []struct {
		name    string
		strings []string
		want    []Operation
	}{
		{
			name:    "empty slice",
			strings: []string{},
			want:    []Operation{},
		},
		{
			name:    "single operation",
			strings: []string{"overview"},
			want:    []Operation{OpOverview},
		},
		{
			name:    "multiple operations",
			strings: []string{"overview", "routing", "security"},
			want:    []Operation{"overview", "routing", "security"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toOperations(tt.strings); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}
