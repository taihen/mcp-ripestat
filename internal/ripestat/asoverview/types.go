package asoverview

// Response is the top-level structure for the RIPEstat AS Overview API response.
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

// Data represents the core data of the AS Overview response.
type Data struct {
	Type           string `json:"type"`
	Resource       string `json:"resource"`
	Block          Block  `json:"block"`
	Holder         string `json:"holder"`
	Announced      bool   `json:"announced"`
	QueryStartTime string `json:"query_starttime"`
	QueryEndTime   string `json:"query_endtime"`
}

// Block contains information about the AS block.
type Block struct {
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
}
