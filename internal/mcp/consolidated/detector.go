package consolidated

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	asnPattern     = regexp.MustCompile(`^AS?(\d+)$`)
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
	cidrV4Pattern  = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
	cidrV6Pattern  = regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}/\d{1,3}$`)
)

func DetectResource(input string) (*DetectedResource, error) {
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	input = strings.TrimSpace(input)
	original := input

	normalizedInput := strings.ToUpper(input)

	resource := &DetectedResource{
		Original:  original,
		Value:     input,
		Validated: false,
	}

	if matches := asnPattern.FindStringSubmatch(normalizedInput); matches != nil {
		asn, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid ASN number: %s", input)
		}
		if asn < 1 || asn > 4294967295 {
			return nil, fmt.Errorf("ASN out of valid range: %d", asn)
		}

		resource.Type = ASN
		resource.Value = fmt.Sprintf("AS%d", asn)
		resource.Validated = true
		return resource, nil
	}

	if countryPattern.MatchString(normalizedInput) {
		resource.Type = Country
		resource.Value = normalizedInput
		resource.Validated = true
		return resource, nil
	}

	if cidrV4Pattern.MatchString(input) || cidrV6Pattern.MatchString(input) {
		_, network, err := net.ParseCIDR(input)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR: %s", input)
		}
		resource.Type = IPPrefix
		resource.Value = network.String()
		if strings.Contains(input, ":") {
			resource.Version = 6
		} else {
			resource.Version = 4
		}
		resource.Validated = true
		return resource, nil
	}

	if ip := net.ParseIP(input); ip != nil {
		resource.Type = IPAddress
		resource.Validated = true
		if ip.To4() != nil {
			resource.Version = 4
		} else {
			resource.Version = 6
		}
		return resource, nil
	}

	resource.Type = Invalid
	return resource, fmt.Errorf("unrecognized resource format: %s", input)
}

func GetIPVersion(resource string) (int, error) {
	if ip := net.ParseIP(resource); ip != nil {
		if ip.To4() != nil {
			return 4, nil
		}
		return 6, nil
	}

	if _, _, err := net.ParseCIDR(resource); err == nil {
		if strings.Contains(resource, ":") {
			return 6, nil
		}
		return 4, nil
	}

	return 0, fmt.Errorf("not an IP address or prefix: %s", resource)
}
