package aspathlength

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the AS Path Length API response structure.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the data field in the AS Path Length response.
type Data struct {
	Stats     []Stat `json:"stats"`
	Resource  string `json:"resource"`
	QueryTime string `json:"query_time"`
	SortBy    string `json:"sort_by"`
}

// Stat represents a single AS path length statistic entry.
type Stat struct {
	Number     int       `json:"number"`
	Count      int       `json:"count"`
	Location   string    `json:"location"`
	Stripped   PathStats `json:"stripped"`
	Unstripped PathStats `json:"unstripped"`
}

// PathStats represents the path length statistics.
type PathStats struct {
	Sum int     `json:"sum"`
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}
