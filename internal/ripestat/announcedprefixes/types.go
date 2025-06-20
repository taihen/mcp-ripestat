package announcedprefixes

// AnnouncedPrefixesResponse represents the top-level response from the RIPEstat announced-prefixes endpoint.
type AnnouncedPrefixesResponse struct {
	Messages       [][]string            `json:"messages"`
	SeeAlso        []interface{}         `json:"see_also"`
	Version        string                `json:"version"`
	DataCallName   string                `json:"data_call_name"`
	DataCallStatus string                `json:"data_call_status"`
	Cached         bool                  `json:"cached"`
	Data           AnnouncedPrefixesData `json:"data"`
	QueryID        string                `json:"query_id"`
	ProcessTime    int                   `json:"process_time"`
	ServerID       string                `json:"server_id"`
	BuildVersion   string                `json:"build_version"`
	Status         string                `json:"status"`
	StatusCode     int                   `json:"status_code"`
	Time           string                `json:"time"`
}

// AnnouncedPrefixesData represents the 'data' field in the response.
type AnnouncedPrefixesData struct {
	Prefixes       []Prefix `json:"prefixes"`
	QueryStarttime string   `json:"query_starttime"`
	QueryEndtime   string   `json:"query_endtime"`
	Resource       string   `json:"resource"`
	LatestTime     string   `json:"latest_time"`
	EarliestTime   string   `json:"earliest_time"`
}

// Prefix represents a single prefix entry with its visibility timeline.
type Prefix struct {
	Prefix    string     `json:"prefix"`
	Timelines []Timeline `json:"timelines"`
}

// Timeline represents the period when a prefix was announced.
type Timeline struct {
	Starttime string `json:"starttime"`
	Endtime   string `json:"endtime"`
}
