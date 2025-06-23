// Package announcedprefixes provides access to the RIPEstat announced-prefixes API.
package announcedprefixes

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat announced-prefixes endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Resource  string   `json:"resource"`
	Prefixes  []Prefix `json:"prefixes"`
	QueryTime string   `json:"query_time"`
}

// Prefix represents a single prefix announced by an AS.
type Prefix struct {
	Prefix    string     `json:"prefix"`
	Timelines []Timeline `json:"timelines"`
}

// Timeline represents the visibility timeline of a prefix.
type Timeline struct {
	StartTime types.CustomTime `json:"starttime"`
	EndTime   types.CustomTime `json:"endtime"`
}
