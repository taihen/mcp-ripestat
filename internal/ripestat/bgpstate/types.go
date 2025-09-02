// Package bgpstate provides access to the RIPEstat bgp-state API.
package bgpstate

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat BGP State API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the BGP State response.
type Data struct {
	Resource  string      `json:"resource"`
	Timestamp interface{} `json:"timestamp"` // Can be string or number depending on unix_timestamps parameter
	BGPState  []BGPRoute  `json:"bgp_state"`
}

// BGPRoute represents a single BGP route entry.
type BGPRoute struct {
	TargetPrefix string  `json:"target_prefix"`
	SourceID     string  `json:"source_id"`
	Path         []int   `json:"path"`
	Community    [][]int `json:"community"`
}
