# 🍳 Next-level investigations

Here’s a grab-bag of “next-level” questions you can throw at the
`mcp-ripestat` server. Feel free to contribute your own with a PR.

I’ve grouped them by investigation style and shown the workflow call(s) that will be issued under the hood for each prompt.

## BGP & RPKI threat hunting

| 🍳 Prompt | 🔧 Workflow |
| “For AS 20940 (Akamai) list every prefix it originated in the last 48 h that is RPKI-invalid and tell me which RIS collectors first saw the leak.” | • announced-prefixes returns the live prefix set for AS 20940 ￼ • Each prefix/ASN pair is piped into rpki-validation for status=invalid_asn/invalid_length ￼ • bgp-updates filtered to those prefixes + “A”nnouncements surfaces the first RRC/time seen |￼
| “Show me any /24s in 185.0.0.0/14 that went from ‘unknown’ to ‘valid’ RPKI state in the last week.” | • Sliding-window diff of rpki-history (counts of VRPs) • Compare status snapshots, emit changed prefixes |

**Why it’s fancy**:

You get instant leak hijack detection without writing BGP parsers,
plus provenance (which collector saw it first).

## Real-time outage triage

| 🍳 Prompt | 🔧 Workflow |
| “Is there a routing black-hole around IP 203.0.113.45 right now? Show which collectors still see a path and the last AS hop.” | looking-glass gives per-RRC visibility and full AS-PATHs ￼; the LLM groups peers by last-updated timestamp and highlights gaps. |
| “Compare the upstream set for AS 6453 today vs. 72 hours ago and highlight new or missing peers.” | Diff two asn-neighbours snapshots; render a before/after table. |

**Why it’s fancy**:

You’re effectively turning the RIS network into a distributed “ping” without touching a router.

## Abuse & takedown workflows (cross-dataset)

| 🍳 Prompt | 🔧 Workflow |
“Give me the abuse-mailbox for every prefix that belongs to the IPs hosting examplephish[.]com and tell me which of those IPs expose RDP.” 1. mcp-censys → lookup_domain to enumerate host IPs & ports ￼ 2. abuse-contact-finder for each IP/prefix ￼ 3. The LLM correlates and outputs a ready-to-mail list.
“Find all Shodan-indexed hosts inside AS 9808 that run OpenSSH < 8.2 and whose RPKI status is invalid.” 1. mcp-shodan → search query org:"AS9808" product:"OpenSSH" version:<8.2 ￼ 2. For each hit, call rpki-validation to check prefix/ASN combo; filter status != valid.

## Geo-policy & compliance checks

| 🍳 Prompt | 🔧 Workflow |
| “List every routed ASN registered in 🇷🇺 Russia and the countries where their prefixes are actually being announced from.” | country-asns (registered vs routed) + prefix-overview for geolocation per prefix |￼
| “Which ASNs that appear in OFAC-sanctioned countries are transiting traffic through EU IXPs?” | Combine previous query with public IX-prefix lists (or IX-API via another MCP server) and looking-glass visibility. |

## Historical forensics

| 🍳 Prompt | 🔧 Workflow |
| “When did AS 212238 first start announcing 2a0c:9a40::/29 and what other ASNs announced it beforehand?” | routing-history for that prefix; LLM finds earliest time & origin-change events. |
| “Plot the VRP count for 8.8.8.0/24 over the past year and annotate dips.” | rpki-history time-series; LLM (or a python_user_visible plot) highlights anomalies. |

## Putting it all together in one sentence

> “Over the last 24 h, which prefixes newly originated by AS 61138 are invalid in RPKI, have at least one open Telnet port according to Shodan, and lack an abuse mailbox?”

> “Give me a timeline of BGP withdrawals for 2400:cb00::/32 during Cloudflare’s Oct-2024 outage and overlay it with the count of probes failing HTTPS from RIPE Atlas.”

The LLM will federate: 1. RIPEstat (mcp-ripestat) for route, RPKI, Whois, visibility. 2. Shodan (mcp-shodan) for service & vuln intel. 3. Censys (mcp-censys) or any other MCP OSINT source for certificates/DNS. 4. Optionally, an Atlas or Pingdom MCP server for active-measurements.

## Tip: Hint the tool names

If a client supports explicit tool selection, prefix can be specified
to the prompt:

```sh
@ripestat announced_prefixes AS61138 starttime=2025-06-24T00:00Z
```

…but 90 % of the time you can stay high-level and just say
“Show/Get me...” — the LLM will decide which function to invoke.
