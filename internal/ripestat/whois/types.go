
package whois

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	Records     [][]Record `json:"records"`
	IRRRecords  [][]Record `json:"irr_records"`
	Authorities []string   `json:"authorities"`
	Resource    string     `json:"resource"`
	QueryTime   string     `json:"query_time"`
}


type Record struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	DetailsLink *string `json:"details_link"`
}
