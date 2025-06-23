package util

import (
	"strings"
	"testing"
	"time"
)

func TestIsValidIPv4(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"256.0.0.1", false},
		{"192.168.1", false},
		{"192.168.1.1.1", false},
		{"192.168.1.a", false},
		{"2001:db8::1", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			result := IsValidIPv4(tc.ip)
			if result != tc.expected {
				t.Errorf("IsValidIPv4(%q) = %v, expected %v", tc.ip, result, tc.expected)
			}
		})
	}
}

func TestIsValidIPv6(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"2001:db8::1", true},
		{"::1", true},
		{"fe80::1", true},
		{"2001:db8:0:0:0:0:0:1", true},
		{"2001:db8::", true},
		{"192.168.1.1", false},
		{"2001:db8:g::", false},
		{"2001:db8:::1", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			result := IsValidIPv6(tc.ip)
			if result != tc.expected {
				t.Errorf("IsValidIPv6(%q) = %v, expected %v", tc.ip, result, tc.expected)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"2001:db8::1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"invalid", false},
		{"256.0.0.1", false},
		{"2001:db8:g::", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			result := IsValidIP(tc.ip)
			if result != tc.expected {
				t.Errorf("IsValidIP(%q) = %v, expected %v", tc.ip, result, tc.expected)
			}
		})
	}
}

func TestIsValidCIDR(t *testing.T) {
	testCases := []struct {
		cidr     string
		expected bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/8", true},
		{"2001:db8::/32", true},
		{"192.168.1.1", false},
		{"192.168.1.0/33", false},
		{"2001:db8::/129", false},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.cidr, func(t *testing.T) {
			result := IsValidCIDR(tc.cidr)
			if result != tc.expected {
				t.Errorf("IsValidCIDR(%q) = %v, expected %v", tc.cidr, result, tc.expected)
			}
		})
	}
}

func TestIsValidASN(t *testing.T) {
	testCases := []struct {
		asn      string
		expected bool
	}{
		{"123", true},
		{"AS123", true},
		{"as123", true},
		{"AS65536", true},
		{"0", true},
		{"AS0", true},
		{"ASN123", false},
		{"AS-123", false},
		{"AS123a", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.asn, func(t *testing.T) {
			result := IsValidASN(tc.asn)
			if result != tc.expected {
				t.Errorf("IsValidASN(%q) = %v, expected %v", tc.asn, result, tc.expected)
			}
		})
	}
}

func TestFormatASN(t *testing.T) {
	testCases := []struct {
		asn      string
		expected string
	}{
		{"123", "AS123"},
		{"AS123", "AS123"},
		{"as123", "AS123"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.asn, func(t *testing.T) {
			result := FormatASN(tc.asn)
			if result != tc.expected {
				t.Errorf("FormatASN(%q) = %q, expected %q", tc.asn, result, tc.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Create a fixed time for testing
	testTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)

	testCases := []struct {
		layout   string
		expected string
	}{
		{"", "2023-01-02T15:04:05Z"},                           // Default RFC3339
		{time.RFC3339, "2023-01-02T15:04:05Z"},                 // Explicit RFC3339
		{time.RFC822, "02 Jan 23 15:04 UTC"},                   // RFC822
		{"2006-01-02", "2023-01-02"},                           // Custom format
		{"15:04:05", "15:04:05"},                               // Time only
		{"Monday, January 2, 2006", "Monday, January 2, 2023"}, // Full date
	}

	for _, tc := range testCases {
		t.Run(tc.layout, func(t *testing.T) {
			result := FormatTime(testTime, tc.layout)
			if result != tc.expected {
				t.Errorf("FormatTime(testTime, %q) = %q, expected %q", tc.layout, result, tc.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	// Create a fixed time for testing
	expectedTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)

	testCases := []struct {
		timeStr  string
		layout   string
		expected time.Time
		hasError bool
	}{
		{"2023-01-02T15:04:05Z", "", expectedTime, false},
		{"2023-01-02T15:04:05Z", time.RFC3339, expectedTime, false},
		{"02 Jan 23 15:04 UTC", time.RFC822, time.Date(2023, 1, 2, 15, 4, 0, 0, time.UTC), false},
		{"2023-01-02", "2006-01-02", time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"invalid", "", time.Time{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.timeStr, func(t *testing.T) {
			result, err := ParseTime(tc.timeStr, tc.layout)
			if tc.hasError && err == nil {
				t.Errorf("ParseTime(%q, %q) expected error, got nil", tc.timeStr, tc.layout)
			}

			if !tc.hasError && err != nil {
				t.Errorf("ParseTime(%q, %q) unexpected error: %v", tc.timeStr, tc.layout, err)
			}

			if !tc.hasError && !result.Equal(tc.expected) {
				t.Errorf("ParseTime(%q, %q) = %v, expected %v", tc.timeStr, tc.layout, result, tc.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		str      string
		maxLen   int
		expected string
	}{
		{"Hello, World!", 13, "Hello, World!"},
		{"Hello, World!", 10, "Hello, ..."},
		{"Hello", 10, "Hello"},
		{"", 10, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			result := TruncateString(tc.str, tc.maxLen)
			if result != tc.expected {
				t.Errorf("TruncateString(%q, %d) = %q, expected %q", tc.str, tc.maxLen, result, tc.expected)
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	testCases := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"a", "", "c"}, ",", "a,c"},
		{[]string{}, ",", ""},
		{nil, ",", ""},
		{[]string{"a"}, ",", "a"},
		{[]string{"a", "b", "c"}, " - ", "a - b - c"},
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := JoinStrings(tc.strs, tc.sep)
			if result != tc.expected {
				t.Errorf("JoinStrings(%v, %q) = %q, expected %q", tc.strs, tc.sep, result, tc.expected)
			}
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	testCases := []struct {
		str      string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"a, b, c", ",", []string{"a", "b", "c"}},
		{" a , b , c ", ",", []string{"a", "b", "c"}},
		{"a,,c", ",", []string{"a", "c"}},
		{"", ",", nil},
		{"a", ",", []string{"a"}},
		{"a b c", " ", []string{"a", "b", "c"}},
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := SplitAndTrim(tc.str, tc.sep)
			if len(result) != len(tc.expected) {
				t.Errorf("SplitAndTrim(%q, %q) = %v, expected %v", tc.str, tc.sep, result, tc.expected)
				return
			}

			for j := range result {
				if result[j] != tc.expected[j] {
					t.Errorf("SplitAndTrim(%q, %q)[%d] = %q, expected %q", tc.str, tc.sep, j, result[j], tc.expected[j])
				}
			}
		})
	}
}

func TestMapToString(t *testing.T) {
	testCases := []struct {
		m        map[string]interface{}
		contains []string // We check contains instead of exact match due to map iteration order
	}{
		{map[string]interface{}{"a": 1, "b": "two"}, []string{"a: 1", "b: two"}},
		{map[string]interface{}{}, []string{}},
		{nil, []string{}},
		{map[string]interface{}{"x": true}, []string{"x: true"}},
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := MapToString(tc.m)
			if len(tc.m) == 0 && result != "{}" {
				t.Errorf("MapToString(%v) = %q, expected \"{}\"", tc.m, result)
				return
			}

			for _, s := range tc.contains {
				if !strings.Contains(result, s) {
					t.Errorf("MapToString(%v) = %q, expected to contain %q", tc.m, result, s)
				}
			}
		})
	}
}

func TestStringSliceContains(t *testing.T) {
	testCases := []struct {
		slice    []string
		s        string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := StringSliceContains(tc.slice, tc.s)
			if result != tc.expected {
				t.Errorf("StringSliceContains(%v, %q) = %v, expected %v", tc.slice, tc.s, result, tc.expected)
			}
		})
	}
}

func TestStringSliceEquals(t *testing.T) {
	testCases := []struct {
		a        []string
		b        []string
		expected bool
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{[]string{"a", "b", "c"}, []string{"a", "b", "d"}, false},
		{[]string{"a", "b"}, []string{"a", "b", "c"}, false},
		{[]string{}, []string{}, true},
		{nil, nil, true},
		{nil, []string{}, true}, // nil and empty slice are considered equal
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := StringSliceEquals(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("StringSliceEquals(%v, %v) = %v, expected %v", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	testCases := []struct {
		slice    []string
		expected []string
	}{
		{[]string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{}, []string{}},
		{nil, nil},
		{[]string{"a"}, []string{"a"}},
	}

	for i, tc := range testCases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result := RemoveDuplicates(tc.slice)
			if len(result) != len(tc.expected) {
				t.Errorf("RemoveDuplicates(%v) = %v, expected %v", tc.slice, result, tc.expected)
				return
			}

			for j := range result {
				if result[j] != tc.expected[j] {
					t.Errorf("RemoveDuplicates(%v)[%d] = %q, expected %q", tc.slice, j, result[j], tc.expected[j])
				}
			}
		})
	}
}
