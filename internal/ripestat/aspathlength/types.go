package aspathlength

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Stats     []Stat `json:"stats"`
	Resource  string `json:"resource"`
	QueryTime string `json:"query_time"`
	SortBy    string `json:"sort_by"`
}

type Stat struct {
	Number     int       `json:"number"`
	Count      int       `json:"count"`
	Location   string    `json:"location"`
	Stripped   PathStats `json:"stripped"`
	Unstripped PathStats `json:"unstripped"`
}

type PathStats struct {
	Sum int     `json:"sum"`
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}
