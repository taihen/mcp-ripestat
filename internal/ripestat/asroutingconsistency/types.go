// Package asroutingconsistency provides access to the RIPEstat as-routing-consistency API.
package asroutingconsistency

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat AS Routing Consistency API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the AS Routing Consistency response.
type Data struct {
	Prefixes []Prefix `json:"prefixes"`
	Imports  []Import `json:"imports"`
}

// Prefix represents a prefix in the routing consistency data.
type Prefix struct {
	InBGP      bool     `json:"in_bgp"`
	InWhois    bool     `json:"in_whois"`
	IRRSources []string `json:"irr_sources"`
	Prefix     string   `json:"prefix"`
}

// Import represents an import in the routing consistency data.
type Import struct {
	InBGP   bool `json:"in_bgp"`
	InWhois bool `json:"in_whois"`
	Peer    int  `json:"peer"`
}
