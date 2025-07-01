// Package prefixroutingconsistency provides access to the RIPEstat prefix-routing-consistency API.
package prefixroutingconsistency

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat prefix-routing-consistency endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Resource       string     `json:"resource"`
	Routes         []Route    `json:"routes"`
	Parameters     Parameters `json:"parameters"`
	QueryStartTime string     `json:"query_starttime"`
	QueryEndTime   string     `json:"query_endtime"`
}

// Route represents routing information for a prefix.
type Route struct {
	InBGP      bool     `json:"in_bgp"`
	InWHOIS    bool     `json:"in_whois"`
	Prefix     string   `json:"prefix"`
	Origin     int      `json:"origin"`
	IRRSources []string `json:"irr_sources"`
	ASNName    string   `json:"asn_name"`
}

// Parameters represents the parameters used in the query.
type Parameters struct {
	Resource          string `json:"resource"`
	DataOverloadLimit string `json:"data_overload_limit"`
}
