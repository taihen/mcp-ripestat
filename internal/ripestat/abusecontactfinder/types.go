// Package abusecontactfinder provides access to the RIPEstat abuse-contact-finder API.
package abusecontactfinder

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat abuse-contact-finder endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	AbuseContacts    []string `json:"abuse_contacts"`
	AuthoritativeRIR string   `json:"authoritative_rir"`
	LatestTime       string   `json:"latest_time"`
	EarliestTime     string   `json:"earliest_time"`
	Parameters       struct {
		Resource string      `json:"resource"`
		Cache    interface{} `json:"cache"`
	} `json:"parameters"`
}

// APIResponse represents the response format expected by the MCP client.
type APIResponse struct {
	Contacts  []string `json:"contacts"`
	FetchedAt string   `json:"fetched_at"`
}
