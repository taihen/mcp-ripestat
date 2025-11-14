package rpkihistory

import (
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Timeseries []TimeseriesEntry `json:"timeseries"`
}

type TimeseriesEntry struct {
	Prefix    string    `json:"prefix"`
	Time      time.Time `json:"time"`
	VRPCount  int       `json:"vrp_count"`
	Count     int       `json:"count"`
	Family    int       `json:"family"`
	MaxLength int       `json:"max_length"`
}
