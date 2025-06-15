package networkinfo

// NetworkInfoResponse represents the top-level response from the RIPEstat network-info endpoint.
type NetworkInfoResponse struct {
	Messages       []interface{}   `json:"messages"`
	SeeAlso        []interface{}   `json:"see_also"`
	Version        string          `json:"version"`
	DataCallName   string          `json:"data_call_name"`
	DataCallStatus string          `json:"data_call_status"`
	Cached         bool            `json:"cached"`
	Data           NetworkInfoData `json:"data"`
	QueryID        string          `json:"query_id"`
	ProcessTime    int             `json:"process_time"`
	ServerID       string          `json:"server_id"`
	BuildVersion   string          `json:"build_version"`
	Status         string          `json:"status"`
	StatusCode     int             `json:"status_code"`
	Time           string          `json:"time"`
}

// NetworkInfoData represents the 'data' field in the response.
type NetworkInfoData struct {
	ASNs   []string `json:"asns"`
	Prefix string   `json:"prefix"`
}
