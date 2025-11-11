package consolidated

// Depth levels for investigation detail.
const (
	DepthBasic         = "basic"
	DepthDetailed      = "detailed"
	DepthComprehensive = "comprehensive"
)

// Analysis types for routing analysis.
const (
	AnalysisConsistency      = "consistency"
	AnalysisPathOptimization = "path-optimization"
	AnalysisUpdates          = "updates"
	AnalysisLookingGlass     = "looking-glass"
)

// Registry data types.
const (
	DataTypeWhois             = "whois"
	DataTypeAllocationHistory = "allocation-history"
	DataTypeHierarchy         = "hierarchy"
	DataTypeContacts          = "contacts"
)

// Security check types.
const (
	SecurityCheckRPKI          = "rpki"
	SecurityCheckAbuseContacts = "abuse-contacts"
	SecurityCheckBGPHijacking  = "bgp-hijacking"
)

// Relationship types.
const (
	RelationshipNeighbors         = "neighbors"
	RelationshipAnnouncedPrefixes = "announced-prefixes"
	RelationshipRelatedNetworks   = "related-networks"
)

// Location search types.
const (
	LocationTypeASNs       = "asns"
	LocationTypePrefixes   = "prefixes"
	LocationTypeStatistics = "statistics"
)

// Timeframe options.
const (
	TimeframeCurrent = "current"
	Timeframe1Day    = "1d"
	Timeframe1Week   = "1w"
	Timeframe1Month  = "1m"
)

// Scope options.
const (
	ScopeDirect   = "direct"
	ScopeExtended = "extended"
)

// Format options.
const (
	FormatSummary  = "summary"
	FormatDetailed = "detailed"
)
