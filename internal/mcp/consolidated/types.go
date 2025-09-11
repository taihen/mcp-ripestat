package consolidated

// ResourceType represents the type of resource being analyzed
type ResourceType int

const (
	IPAddress ResourceType = iota
	IPPrefix
	ASN
	Country
	Invalid
)

// String returns the string representation of ResourceType
func (rt ResourceType) String() string {
	switch rt {
	case IPAddress:
		return "ip_address"
	case IPPrefix:
		return "ip_prefix"
	case ASN:
		return "asn"
	case Country:
		return "country"
	default:
		return "invalid"
	}
}

// DetectedResource contains information about a detected resource
type DetectedResource struct {
	Type      ResourceType `json:"type"`
	Value     string       `json:"value"`
	Version   int          `json:"version,omitempty"` // 4 or 6 for IP addresses
	Validated bool         `json:"validated"`
	Original  string       `json:"original"`
}

// Operation represents a high-level operation that can be performed on a resource
type Operation string

const (
	OpOverview      Operation = "overview"
	OpRouting       Operation = "routing"
	OpSecurity      Operation = "security"
	OpHistory       Operation = "history"
	OpNeighbors     Operation = "neighbors"
	OpConsistency   Operation = "consistency"
	OpUpdates       Operation = "updates"
	OpLookingGlass  Operation = "looking-glass"
	OpRelationships Operation = "relationships"
	OpHierarchy     Operation = "hierarchy"
)

// RouteResult contains the routing information for executing operations
type RouteResult struct {
	Endpoints    []string            `json:"endpoints"`
	Order        []int               `json:"order"`
	Dependencies map[string][]string `json:"dependencies"`
}

// ConsolidatedResult represents the aggregated result from multiple endpoint calls
type ConsolidatedResult struct {
	Resource   *DetectedResource      `json:"resource"`
	Operations []Operation            `json:"operations"`
	Results    map[string]interface{} `json:"results"`
	Errors     map[string]string      `json:"errors,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}