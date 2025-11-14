
package bgpstate

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	Resource  string      `json:"resource"`
	Timestamp interface{} `json:"timestamp"`
	BGPState  []BGPRoute  `json:"bgp_state"`
}


type BGPRoute struct {
	TargetPrefix string   `json:"target_prefix"`
	SourceID     string   `json:"source_id"`
	Path         []int    `json:"path"`
	Community    []string `json:"community"`
}
