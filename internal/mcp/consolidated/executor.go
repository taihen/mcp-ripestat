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
	"github.com/taihen/mcp-ripestat/internal/ripestat/client"
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

type DirectExecutor struct {
	ripe *client.Client
}

func NewDirectExecutor(ripe *client.Client) *DirectExecutor {
	if ripe == nil {
		panic("mcp: ripe client is nil")
	}
	return &DirectExecutor{ripe: ripe}
}

func (de *DirectExecutor) RIPEClient() *client.Client {
	return de.ripe
}

func (de *DirectExecutor) ExecuteEndpoint(ctx context.Context, endpoint string, resource string, params map[string]interface{}) (interface{}, error) {
	if de == nil || de.ripe == nil {
		panic("mcp: ripe client is nil")
	}
	switch endpoint {
	case "getNetworkInfo":
		return networkinfo.NewClient(de.ripe).Get(ctx, resource)
	case "getASOverview":
		return asoverview.NewClient(de.ripe).Get(ctx, resource)
	case "getAnnouncedPrefixes":
		return announcedprefixes.NewClient(de.ripe).Get(ctx, resource)
	case "getRelatedPrefixes":
		return relatedprefixes.NewClient(de.ripe).Get(ctx, resource)
	case "getRoutingStatus":
		return routingstatus.NewClient(de.ripe).Get(ctx, resource)
	case "getRoutingHistory":
		return de.handleRoutingHistory(ctx, resource, params)
	case "getWhois":
		return whois.NewClient(de.ripe).Get(ctx, resource)
	case "getAbuseContactFinder":
		return abusecontactfinder.NewClient(de.ripe).Get(ctx, resource)
	case "getRPKIValidation":
		return de.handleRPKIValidation(ctx, resource, params)
	case "getRPKIHistory":
		return rpkihistory.NewClient(de.ripe).Get(ctx, resource)
	case "getASNNeighbours":
		return de.handleASNNeighbours(ctx, resource, params)
	case "getLookingGlass":
		return de.handleLookingGlass(ctx, resource, params)
	case "getCountryASNs":
		return de.handleCountryASNs(ctx, resource, params)
	case "getBGPlay":
		return bgplay.NewClient(de.ripe).Get(ctx, resource)
	case "getBGPUpdates":
		return de.handleBGPUpdates(ctx, resource, params)
	case "getBGPState":
		return de.handleBGPState(ctx, resource, params)
	case "getPrefixRoutingConsistency":
		return prefixroutingconsistency.NewClient(de.ripe).Get(ctx, resource)
	case "getPrefixOverview":
		return prefixoverview.NewClient(de.ripe).Get(ctx, resource)
	case "getAddressSpaceHierarchy":
		return addressspacehierarchy.NewClient(de.ripe).Get(ctx, resource)
	case "getAllocationHistory":
		return allocationhistory.NewClient(de.ripe).Get(ctx, resource)
	case "getASPathLength":
		return aspathlength.NewClient(de.ripe).Get(ctx, resource)
	case "getASRoutingConsistency":
		return asroutingconsistency.NewClient(de.ripe).Get(ctx, resource)
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

func (de *DirectExecutor) handleBGPUpdates(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	opts, err := bgpUpdateOptions(params, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	endpoint := bgpupdates.NewClient(de.ripe)
	if opts == nil {
		return endpoint.Get(ctx, resource)
	}
	return endpoint.GetWithOptions(ctx, resource, opts)
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

	endpoint := routinghistory.NewClient(de.ripe)
	if startTime != "" || endTime != "" || maxResults > 0 {
		return endpoint.GetWithOptions(ctx, resource, startTime, endTime, maxResults)
	}
	return endpoint.Get(ctx, resource)
}

func (de *DirectExecutor) handleRPKIValidation(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	prefix := getOptionalStringParam(params, "prefix")
	if prefix == "" {
		return nil, fmt.Errorf("prefix parameter is required for RPKI validation")
	}
	return rpkivalidation.NewClient(de.ripe).Get(ctx, resource, prefix)
}

func (de *DirectExecutor) handleASNNeighbours(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lod := getOptionalIntParam(params, "lod")
	queryTime := getOptionalStringParam(params, "query_time")
	return asnneighbours.NewClient(de.ripe).Get(ctx, resource, lod, queryTime)
}

func (de *DirectExecutor) handleLookingGlass(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lookBackLimit := getOptionalIntParam(params, "look_back_limit")
	return lookingglass.NewClient(de.ripe).Get(ctx, resource, lookBackLimit)
}

func (de *DirectExecutor) handleCountryASNs(ctx context.Context, resource string, params map[string]interface{}) (interface{}, error) {
	lod := getOptionalIntParam(params, "lod")
	opts := &countryasns.GetOptions{LOD: lod}
	return countryasns.NewClient(de.ripe).Get(ctx, resource, opts)
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

	return bgpstate.NewClient(de.ripe).Get(ctx, opts)
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
