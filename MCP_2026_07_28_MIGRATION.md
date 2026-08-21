# MCP 2026-07-28 migration report

## Scope

This migration updates `mcp-ripestat` from its `2025-06-18`-oriented
Streamable HTTP implementation to the MCP `2026-07-28` stateless protocol.
The production endpoint continues to use the official Go SDK and remains
modern-only by default. Initialization-based compatibility remains available as
an explicit deployment opt-in through the SDK's dual-era fallback.

Primary references:

- [Release announcement](https://blog.modelcontextprotocol.io/posts/2026-07-28/)
- [Specification changelog](https://modelcontextprotocol.io/specification/2026-07-28/changelog)
- [Versioning and compatibility](https://modelcontextprotocol.io/specification/2026-07-28/basic/versioning)
- [Streamable HTTP](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/streamable-http)
- [Discovery](https://modelcontextprotocol.io/specification/2026-07-28/server/discover)
- [Caching](https://modelcontextprotocol.io/specification/2026-07-28/server/utilities/caching)
- [Tools](https://modelcontextprotocol.io/specification/2026-07-28/server/tools)

## Previous implementation

The repository contained two protocol layers:

1. The production HTTP server used `github.com/modelcontextprotocol/go-sdk`
   `v1.6.1`. A local middleware admitted only `2025-06-18`, defaulted a missing
   version header to that version, exposed GET endpoint information on `/mcp`,
   and advertised GET/DELETE/session headers through CORS.
2. `internal/mcp/server.go` provided a hand-written, pre-2026 JSON-RPC
   implementation used only by unit tests. It gated tool operations on
   `initialize` state and implemented `ping`; repository search found no
   production caller.

The runtime SDK already handled deterministic feature ordering, origin checks,
request-body limits, legacy negotiation, and JSON-RPC framing. The application
wrapper's single-version gate prevented the SDK from negotiating any other
revision.

## Change map

| Specification change | Applicability | Implementation |
| --- | --- | --- |
| Stateless requests; no modern initialize/session | Required | SDK `v1.7.0`; `StreamableHTTPOptions.Stateless = true`. |
| Per-request protocol version and client capabilities | Required | SDK validates typed modern `_meta` on every request. |
| Server identity in each result | Recommended | SDK `v1.7.0` adds `io.modelcontextprotocol/serverInfo` automatically. |
| `server/discover` | Required | SDK registers it automatically and advertises supported revisions. |
| `resultType` on every result | Required | SDK `v1.7.0` serializes `complete`. No RIPEstat tool currently needs MRTR. |
| `Mcp-Method` and `Mcp-Name` validation | Required over HTTP | SDK `v1.7.0` validates headers against the body and returns `HeaderMismatch` (`-32020`). Wire tests cover a mismatched tool name. |
| Cache hints | Required for discovery and tool lists | SDK response middleware sets a one-hour public TTL. Public scope is safe because the catalog is static and caller-independent. |
| Deterministic tool order | Recommended | The SDK's feature set sorts by tool name; wire tests assert sorted runtime output. |
| GET stream, DELETE session termination, SSE resumption removed | Required | `/mcp` delegates these methods to stateless SDK behavior (`405`, `Allow: POST`); the local GET information response was removed. |
| Request cancellation through response-stream closure | Required | `PropagateRequestCancellation` is enabled so disconnected modern HTTP requests cancel tool contexts. |
| `subscriptions/listen` | Not currently used | The catalog is static and the server does not emit list/resource change notifications. SDK support remains available if dynamic features are added. |
| MRTR | Not currently used | Tools are read-only RIPEstat queries and do not initiate sampling, roots, or elicitation requests. SDK support is available for future input-required results. |
| Roots, sampling, and logging deprecated | Applicable to capabilities | The production server now starts with empty explicit capabilities, avoiding the SDK's historical default logging capability. Tools are inferred when registered. |
| Tasks moved to an extension | Not used | No long-running task API is exposed; the tasks extension is not advertised. |
| OAuth issuer/DCR changes | Not implemented here | The server has no OAuth client or authorization-server implementation. Authentication can be added by a deployment gateway or future server middleware without changing tool semantics. |
| `x-mcp-header` | Optional and not used | No tool schema declares it. Standard routing headers are still required and validated. |

The hand-written server was deliberately kept as a deprecated legacy test
fixture instead of becoming a second owner of the 2026 protocol. It retains
initialization, `ping`, and legacy result shapes, and mirrors the SDK's legacy
version negotiation. Production traffic is owned exclusively by the official
SDK server.

## Compatibility behavior

- Modern `2026-07-28` HTTP calls are independent POST requests and never receive
  `Mcp-Session-Id`.
- Legacy compatibility is disabled by default. Setting
  `MCP_ENABLE_LEGACY_PROTOCOLS=true` enables real initialize → initialized →
  tools/list/tools/call sequences for the Streamable HTTP revisions
  `2025-11-25`, `2025-06-18`, and `2025-03-26`.
- Legacy HTTP requests are served through the SDK's compatibility model without
  restoring the removed GET stream or protocol session on the modern endpoint.
- Explicit unsupported or disabled versions return JSON-RPC code `-32022`;
  a headerless request on the default modern-only endpoint returns header
  mismatch code `-32020`. Both responses preserve the JSON-RPC request ID.

Legacy requests lack modern routing headers. A gateway that enables them must
authenticate independently and inspect the JSON-RPC body for method-level
policy; it must not rely on `Mcp-Method` or `Mcp-Name` being present. The
stateless compatibility shim does not make initialize an authorization gate,
and headerless non-initialize requests are treated as `2025-03-26`, which
predates the protocol-version header.

## Security hardening

- Origin validation now precedes preflight and protocol processing; invalid
  origins receive `403`, preflights do not consume rate-limit tokens, and CORS
  responses vary on origin/request headers.
- Browser access defaults to loopback origins. Third-party origins require
  `CORS_ALLOWED_ORIGINS`.
- Forwarding headers are ignored unless the direct peer is in an explicitly
  configured `TRUSTED_PROXIES` CIDR. X-Forwarded-For is evaluated right-to-left
  and malformed chains fail closed.
- Per-client rate-limit storage is bounded; overflow identities share one
  bucket. HTTP header, body-read, and idle timeouts limit slow-connection use.

## Verification

Protocol tests cover:

- stateless discovery without initialization;
- absence of a session ID;
- required result types, server identity, cache scope, and TTL;
- deterministic tool ordering;
- independent tool calls;
- header/body mismatch rejection (`-32020`);
- unsupported version negotiation (`-32022`);
- removal of modern `ping`;
- required per-request client capability metadata; and
- real tagged end-to-end flows for each claimed legacy Streamable HTTP revision,
  using raw legacy bodies without modern metadata or routing headers.

The final review findings and verification commands are reported with the
implementation handoff.
