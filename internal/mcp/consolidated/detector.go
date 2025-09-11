package consolidated

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Regex patterns for resource detection
	asnPattern     = regexp.MustCompile(`^AS?(\d+)$`)
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
	ipv4Pattern    = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	ipv6Pattern    = regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}$`)
	cidrV4Pattern  = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
	cidrV6Pattern  = regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}/\d{1,3}$`)
)

// DetectResource analyzes input string and returns detected resource information
func DetectResource(input string) (*DetectedResource, error) {
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	input = strings.TrimSpace(input)
	original := input

	// Normalize ASN input (remove AS prefix if present)
	normalizedInput := strings.ToUpper(input)

	resource := &DetectedResource{
		Original:  original,
		Value:     input,
		Validated: false,
	}

	// Check for ASN
	if matches := asnPattern.FindStringSubmatch(normalizedInput); matches != nil {
		asn, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid ASN number: %s", input)
		}
		if asn < 1 || asn > 4294967295 { // 32-bit ASN range
			return nil, fmt.Errorf("ASN out of valid range: %d", asn)
		}
		
		resource.Type = ASN
		resource.Value = fmt.Sprintf("AS%d", asn)
		resource.Validated = true
		return resource, nil
	}

	// Check for Country Code
	if countryPattern.MatchString(normalizedInput) {
		resource.Type = Country
		resource.Value = normalizedInput
		resource.Validated = true
		return resource, nil
	}

	// Check for IPv4 CIDR
	if cidrV4Pattern.MatchString(input) {
		if _, _, err := net.ParseCIDR(input); err != nil {
			return nil, fmt.Errorf("invalid IPv4 CIDR: %s", input)
		}
		resource.Type = IPPrefix
		resource.Version = 4
		resource.Validated = true
		return resource, nil
	}

	// Check for IPv6 CIDR
	if cidrV6Pattern.MatchString(input) {
		if _, _, err := net.ParseCIDR(input); err != nil {
			return nil, fmt.Errorf("invalid IPv6 CIDR: %s", input)
		}
		resource.Type = IPPrefix
		resource.Version = 6
		resource.Validated = true
		return resource, nil
	}

	// Check for IPv4 address
	if ipv4Pattern.MatchString(input) {
		ip := net.ParseIP(input)
		if ip == nil {
			return nil, fmt.Errorf("invalid IPv4 address: %s", input)
		}
		resource.Type = IPAddress
		resource.Version = 4
		resource.Validated = true
		return resource, nil
	}

	// Check for IPv6 address
	if ipv6Pattern.MatchString(input) || strings.Contains(input, ":") {
		ip := net.ParseIP(input)
		if ip == nil {
			return nil, fmt.Errorf("invalid IPv6 address: %s", input)
		}
		resource.Type = IPAddress
		resource.Version = 6
		resource.Validated = true
		return resource, nil
	}

	// If nothing matches, return invalid
	resource.Type = Invalid
	return resource, fmt.Errorf("unrecognized resource format: %s", input)
}

// ValidateASN validates ASN format and range
func ValidateASN(asn string) error {
	matches := asnPattern.FindStringSubmatch(strings.ToUpper(asn))
	if matches == nil {
		return fmt.Errorf("invalid ASN format: %s", asn)
	}

	asnNum, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("invalid ASN number: %s", asn)
	}

	if asnNum < 1 || asnNum > 4294967295 {
		return fmt.Errorf("ASN out of valid range: %d", asnNum)
	}

	return nil
}

// ValidateIPAddress validates IP address format
func ValidateIPAddress(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// ValidateIPPrefix validates IP prefix/CIDR format
func ValidateIPPrefix(prefix string) error {
	_, _, err := net.ParseCIDR(prefix)
	if err != nil {
		return fmt.Errorf("invalid IP prefix: %s", prefix)
	}
	return nil
}

// GetIPVersion returns the IP version (4 or 6) for IP addresses and prefixes
func GetIPVersion(resource string) (int, error) {
	// Try parsing as IP first
	if ip := net.ParseIP(resource); ip != nil {
		if ip.To4() != nil {
			return 4, nil
		}
		return 6, nil
	}

	// Try parsing as CIDR
	if _, _, err := net.ParseCIDR(resource); err == nil {
		if strings.Contains(resource, ":") {
			return 6, nil
		}
		return 4, nil
	}

	return 0, fmt.Errorf("not an IP address or prefix: %s", resource)
}