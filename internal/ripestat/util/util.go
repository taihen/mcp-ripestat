
package util

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)


func IsValidIPv4(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}


	return parsedIP.To4() != nil
}


func IsValidIPv6(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}


	return parsedIP.To4() == nil
}


func IsValidIP(ip string) bool {
	return IsValidIPv4(ip) || IsValidIPv6(ip)
}


func IsValidCIDR(cidr string) bool {
	if cidr == "" {
		return false
	}

	_, _, err := net.ParseCIDR(cidr)

	return err == nil
}



func IsValidASN(asn string) bool {
	if asn == "" {
		return false
	}


	asn = strings.TrimPrefix(strings.ToUpper(asn), "AS")


	asnRegex := regexp.MustCompile(`^[0-9]+$`)

	return asnRegex.MatchString(asn)
}


func FormatASN(asn string) string {
	if asn == "" {
		return ""
	}


	asn = strings.TrimPrefix(strings.ToUpper(asn), "AS")


	return "AS" + asn
}



func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return t.Format(layout)
}



func ParseTime(s string, layout string) (time.Time, error) {
	if layout == "" {
		layout = time.RFC3339
	}

	return time.Parse(layout, s)
}


func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}



func JoinStrings(strs []string, sep string) string {
	var nonEmpty []string

	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	return strings.Join(nonEmpty, sep)
}


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


func StringSliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}


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
