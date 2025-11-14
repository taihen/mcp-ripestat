package routinghistory

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	ByOrigin []OriginData `json:"by_origin"`
	Resource string       `json:"resource"`
}


type OriginData struct {
	Origin   string       `json:"origin"`
	Prefixes []PrefixData `json:"prefixes"`
}


type PrefixData struct {
	Prefix    string          `json:"prefix"`
	Timelines []TimelineEntry `json:"timelines"`
}


type TimelineEntry struct {
	StartTime       types.CustomTime `json:"starttime"`
	EndTime         types.CustomTime `json:"endtime"`
	FullPeersSeeing float64          `json:"full_peers_seeing"`
}
