// Package asnneighbours provides access to the RIPEstat asn-neighbours API.
package asnneighbours

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat asn-neighbours endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Resource        string          `json:"resource"`
	QueryStartTime  string          `json:"query_starttime"`
	QueryEndTime    string          `json:"query_endtime"`
	LatestTime      string          `json:"latest_time"`
	EarliestTime    string          `json:"earliest_time"`
	NeighbourCounts NeighbourCounts `json:"neighbour_counts"`
	Neighbours      []Neighbour     `json:"neighbours"`
}

// NeighbourCounts represents the counts of different types of neighbours.
type NeighbourCounts struct {
	Left      int `json:"left"`
	Right     int `json:"right"`
	Unique    int `json:"unique"`
	Uncertain int `json:"uncertain"`
}

// Neighbour represents a single ASN neighbour.
type Neighbour struct {
	ASN     int    `json:"asn"`
	Type    string `json:"type"`
	Power   *int   `json:"power,omitempty"`    // Only present when lod=1
	V4Peers *int   `json:"v4_peers,omitempty"` // Only present when lod=1
	V6Peers *int   `json:"v6_peers,omitempty"` // Only present when lod=1
}

// APIResponse represents the response format expected by the MCP client.
type APIResponse struct {
	Resource        string          `json:"resource"`
	QueryTime       string          `json:"query_time"`
	NeighbourCounts NeighbourCounts `json:"neighbour_counts"`
	Neighbours      []Neighbour     `json:"neighbours"`
	FetchedAt       string          `json:"fetched_at"`
}
