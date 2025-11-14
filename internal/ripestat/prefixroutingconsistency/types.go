package prefixroutingconsistency

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Resource       string     `json:"resource"`
	Routes         []Route    `json:"routes"`
	Parameters     Parameters `json:"parameters"`
	QueryStartTime string     `json:"query_starttime"`
	QueryEndTime   string     `json:"query_endtime"`
}

type Route struct {
	InBGP      bool     `json:"in_bgp"`
	InWHOIS    bool     `json:"in_whois"`
	Prefix     string   `json:"prefix"`
	Origin     int      `json:"origin"`
	IRRSources []string `json:"irr_sources"`
	ASNName    string   `json:"asn_name"`
}

type Parameters struct {
	Resource          string `json:"resource"`
	DataOverloadLimit string `json:"data_overload_limit"`
}
