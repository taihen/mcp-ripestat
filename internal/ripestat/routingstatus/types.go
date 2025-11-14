
package routingstatus

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	Resource  string   `json:"resource"`
	Announced bool     `json:"announced"`
	ASNs      []string `json:"asns"`
	QueryTime string   `json:"query_time"`
}
