
package networkinfo

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	ASNs   []interface{} `json:"asns"`
	Prefix string        `json:"prefix"`
}
