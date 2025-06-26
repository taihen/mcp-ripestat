package rpkihistory

import (
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response represents the RIPEStat RPKI History response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the data field in the RPKI History response.
type Data struct {
	Timeseries []TimeseriesEntry `json:"timeseries"`
}

// TimeseriesEntry represents a single entry in the RPKI History timeseries.
type TimeseriesEntry struct {
	Prefix    string    `json:"prefix"`
	Time      time.Time `json:"time"`
	VRPCount  int       `json:"vrp_count"`
	Count     int       `json:"count"`
	Family    int       `json:"family"`
	MaxLength int       `json:"max_length"`
}
