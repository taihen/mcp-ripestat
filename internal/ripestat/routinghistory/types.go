package routinghistory

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat routing-history endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	ByOrigin []OriginData `json:"by_origin"`
	Resource string       `json:"resource"`
}

// OriginData represents routing data for a specific origin ASN.
type OriginData struct {
	Origin   string       `json:"origin"`
	Prefixes []PrefixData `json:"prefixes"`
}

// PrefixData represents routing history for a specific prefix.
type PrefixData struct {
	Prefix   string          `json:"prefix"`
	Timeline []TimelineEntry `json:"timeline"`
}

// TimelineEntry represents a single point in the routing history timeline.
type TimelineEntry struct {
	StartTime       string `json:"starttime"`
	EndTime         string `json:"endtime"`
	FullPeersSeeing int    `json:"full_peers_seeing"`
}
