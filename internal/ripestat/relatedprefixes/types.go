package relatedprefixes

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Resource  string   `json:"resource"`
	Prefixes  []Prefix `json:"prefixes"`
	QueryTime string   `json:"query_time"`
}

type Prefix struct {
	Prefix       string `json:"prefix"`
	OriginASN    string `json:"origin_asn"`
	ASNName      string `json:"asn_name"`
	Relationship string `json:"relationship"`
}
