// Package relatedprefixes provides access to the RIPEstat related-prefixes API.
package relatedprefixes

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat Related Prefixes API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the Related Prefixes response.
type Data struct {
	Resource  string   `json:"resource"`
	Prefixes  []Prefix `json:"prefixes"`
	QueryTime string   `json:"query_time"`
}

// Prefix represents a related prefix with its relationship information.
type Prefix struct {
	Prefix       string `json:"prefix"`
	OriginASN    string `json:"origin_asn"`
	ASNName      string `json:"asn_name"`
	Relationship string `json:"relationship"`
}
