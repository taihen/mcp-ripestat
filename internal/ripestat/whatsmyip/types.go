// Package whatsmyip provides access to the RIPEstat whats-my-ip API.
package whatsmyip

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat whats-my-ip endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	IP string `json:"ip"`
}

// APIResponse represents the response format expected by the MCP client.
type APIResponse struct {
	IP        string `json:"ip"`
	FetchedAt string `json:"fetched_at"`
}
