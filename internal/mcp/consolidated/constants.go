package consolidated

const (
	DepthBasic         = "basic"
	DepthDetailed      = "detailed"
	DepthComprehensive = "comprehensive"
)

const (
	AnalysisConsistency      = "consistency"
	AnalysisPathOptimization = "path-optimization"
	AnalysisUpdates          = "updates"
	AnalysisLookingGlass     = "looking-glass"
)

const (
	DataTypeWhois             = "whois"
	DataTypeAllocationHistory = "allocation-history"
	DataTypeHierarchy         = "hierarchy"
	DataTypeContacts          = "contacts"
)

const (
	SecurityCheckRPKI          = "rpki"
	SecurityCheckAbuseContacts = "abuse-contacts"
	SecurityCheckBGPHijacking  = "bgp-hijacking"
)

const (
	RelationshipNeighbors         = "neighbors"
	RelationshipAnnouncedPrefixes = "announced-prefixes"
	RelationshipRelatedNetworks   = "related-networks"
)

const (
	LocationTypeASNs       = "asns"
	LocationTypePrefixes   = "prefixes"
	LocationTypeStatistics = "statistics"
)

const (
	TimeframeCurrent = "current"
	Timeframe1Day    = "1d"
	Timeframe1Week   = "1w"
	Timeframe1Month  = "1m"
)

const (
	ScopeDirect   = "direct"
	ScopeExtended = "extended"
)

const (
	FormatSummary  = "summary"
	FormatDetailed = "detailed"
)
