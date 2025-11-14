
package countryasns

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	Countries  []Country `json:"countries"`
	Resource   []string  `json:"resource"`
	QueryTime  string    `json:"query_time"`
	LOD        []string  `json:"lod"`
	Cache      string    `json:"cache"`
	LatestTime string    `json:"latest_time"`
}


type Country struct {
	Stats     Stats  `json:"stats"`
	Resource  string `json:"resource"`
	Routed    string `json:"routed,omitempty"`
	NonRouted string `json:"non_routed,omitempty"`
}


type Stats struct {
	Registered int `json:"registered"`
	Routed     int `json:"routed"`
}
