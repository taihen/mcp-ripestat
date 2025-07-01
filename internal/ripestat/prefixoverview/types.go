// Package prefixoverview provides access to the RIPEstat prefix-overview API.
package prefixoverview

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat Prefix Overview API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the Prefix Overview response.
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

// ASN represents an ASN in the prefix overview.
type ASN struct {
	ASN    int    `json:"asn"`
	Holder string `json:"holder"`
}

// RelatedPrefix represents a related prefix.
type RelatedPrefix struct {
	Resource string `json:"resource"`
	Type     string `json:"type"`
}

// Block contains information about the prefix block.
type Block struct {
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
}
