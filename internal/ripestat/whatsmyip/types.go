package whatsmyip

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	IP string `json:"ip"`
}

type APIResponse struct {
	IP        string `json:"ip"`
	FetchedAt string `json:"fetched_at"`
}
