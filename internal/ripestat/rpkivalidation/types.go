// Package rpkivalidation provides access to the RIPEstat rpki-validation API.
package rpkivalidation

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat rpki-validation endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	ValidatingROAs []ValidatingROA `json:"validating_roas"`
	Status         string          `json:"status"`
	Validator      string          `json:"validator"`
	Resource       string          `json:"resource"`
	Prefix         string          `json:"prefix"`
}

// ValidatingROA represents a single validating ROA entry.
type ValidatingROA struct {
	Origin    string `json:"origin"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"`
	Validity  string `json:"validity"`
}

// APIResponse represents the response format expected by the MCP client.
type APIResponse struct {
	Status         string          `json:"status"`
	Validator      string          `json:"validator"`
	Resource       string          `json:"resource"`
	Prefix         string          `json:"prefix"`
	ValidatingROAs []ValidatingROA `json:"validating_roas"`
	FetchedAt      string          `json:"fetched_at"`
}
