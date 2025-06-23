// Package util provides utility functions for the RIPEstat API client.
package util

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// IsValidIPv4 checks if the given string is a valid IPv4 address.
func IsValidIPv4(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check if it's an IPv4 address
	return parsedIP.To4() != nil
}

// IsValidIPv6 checks if the given string is a valid IPv6 address.
func IsValidIPv6(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check if it's an IPv6 address
	return parsedIP.To4() == nil
}

// IsValidIP checks if the given string is a valid IP address (either IPv4 or IPv6).
func IsValidIP(ip string) bool {
	return IsValidIPv4(ip) || IsValidIPv6(ip)
}

// IsValidCIDR checks if the given string is a valid CIDR notation.
func IsValidCIDR(cidr string) bool {
	if cidr == "" {
		return false
	}

	_, _, err := net.ParseCIDR(cidr)

	return err == nil
}

// IsValidASN checks if the given string is a valid Autonomous System Number.
// Valid formats: "AS123", "123".
func IsValidASN(asn string) bool {
	if asn == "" {
		return false
	}

	// Remove "AS" prefix if present
	asn = strings.TrimPrefix(strings.ToUpper(asn), "AS")

	// Check if the remaining string is a valid number
	asnRegex := regexp.MustCompile(`^[0-9]+$`)

	return asnRegex.MatchString(asn)
}

// FormatASN ensures the ASN is in the correct format (with "AS" prefix).
func FormatASN(asn string) string {
	if asn == "" {
		return ""
	}

	// Remove "AS" prefix if present
	asn = strings.TrimPrefix(strings.ToUpper(asn), "AS")

	// Add "AS" prefix
	return "AS" + asn
}

// FormatTime formats a time.Time value according to the specified layout.
// If layout is empty, RFC3339 is used.
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return t.Format(layout)
}

// ParseTime parses a time string according to the specified layout.
// If layout is empty, RFC3339 is used.
func ParseTime(s string, layout string) (time.Time, error) {
	if layout == "" {
		layout = time.RFC3339
	}

	return time.Parse(layout, s)
}

// TruncateString truncates a string to the specified length and adds an ellipsis if truncated.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// JoinStrings joins a slice of strings with the specified separator.
// It filters out empty strings.
func JoinStrings(strs []string, sep string) string {
	var nonEmpty []string

	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	return strings.Join(nonEmpty, sep)
}

// SplitAndTrim splits a string by the specified separator and trims whitespace from each part.
func SplitAndTrim(s string, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)

	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// MapToString converts a map to a string representation.
func MapToString(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// StringSliceContains checks if a slice of strings contains a specific string.
func StringSliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}

// StringSliceEquals checks if two string slices are equal (same elements in the same order).
func StringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// RemoveDuplicates removes duplicate strings from a slice while preserving order.
func RemoveDuplicates(slice []string) []string {
	if len(slice) <= 1 {
		return slice
	}

	seen := make(map[string]struct{}, len(slice))
	result := make([]string, 0, len(slice))

	for _, s := range slice {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}

			result = append(result, s)
		}
	}

	return result
}
