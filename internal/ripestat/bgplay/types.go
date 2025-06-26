package bgplay

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the top-level response from the RIPEstat bgplay endpoint.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the 'data' field in the response.
type Data struct {
	Resource       string        `json:"resource"`
	QueryStartTime string        `json:"query_starttime"`
	QueryEndTime   string        `json:"query_endtime"`
	InitialState   []RouteRecord `json:"initial_state"`
	Events         []Event       `json:"events"`
	TargetPrefix   string        `json:"target_prefix,omitempty"`
	RRCs           []int         `json:"rrcs,omitempty"`
	Nodes          []interface{} `json:"nodes,omitempty"`
	Sources        []interface{} `json:"sources,omitempty"`
	Targets        []interface{} `json:"targets,omitempty"`
}

// RouteRecord represents a BGP route record.
type RouteRecord struct {
	TargetPrefix string   `json:"target_prefix"`
	SourceID     string   `json:"source_id"`
	Path         []int    `json:"path"`
	Community    []string `json:"community"`
}

// Event represents a BGP event in the timeline.
type Event struct {
	Type      string           `json:"type"`
	Timestamp types.CustomTime `json:"timestamp"`
	Attrs     EventAttrs       `json:"attrs"`
}

// EventAttrs represents event attributes.
type EventAttrs struct {
	TargetPrefix string   `json:"target_prefix"`
	SourceID     string   `json:"source_id"`
	Path         []int    `json:"path"`
	Community    []string `json:"community"`
}
