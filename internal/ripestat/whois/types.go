// Package whois provides access to the RIPEstat whois API.
package whois

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat whois endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Records     [][]Record `json:"records"`
	IRRRecords  []Record   `json:"irr_records"`
	Authorities []string   `json:"authorities"`
	Resource    string     `json:"resource"`
	QueryTime   string     `json:"query_time"`
}

// Record represents a single whois record entry.
type Record struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	DetailsLink *string `json:"details_link"`
}
