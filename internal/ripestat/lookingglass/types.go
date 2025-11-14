package lookingglass

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	RRCs []RRC `json:"rrcs"`
}

type RRC struct {
	RRC      string `json:"rrc"`
	Location string `json:"location"`
	Peers    []Peer `json:"peers"`
}

type Peer struct {
	ASNOrigin         string `json:"asn_origin"`
	ASPath            string `json:"as_path"`
	Community         string `json:"community"`
	LargeCommunity    string `json:"largeCommunity"`
	ExtendedCommunity string `json:"extendedCommunity"`
	LastUpdated       string `json:"last_updated"`
	Prefix            string `json:"prefix"`
	Peer              string `json:"peer"`
	Origin            string `json:"origin"`
	NextHop           string `json:"next_hop"`
	LatestTime        string `json:"latest_time"`
}

type APIResponse struct {
	RRCs      []RRC  `json:"rrcs"`
	FetchedAt string `json:"fetched_at"`
}
