package consolidated

import (
	"reflect"
	"testing"
)

func TestResult_AddMetadata(t *testing.T) {
	tests := []struct {
		name         string
		initialMeta  map[string]interface{}
		key          string
		value        interface{}
		wantMetadata map[string]interface{}
	}{
		{
			name:        "add to nil metadata",
			initialMeta: nil,
			key:         "test",
			value:       "value",
			wantMetadata: map[string]interface{}{
				"test": "value",
			},
		},
		{
			name: "add to existing metadata",
			initialMeta: map[string]interface{}{
				"existing": "data",
			},
			key:   "new",
			value: "value",
			wantMetadata: map[string]interface{}{
				"existing": "data",
				"new":      "value",
			},
		},
		{
			name: "overwrite existing key",
			initialMeta: map[string]interface{}{
				"key": "old",
			},
			key:   "key",
			value: "new",
			wantMetadata: map[string]interface{}{
				"key": "new",
			},
		},
		{
			name:        "add complex value",
			initialMeta: nil,
			key:         "data",
			value:       map[string]string{"nested": "value"},
			wantMetadata: map[string]interface{}{
				"data": map[string]string{"nested": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				Metadata: tt.initialMeta,
			}
			r.AddMetadata(tt.key, tt.value)

			if !reflect.DeepEqual(r.Metadata, tt.wantMetadata) {
				t.Errorf("AddMetadata() metadata = %v, want %v", r.Metadata, tt.wantMetadata)
			}
		})
	}
}

func TestResult_AddMetadataMap(t *testing.T) {
	tests := []struct {
		name         string
		initialMeta  map[string]interface{}
		data         map[string]interface{}
		wantMetadata map[string]interface{}
	}{
		{
			name:        "add to nil metadata",
			initialMeta: nil,
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantMetadata: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "add to existing metadata",
			initialMeta: map[string]interface{}{
				"existing": "data",
			},
			data: map[string]interface{}{
				"new1": "value1",
				"new2": "value2",
			},
			wantMetadata: map[string]interface{}{
				"existing": "data",
				"new1":     "value1",
				"new2":     "value2",
			},
		},
		{
			name: "merge with overlapping keys",
			initialMeta: map[string]interface{}{
				"key1": "old1",
				"key2": "old2",
			},
			data: map[string]interface{}{
				"key2": "new2",
				"key3": "new3",
			},
			wantMetadata: map[string]interface{}{
				"key1": "old1",
				"key2": "new2",
				"key3": "new3",
			},
		},
		{
			name:         "add empty map",
			initialMeta:  nil,
			data:         map[string]interface{}{},
			wantMetadata: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				Metadata: tt.initialMeta,
			}
			r.AddMetadataMap(tt.data)

			if !reflect.DeepEqual(r.Metadata, tt.wantMetadata) {
				t.Errorf("AddMetadataMap() metadata = %v, want %v", r.Metadata, tt.wantMetadata)
			}
		})
	}
}

func TestResourceType_String(t *testing.T) {
	tests := []struct {
		name string
		rt   ResourceType
		want string
	}{
		{
			name: "IPAddress",
			rt:   IPAddress,
			want: "ip_address",
		},
		{
			name: "IPPrefix",
			rt:   IPPrefix,
			want: "ip_prefix",
		},
		{
			name: "ASN",
			rt:   ASN,
			want: "asn",
		},
		{
			name: "Country",
			rt:   Country,
			want: "country",
		},
		{
			name: "Invalid",
			rt:   Invalid,
			want: "invalid",
		},
		{
			name: "Unknown value",
			rt:   ResourceType(99),
			want: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rt.String(); got != tt.want {
				t.Errorf("ResourceType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
