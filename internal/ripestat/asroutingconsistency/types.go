package asroutingconsistency

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Prefixes []Prefix `json:"prefixes"`
	Imports  []Import `json:"imports"`
}

type Prefix struct {
	InBGP      bool     `json:"in_bgp"`
	InWhois    bool     `json:"in_whois"`
	IRRSources []string `json:"irr_sources"`
	Prefix     string   `json:"prefix"`
}

type Import struct {
	InBGP   bool `json:"in_bgp"`
	InWhois bool `json:"in_whois"`
	Peer    int  `json:"peer"`
}
