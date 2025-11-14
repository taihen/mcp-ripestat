package consolidated

import (
	"testing"
)

func TestDetectResource_ASN(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  ResourceType
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid ASN uppercase",
			input:     "AS15169",
			wantType:  ASN,
			wantValue: "AS15169",
			wantErr:   false,
		},
		{
			name:      "valid ASN lowercase",
			input:     "as15169",
			wantType:  ASN,
			wantValue: "AS15169",
			wantErr:   false,
		},
		{
			name:      "valid ASN with A prefix",
			input:     "A15169",
			wantType:  ASN,
			wantValue: "AS15169",
			wantErr:   false,
		},
		{
			name:      "valid ASN with spaces",
			input:     "  AS15169  ",
			wantType:  ASN,
			wantValue: "AS15169",
			wantErr:   false,
		},
		{
			name:      "invalid ASN - too large",
			input:     "AS4294967296",
			wantType:  Invalid,
			wantValue: "AS4294967296",
			wantErr:   true,
		},
		{
			name:      "invalid ASN - zero",
			input:     "AS0",
			wantType:  Invalid,
			wantValue: "AS0",
			wantErr:   true,
		},
		{
			name:      "invalid ASN - negative",
			input:     "AS-1",
			wantType:  Invalid,
			wantValue: "AS-1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if !tt.wantErr {
					t.Error("DetectResource() returned nil resource")
				}
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("DetectResource() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Value != tt.wantValue {
				t.Errorf("DetectResource() Value = %v, want %v", got.Value, tt.wantValue)
			}
		})
	}
}

func TestDetectResource_Country(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  ResourceType
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid country code uppercase",
			input:     "US",
			wantType:  Country,
			wantValue: "US",
			wantErr:   false,
		},
		{
			name:      "valid country code lowercase",
			input:     "us",
			wantType:  Country,
			wantValue: "US",
			wantErr:   false,
		},
		{
			name:      "valid country code with spaces",
			input:     "  NL  ",
			wantType:  Country,
			wantValue: "NL",
			wantErr:   false,
		},
		{
			name:      "invalid country code - too short",
			input:     "U",
			wantType:  Invalid,
			wantValue: "U",
			wantErr:   true,
		},
		{
			name:      "invalid country code - too long",
			input:     "USA",
			wantType:  Invalid,
			wantValue: "USA",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if !tt.wantErr {
					t.Error("DetectResource() returned nil resource")
				}
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("DetectResource() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Value != tt.wantValue {
				t.Errorf("DetectResource() Value = %v, want %v", got.Value, tt.wantValue)
			}
		})
	}
}

func TestDetectResource_IPPrefix(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  ResourceType
		wantValue string
		wantVer   int
		wantErr   bool
	}{
		{
			name:      "valid IPv4 CIDR",
			input:     "8.8.8.0/24",
			wantType:  IPPrefix,
			wantValue: "8.8.8.0/24",
			wantVer:   4,
			wantErr:   false,
		},
		{
			name:      "valid IPv6 CIDR",
			input:     "2001:db8::/32",
			wantType:  IPPrefix,
			wantValue: "2001:db8::/32",
			wantVer:   6,
			wantErr:   false,
		},
		{
			name:      "valid IPv6 CIDR with spaces",
			input:     "  2001:db8::/32  ",
			wantType:  IPPrefix,
			wantValue: "2001:db8::/32",
			wantVer:   6,
			wantErr:   false,
		},
		{
			name:      "invalid CIDR - malformed",
			input:     "8.8.8.0/",
			wantType:  Invalid,
			wantValue: "8.8.8.0/",
			wantErr:   true,
		},
		{
			name:      "invalid CIDR - invalid IP",
			input:     "999.999.999.999/24",
			wantType:  Invalid,
			wantValue: "999.999.999.999/24",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if !tt.wantErr {
					t.Error("DetectResource() returned nil resource")
				}
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("DetectResource() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Value != tt.wantValue {
				t.Errorf("DetectResource() Value = %v, want %v", got.Value, tt.wantValue)
			}
			if tt.wantVer > 0 && got.Version != tt.wantVer {
				t.Errorf("DetectResource() Version = %v, want %v", got.Version, tt.wantVer)
			}
		})
	}
}

func TestDetectResource_IPAddress(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  ResourceType
		wantValue string
		wantVer   int
		wantErr   bool
	}{
		{
			name:      "valid IPv4 address",
			input:     "8.8.8.8",
			wantType:  IPAddress,
			wantValue: "8.8.8.8",
			wantVer:   4,
			wantErr:   false,
		},
		{
			name:      "valid IPv6 address",
			input:     "2001:db8::1",
			wantType:  IPAddress,
			wantValue: "2001:db8::1",
			wantVer:   6,
			wantErr:   false,
		},
		{
			name:      "valid IPv4 address with spaces",
			input:     "  8.8.8.8  ",
			wantType:  IPAddress,
			wantValue: "8.8.8.8",
			wantVer:   4,
			wantErr:   false,
		},
		{
			name:      "invalid IP address",
			input:     "999.999.999.999",
			wantType:  Invalid,
			wantValue: "999.999.999.999",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if !tt.wantErr {
					t.Error("DetectResource() returned nil resource")
				}
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("DetectResource() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Value != tt.wantValue {
				t.Errorf("DetectResource() Value = %v, want %v", got.Value, tt.wantValue)
			}
			if tt.wantVer > 0 && got.Version != tt.wantVer {
				t.Errorf("DetectResource() Version = %v, want %v", got.Version, tt.wantVer)
			}
		})
	}
}

func TestDetectResource_EmptyInput(t *testing.T) {
	got, err := DetectResource("")
	if err == nil {
		t.Error("DetectResource() expected error for empty input")
	}
	if got != nil {
		t.Error("DetectResource() expected nil resource for empty input")
	}
}

func TestDetectResource_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "random string",
			input: "random-string",
		},
		{
			name:  "number without AS prefix",
			input: "12345",
		},
		{
			name:  "mixed alphanumeric",
			input: "AS123abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectResource(tt.input)
			if err == nil {
				t.Error("DetectResource() expected error for invalid input")
			}
			if got == nil {
				t.Error("DetectResource() expected non-nil resource even for invalid input")
				return
			}
			if got.Type != Invalid {
				t.Errorf("DetectResource() Type = %v, want Invalid", got.Type)
			}
		})
	}
}

func TestGetIPVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVer int
		wantErr bool
	}{
		{
			name:    "valid IPv4 address",
			input:   "8.8.8.8",
			wantVer: 4,
			wantErr: false,
		},
		{
			name:    "valid IPv6 address",
			input:   "2001:db8::1",
			wantVer: 6,
			wantErr: false,
		},
		{
			name:    "valid IPv4 CIDR",
			input:   "8.8.8.0/24",
			wantVer: 4,
			wantErr: false,
		},
		{
			name:    "valid IPv6 CIDR",
			input:   "2001:db8::/32",
			wantVer: 6,
			wantErr: false,
		},
		{
			name:    "invalid input - not IP or prefix",
			input:   "AS15169",
			wantVer: 0,
			wantErr: true,
		},
		{
			name:    "invalid input - random string",
			input:   "random-string",
			wantVer: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVer, err := GetIPVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVer != tt.wantVer {
				t.Errorf("GetIPVersion() = %v, want %v", gotVer, tt.wantVer)
			}
		})
	}
}

