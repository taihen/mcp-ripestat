package routingstatus

import (
	"strings"
	"time"
)

// CustomTime is a wrapper around time.Time that handles time strings without timezone
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface for CustomTime
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

// Response represents the routing-status API response.
type Response struct {
	Messages       []interface{} `json:"messages"`
	SeeAlso        []interface{} `json:"see_also"`
	Version        string        `json:"version"`
	DataCallName   string        `json:"data_call_name"`
	DataCallStatus string        `json:"data_call_status"`
	Cached         bool          `json:"cached"`
	Data           Data          `json:"data"`
	QueryID        string        `json:"query_id"`
	ProcessTime    int           `json:"process_time"`
	ServerID       string        `json:"server_id"`
	BuildVersion   string        `json:"build_version"`
	Status         string        `json:"status"`
	StatusCode     int           `json:"status_code"`
	Time           string        `json:"time"`
}

// Data is the main data structure for the routing-status API response.
type Data struct {
	FirstSeen     RouteInfo  `json:"first_seen"`
	LastSeen      RouteInfo  `json:"last_seen"`
	Visibility    Visibility `json:"visibility"`
	Origins       []Origin   `json:"origins"`
	LessSpecifics []any      `json:"less_specifics"`
	MoreSpecifics []any      `json:"more_specifics"`
	Resource      string     `json:"resource"`
	QueryTime     CustomTime `json:"query_time"`
}

// RouteInfo contains information about when a route was first or last seen.
type RouteInfo struct {
	Prefix string     `json:"prefix"`
	Origin string     `json:"origin"`
	Time   CustomTime `json:"time"`
}

// Visibility provides information about the visibility of a route in the routing system.
type Visibility struct {
	V4 AddressVisibility `json:"v4"`
	V6 AddressVisibility `json:"v6"`
}

// AddressVisibility contains visibility information for either IPv4 or IPv6.
type AddressVisibility struct {
	RISPeersSeeing int `json:"ris_peers_seeing"`
	TotalRISPeers  int `json:"total_ris_peers"`
}

// Origin represents an origin AS and associated route objects.
type Origin struct {
	Origin       int      `json:"origin"`
	RouteObjects []string `json:"route_objects"`
}
