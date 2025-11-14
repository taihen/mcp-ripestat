package asoverview

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	Type           string `json:"type"`
	Resource       string `json:"resource"`
	Block          Block  `json:"block"`
	Holder         string `json:"holder"`
	Announced      bool   `json:"announced"`
	QueryStartTime string `json:"query_starttime"`
	QueryEndTime   string `json:"query_endtime"`
}

type Block struct {
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
}
