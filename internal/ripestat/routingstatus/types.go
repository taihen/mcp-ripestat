// Package routingstatus provides access to the RIPEstat routing-status API.
package routingstatus

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat routing-status endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Resource  string   `json:"resource"`
	Announced bool     `json:"announced"`
	ASNs      []string `json:"asns"`
	QueryTime string   `json:"query_time"`
}
