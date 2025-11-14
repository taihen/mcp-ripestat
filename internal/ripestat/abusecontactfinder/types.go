
package abusecontactfinder

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


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


type APIResponse struct {
	Contacts  []string `json:"contacts"`
	FetchedAt string   `json:"fetched_at"`
}
