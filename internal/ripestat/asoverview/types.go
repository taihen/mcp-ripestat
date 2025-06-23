// Package asoverview provides access to the RIPEstat as-overview API.
package asoverview

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat AS Overview API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the AS Overview response.
type Data struct {
	Type           string `json:"type"`
	Resource       string `json:"resource"`
	Block          Block  `json:"block"`
	Holder         string `json:"holder"`
	Announced      bool   `json:"announced"`
	QueryStartTime string `json:"query_starttime"`
	QueryEndTime   string `json:"query_endtime"`
}

// Block contains information about the AS block.
type Block struct {
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
}
