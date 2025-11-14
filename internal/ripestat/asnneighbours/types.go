package asnneighbours

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Resource        string          `json:"resource"`
	QueryStartTime  string          `json:"query_starttime"`
	QueryEndTime    string          `json:"query_endtime"`
	LatestTime      string          `json:"latest_time"`
	EarliestTime    string          `json:"earliest_time"`
	NeighbourCounts NeighbourCounts `json:"neighbour_counts"`
	Neighbours      []Neighbour     `json:"neighbours"`
}

type NeighbourCounts struct {
	Left      int `json:"left"`
	Right     int `json:"right"`
	Unique    int `json:"unique"`
	Uncertain int `json:"uncertain"`
}

type Neighbour struct {
	ASN     int    `json:"asn"`
	Type    string `json:"type"`
	Power   *int   `json:"power,omitempty"`
	V4Peers *int   `json:"v4_peers,omitempty"`
	V6Peers *int   `json:"v6_peers,omitempty"`
}

type APIResponse struct {
	Resource        string          `json:"resource"`
	QueryTime       string          `json:"query_time"`
	NeighbourCounts NeighbourCounts `json:"neighbour_counts"`
	Neighbours      []Neighbour     `json:"neighbours"`
	FetchedAt       string          `json:"fetched_at"`
}
