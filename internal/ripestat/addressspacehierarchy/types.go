// Package addressspacehierarchy provides access to the RIPEstat address-space-hierarchy API.
package addressspacehierarchy

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

// Response is the top-level structure for the RIPEstat Address Space Hierarchy API response.
type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

// Data represents the core data of the Address Space Hierarchy response.
type Data struct {
	RIR          string         `json:"rir"`
	Resource     string         `json:"resource"`
	Exact        []AddressEntry `json:"exact"`
	LessSpecific []AddressEntry `json:"less_specific"`
	MoreSpecific []AddressEntry `json:"more_specific"`
	QueryTime    string         `json:"query_time"`
	Parameters   Parameters     `json:"parameters"`
}

// AddressEntry represents an address space entry in the hierarchy.
type AddressEntry struct {
	Inetnum      string `json:"inetnum"`
	Netname      string `json:"netname"`
	Descr        string `json:"descr,omitempty"`
	Org          string `json:"org,omitempty"`
	Remarks      string `json:"remarks,omitempty"`
	Country      string `json:"country,omitempty"`
	AdminC       string `json:"admin-c,omitempty"`
	TechC        string `json:"tech-c,omitempty"`
	Status       string `json:"status,omitempty"`
	MntBy        string `json:"mnt-by,omitempty"`
	MntRoutes    string `json:"mnt-routes,omitempty"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"last-modified,omitempty"`
	Source       string `json:"source,omitempty"`
}

// Parameters contains the query parameters used for the request.
type Parameters struct {
	Resource string      `json:"resource"`
	Cache    interface{} `json:"cache"`
}
