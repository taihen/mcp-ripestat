package bgplay

import (
	"encoding/json"

	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)


type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}


type Data struct {
	Resource       string          `json:"resource"`
	QueryStartTime string          `json:"query_starttime"`
	QueryEndTime   string          `json:"query_endtime"`
	InitialState   []RouteRecord   `json:"initial_state"`
	Events         []Event         `json:"events"`
	TargetPrefix   string          `json:"target_prefix,omitempty"`
	RRCs           []int           `json:"rrcs,omitempty"`
	Nodes          json.RawMessage `json:"nodes,omitempty"`
	Sources        json.RawMessage `json:"sources,omitempty"`
	Targets        json.RawMessage `json:"targets,omitempty"`
}


type RouteRecord struct {
	TargetPrefix string   `json:"target_prefix"`
	SourceID     string   `json:"source_id"`
	Path         []int    `json:"path"`
	Community    []string `json:"community"`
}


type Event struct {
	Type      string           `json:"type"`
	Timestamp types.CustomTime `json:"timestamp"`
	Attrs     EventAttrs       `json:"attrs"`
}


type EventAttrs struct {
	TargetPrefix string   `json:"target_prefix"`
	SourceID     string   `json:"source_id"`
	Path         []int    `json:"path"`
	Community    []string `json:"community"`
}
