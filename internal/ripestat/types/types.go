// Package types provides common types for the RIPEstat API.
package types

import (
	"strings"
	"time"
)

// BaseResponse represents the common fields in all RIPEstat API responses.
type BaseResponse struct {
	Messages       []interface{} `json:"messages"`
	SeeAlso        []interface{} `json:"see_also"`
	Version        string        `json:"version"`
	DataCallName   string        `json:"data_call_name"`
	DataCallStatus string        `json:"data_call_status"`
	Cached         bool          `json:"cached"`
	QueryID        string        `json:"query_id"`
	ProcessTime    int           `json:"process_time"`
	ServerID       string        `json:"server_id"`
	BuildVersion   string        `json:"build_version"`
	Status         string        `json:"status"`
	StatusCode     int           `json:"status_code"`
	Time           string        `json:"time"`
}

// CustomTime is a wrapper around time.Time that handles time strings without timezone.
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface for CustomTime.
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	s := strings.Trim(string(data), "\"")

	// If empty, return without error
	if s == "" || s == "null" {
		return nil
	}

	// Try standard RFC3339 format first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Try without timezone - assume local timezone
	// Common formats from RIPEstat API: "2024-05-20T18:05:00" or "2023-07-03 15:49:57"
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ct.Time = t
			return nil
		}
	}

	// Return the last error if all formats fail
	return err
}
