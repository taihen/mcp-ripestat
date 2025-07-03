package allocationhistory

import "github.com/taihen/mcp-ripestat/internal/ripestat/types"

const EndpointPath = "/data/allocation-history/data.json"

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Results        map[string][]Result `json:"results"`
	Resource       string              `json:"resource"`
	QueryStartTime string              `json:"query_starttime"`
	QueryEndTime   string              `json:"query_endtime"`
}

type Result struct {
	Resource  string     `json:"resource"`
	Status    string     `json:"status"`
	Timelines []Timeline `json:"timelines"`
}

type Timeline struct {
	StartTime string `json:"starttime"`
	EndTime   string `json:"endtime"`
}
