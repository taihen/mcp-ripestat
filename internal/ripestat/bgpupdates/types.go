package bgpupdates

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Resource       string      `json:"resource"`
	QueryStartTime CustomTime  `json:"query_starttime"`
	QueryEndTime   CustomTime  `json:"query_endtime"`
	Updates        []BGPUpdate `json:"updates"`
	NumUpdates     int         `json:"nr_updates"`
}

type BGPUpdate struct {
	Sequence   int64            `json:"seq"`
	Timestamp  CustomTime       `json:"timestamp"`
	Type       string           `json:"type"`
	Attributes UpdateAttributes `json:"attrs"`
}

type UpdateAttributes struct {
	SourceID     string   `json:"source_id"`
	TargetPrefix string   `json:"target_prefix"`
	Path         []int    `json:"path"`
	Community    []string `json:"community,omitempty"`
}

// CustomTime handles the custom timestamp format used by the BGP Updates API.
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for CustomTime.
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}

	// Try multiple time formats that the API might return
	formats := []string{
		"2006-01-02T15:04:05",      // Most common format without timezone
		"2006-01-02T15:04:05Z",     // With Z timezone
		"2006-01-02T15:04:05.000Z", // With milliseconds and Z timezone
		time.RFC3339,               // Standard RFC3339 format
		time.RFC3339Nano,           // RFC3339 with nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			ct.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time %q", timeStr)
}
