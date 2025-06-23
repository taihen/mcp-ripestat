// Package networkinfo provides access to the RIPEstat network-info API.
package networkinfo

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat network-info endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	ASNs   []interface{} `json:"asns"`
	Prefix string        `json:"prefix"`
}
