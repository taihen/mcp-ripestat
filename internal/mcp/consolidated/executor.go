package consolidated

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/taihen/mcp-ripestat/internal/ripestat/abusecontactfinder"
	"github.com/taihen/mcp-ripestat/internal/ripestat/addressspacehierarchy"
	"github.com/taihen/mcp-ripestat/internal/ripestat/allocationhistory"
	"github.com/taihen/mcp-ripestat/internal/ripestat/announcedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asnneighbours"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asoverview"
	"github.com/taihen/mcp-ripestat/internal/ripestat/aspathlength"
	"github.com/taihen/mcp-ripestat/internal/ripestat/asroutingconsistency"
	"github.com/taihen/mcp-ripestat/internal/ripestat/bgplay"
	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpstate"
	"github.com/taihen/mcp-ripestat/internal/ripestat/bgpupdates"
	"github.com/taihen/mcp-ripestat/internal/ripestat/countryasns"
	"github.com/taihen/mcp-ripestat/internal/ripestat/lookingglass"
	"github.com/taihen/mcp-ripestat/internal/ripestat/networkinfo"
	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixoverview"
	"github.com/taihen/mcp-ripestat/internal/ripestat/prefixroutingconsistency"
	"github.com/taihen/mcp-ripestat/internal/ripestat/relatedprefixes"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routinghistory"
	"github.com/taihen/mcp-ripestat/internal/ripestat/routingstatus"
	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkihistory"
	"github.com/taihen/mcp-ripestat/internal/ripestat/rpkivalidation"
	"github.com/taihen/mcp-ripestat/internal/ripestat/whois"
)

type DirectExecutor struct{}

func NewDirectExecutor() *DirectExecutor {
	return &DirectExecutor{}
}

func (de *DirectExecutor) ExecuteEndpoint(ctx context.Context, endpoint string, resource string, params map[string]interface{}) (interface{}, error) {
	switch endpoint {
	case "getNetworkInfo":
		return networkinfo.GetNetworkInfo(ctx, resource)
	case "getASOverview":
		return asoverview.GetASOverview(ctx, resource)
	case "getAnnouncedPrefixes":
		return announcedprefixes.GetAnnouncedPrefixes(ctx, resource)
	case "getRelatedPrefixes":
		return relatedprefixes.GetRelatedPrefixes(ctx, resource)
	case "getRoutingStatus":
		return routingstatus.GetRoutingStatus(ctx, resource)
	case "getRoutingHistory":
		return de.handleRoutingHistory(ctx, resource, params)
	case "getWhois":
		return whois.GetWhois(ctx, resource)
	case "getAbuseContactFinder":
		return abusecontactfinder.GetAbuseContactFinder(ctx, resource)
	case "getRPKIValidation":
		return de.handleRPKIValidation(ctx, resource, params)
	case "getRPKIHistory":
		return rpkihistory.GetRPKIHistory(ctx, resource)
	case "getASNNeighbours":
		return de.handleASNNeighbours(ctx, resource, params)
	case "getLookingGlass":
		return de.handleLookingGlass(ctx, resource, params)
	case "getCountryASNs":
		return de.handleCountryASNs(ctx, resource, params)
	case "getBGPlay":
		return bgplay.GetBGPlay(ctx, resource)
	case "getBGPUpdates":
		return de.handleBGPUpdates(ctx, resource, params)
	case "getBGPState":
		return de.handleBGPState(ctx, resource, params)
	case "getPrefixRoutingConsistency":
		return prefixroutingconsistency.GetPrefixRoutingConsistency(ctx, resource)
	case "getPrefixOverview":
		return prefixoverview.GetPrefixOverview(ctx, resource)
	case "getAddressSpaceHierarchy":
		return addressspacehierarchy.GetAddressSpaceHierarchy(ctx, resource)
	case "getAllocationHistory":
		return allocationhistory.GetAllocationHistory(ctx, resource)
	case "getASPathLength":
		return aspathlength.GetASPathLength(ctx, resource)
	case "getASRoutingConsistency":
		return asroutingconsistency.GetASRoutingConsistency(ctx, resource)
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

func (de *DirectExecutor) handleBGPUpdates(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	opts, err := bgpUpdateOptions(params, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if opts == nil {
		return bgpupdates.GetBGPUpdates(ctx, resource)
	}
	return bgpupdates.DefaultClient().GetWithOptions(ctx, resource, opts)
}

func bgpUpdateOptions(params map[string]interface{}, now time.Time) (*bgpupdates.GetOptions, error) {
	timeframe := getOptionalStringParam(params, "timeframe")
	if timeframe == "" {
		return nil, nil
	}
	startTime, endTime, err := timeframeBounds(timeframe, now)
	if err != nil {
		return nil, err
	}
	return &bgpupdates.GetOptions{
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   endTime.Format(time.RFC3339),
	}, nil
}

func timeframeBounds(timeframe string, now time.Time) (time.Time, time.Time, error) {
	var duration time.Duration
	switch timeframe {
	case TimeframeCurrent:
		duration = time.Hour
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported timeframe value: %s", timeframe)
	}

	endTime := now.UTC()
	return endTime.Add(-duration), endTime, nil
}

func (de *DirectExecutor) handleRoutingHistory(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	startTime := getOptionalStringParam(params, "start_time")
	endTime := getOptionalStringParam(params, "end_time")
	maxResults := getOptionalIntParam(params, "max_results")

	if startTime != "" || endTime != "" || maxResults > 0 {
		return routinghistory.GetRoutingHistoryWithOptions(ctx, resource, startTime, endTime, maxResults)
	}
	return routinghistory.GetRoutingHistory(ctx, resource)
}

func (de *DirectExecutor) handleRPKIValidation(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	prefix := getOptionalStringParam(params, "prefix")
	if prefix == "" {
		return nil, fmt.Errorf("prefix parameter is required for RPKI validation")
	}
	return rpkivalidation.GetRPKIValidation(ctx, resource, prefix)
}

func (de *DirectExecutor) handleASNNeighbours(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lod := getOptionalIntParam(params, "lod")
	queryTime := getOptionalStringParam(params, "query_time")
	return asnneighbours.GetASNNeighbours(ctx, resource, lod, queryTime)
}

func (de *DirectExecutor) handleLookingGlass(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lookBackLimit := getOptionalIntParam(params, "look_back_limit")
	return lookingglass.GetLookingGlass(ctx, resource, lookBackLimit)
}

func (de *DirectExecutor) handleCountryASNs(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lod := getOptionalIntParam(params, "lod")
	opts := &countryasns.GetOptions{LOD: lod}
	return countryasns.GetCountryASNs(ctx, resource, opts)
}

func (de *DirectExecutor) handleBGPState(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	opts := bgpstate.Options{Resource: resource}

	if timestamp := getOptionalStringParam(params, "timestamp"); timestamp != "" {
		opts.Timestamp = timestamp
	}
	if rrcs := getOptionalStringParam(params, "rrcs"); rrcs != "" {
		opts.RRCs = rrcs
	}
	if unixTimestamps, ok := params["unix_timestamps"].(bool); ok {
		opts.UnixTimestamps = unixTimestamps
	}

	client := bgpstate.DefaultClient()
	return client.Get(ctx, opts)
}

func getOptionalStringParam(params map[string]interface{}, key string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return ""
}

func getOptionalIntParam(params map[string]interface{}, key string) int {
	value, ok := params[key]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}
