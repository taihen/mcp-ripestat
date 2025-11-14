package rpkivalidation

import (
	"github.com/taihen/mcp-ripestat/internal/ripestat/types"
)

type Response struct {
	types.BaseResponse
	Data Data `json:"data"`
}

type Data struct {
	ValidatingROAs []ValidatingROA `json:"validating_roas"`
	Status         string          `json:"status"`
	Validator      string          `json:"validator"`
	Resource       string          `json:"resource"`
	Prefix         string          `json:"prefix"`
}

type ValidatingROA struct {
	Origin    string `json:"origin"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"`
	Validity  string `json:"validity"`
}

type APIResponse struct {
	Status         string          `json:"status"`
	Validator      string          `json:"validator"`
	Resource       string          `json:"resource"`
	Prefix         string          `json:"prefix"`
	ValidatingROAs []ValidatingROA `json:"validating_roas"`
	FetchedAt      string          `json:"fetched_at"`
}
