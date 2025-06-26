// Package countryasns provides access to the RIPEstat country-asns API.
package countryasns

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat Country ASNs API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the Country ASNs response.
type Data struct {
	Countries  []Country `json:"countries"`
	Resource   []string  `json:"resource"`
	QueryTime  string    `json:"query_time"`
	LOD        []string  `json:"lod"`
	Cache      string    `json:"cache"`
	LatestTime string    `json:"latest_time"`
}

// Country contains information about ASNs for a specific country.
type Country struct {
	Stats     Stats  `json:"stats"`
	Resource  string `json:"resource"`
	Routed    string `json:"routed,omitempty"`
	NonRouted string `json:"non_routed,omitempty"`
}

// Stats contains statistics about registered and routed ASNs.
type Stats struct {
	Registered int `json:"registered"`
	Routed     int `json:"routed"`
}
