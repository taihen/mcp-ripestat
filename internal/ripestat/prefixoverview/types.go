package prefixoverview

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	IsLessSpecific   bool            `json:"is_less_specific"`
	Announced        bool            `json:"announced"`
	ASNs             []ASN           `json:"asns"`
	RelatedPrefixes  []RelatedPrefix `json:"related_prefixes"`
	Resource         string          `json:"resource"`
	Type             string          `json:"type"`
	Block            Block           `json:"block"`
	ActualNumRelated int             `json:"actual_num_related"`
	QueryTime        string          `json:"query_time"`
	NumFilteredOut   int             `json:"num_filtered_out"`
}

type ASN struct {
	ASN    int    `json:"asn"`
	Holder string `json:"holder"`
}

type RelatedPrefix struct {
	Resource string `json:"resource"`
	Type     string `json:"type"`
}

type Block struct {
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
}
