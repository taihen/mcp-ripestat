// Package lookingglass provides access to the RIPEstat looking-glass API.
package lookingglass

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat looking-glass endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	RRCs []RRC `json:"rrcs"`
}

// RRC represents a Route Reflection Collector.
type RRC struct {
	RRC      string `json:"rrc"`
	Location string `json:"location"`
	Peers    []Peer `json:"peers"`
}

// Peer represents a BGP peer with route information.
type Peer struct {
	ASNOrigin         string `json:"asn_origin"`
	ASPath            string `json:"as_path"`
	Community         string `json:"community"`
	LargeCommunity    string `json:"largeCommunity"`
	ExtendedCommunity string `json:"extendedCommunity"`
	LastUpdated       string `json:"last_updated"`
	Prefix            string `json:"prefix"`
	Peer              string `json:"peer"`
	Origin            string `json:"origin"`
	NextHop           string `json:"next_hop"`
	LatestTime        string `json:"latest_time"`
}

// APIResponse represents the response format expected by the MCP client.
type APIResponse struct {
	RRCs      []RRC  `json:"rrcs"`
	FetchedAt string `json:"fetched_at"`
}
