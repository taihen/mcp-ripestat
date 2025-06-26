# 🌊 RIPEstat Investigation Flows

Below you will find example prompts for the `mcp-ripestat` server.

> [!NOTE]
> These prompts serve as both user examples and development requirements for
> `mcp-ripestat`. They help prioritize new features based on real-world
> investigation scenarios.

## 🛠 Available Tools

The current implementation provides these RIPEstat API endpoints:

**Core Network Analysis:**

- `getNetworkInfo` - IP/prefix ownership and registration details
- `getASOverview` - Autonomous System information and statistics
- `getWhois` - Registry information for IPs, prefixes, or ASNs

**Routing Intelligence:**

- `getAnnouncedPrefixes` - Prefixes currently announced by an AS
- `getRoutingStatus` - Current routing visibility for a prefix
- `getRoutingHistory` - Historical routing changes and announcements
- `getASNNeighbours` - Upstream/downstream AS relationships
- `getLookingGlass` - Real-time BGP data from RIPE RRCs

**Security & Compliance:**

- `getRPKIValidation` - RPKI validation status for ASN/prefix pairs
- `getAbuseContactFinder` - Abuse contact information for resources

**Geographic & Regional Analysis:**

- `getCountryASNs` - Autonomous System Numbers for a given country code

**Utility:**

- `getWhatsMyIP` - Caller's public IP address (supports proxy headers for accurate client IP detection)

---

## 📊 Basic Prompts by Input Type

### IP Address Queries

```shell
"What network information is available for 8.8.8.8?"
"Show me WHOIS data for 192.0.2.1"
"Find abuse contacts for IP 203.0.113.50"
"Get routing history for 1.1.1.1"
"What's the network ownership of 2001:db8::1?"
```

### IP Prefix Queries

```shell
"Analyze the prefix 193.0.0.0/21"
"What's the routing status for 8.8.8.0/24?"
"Show BGP data for 2001:7fb::/32 from RIPE looking glass"
"Get historical routing changes for 104.16.0.0/13"
"Find the network owner of 198.51.100.0/24"
```

### Autonomous System (ASN) Queries

```shell
"Give me an overview of AS3333"
"What prefixes does AS15169 announce?"
"Show me the neighbors of AS1205"
"Get routing history for AS64512"
"What's the WHOIS information for AS13335?"
```

### RPKI Validation Queries

```shell
"Validate RPKI for AS3333 announcing 193.0.0.0/21"
"Check if AS15169 is authorized for 8.8.8.0/24"
"Verify RPKI status of AS13335 and 104.16.0.0/13"
"Is AS64496 valid for announcing 203.0.113.0/24?"
```

### Country & Geographic Queries

```shell
"Show me all ASNs registered in the Netherlands"
"Get ASN statistics for Germany with detailed lists"
"How many ASNs does France have?"
"List all routed and non-routed ASNs for the United States"
"Compare ASN counts between Switzerland and Austria"
```

### Utility Queries

```shell
"What's my public IP address?"
"Detect my current IP and show its network information"
"Show me my IP and find its abuse contact"
"Get my real client IP address (bypassing proxies/load balancers)"
```

---

## 🔗 Advanced Chained Prompts

### Security Analysis Chains

```shell
"For AS 20940, show me all announced prefixes that fail RPKI validation,
then get the routing history for each invalid prefix to see when the
announcements first appeared."

"Investigate AS 64496: get an overview, list all announced prefixes,
check RPKI validation for each, and find abuse contacts for any
invalid routes."

“Show me any /24s in 185.0.0.0/14 that went from ‘unknown’ to ‘valid’ RPKI
state in the last week.”

"For the prefix 8.8.8.0/24, show current routing status, check if
Google (AS15169) is the only announcer in BGP history, and validate
RPKI authorization."
```

### Network Forensics Chains

```shell
"Analyze 203.0.113.0/24: get network ownership details, check current
routing status, review historical announcements, and find all abuse
contact information."

"For AS 174, compare current neighbors with those from 30 days ago,
show any new peering relationships, and get WHOIS details for any
new upstream or downstream partners."

“Compare the upstream set for AS 6453 today vs. 72 hours ago and highlight new
or missing peers.”

"Investigate routing instability for 192.0.2.0/24: show BGP visibility
across RIPE collectors, check for multiple origin announcements, and
verify RPKI status for all announcing ASNs."

“Is there a routing black-hole around IP 203.0.113.45 right now?
Show which collectors still see a path and the last AS hop.”

“Plot the VRP count for 8.8.8.0/24 over the past year and annotate dips.”

“When did AS 212238 first start announcing 2a0c:9a40::/29 and what other ASNs
announced it beforehand?”


```

### Compliance Verification Chains

```shell
"Audit AS 13335: list all announced prefixes, validate RPKI status
for each, check for recent routing changes, and identify any prefixes
announced in the last 7 days."

"For organization compliance check: get AS overview for AS3333,
verify all announced prefixes have valid RPKI, check abuse contact
availability, and ensure WHOIS data is current."
```

### Infrastructure Analysis Chains

```shell
"Compare Google DNS (8.8.8.8) and Cloudflare DNS (1.1.1.1): get
network information for both IPs, analyze their respective AS
relationships, compare BGP paths from RIPE collectors, and check
RPKI validation status."

"Trace the network path for cloudflare.com: get IP addresses,
identify owning prefixes and ASNs, show AS neighbor relationships,
and verify routing stability over the past week."
```

### Geopolitical Analysis Chains

```shell
“List every routed ASN registered in 🇷🇺 Russia and the countries where their
prefixes are actually being announced from.”

“Which ASNs that appear in OFAC-sanctioned countries are transiting traffic
through EU IXPs?"
```

---

## 🌐 Multi-MCP Server Integration

Combine `mcp-ripestat` with other MCP servers for comprehensive investigations:

### Security Research with mcp-shodan

```shell
"Find all hosts in AS 64496 running SSH services (via Shodan),
then use RIPEstat to check RPKI validation status for their
announced prefixes and get abuse contact information for
responsible disclosure."

"Identify vulnerable HTTP servers in the 203.0.113.0/24 range
(via Shodan), then use RIPEstat to verify network ownership,
check routing legitimacy, and find appropriate abuse contacts."
```

---

## 💡 Tips and Hints

### Tool selection

- If your client supports explicit tool selection, prefix the prompt.

```shell
@ripestat announced_prefixes AS61138 starttime=2025-06-24T00:00Z
```

### Query Optimization

- **Be specific with timeframes**: Use "last 24 hours", "past week", or specific dates for historical queries
- **Combine related tools**: Chain network info → WHOIS → abuse contacts for complete investigations
- **Use AS numbers and names**: "AS3333" and "RIPE NCC" both work for queries
- **Leverage LOD parameters**: Request detailed neighbor information with "detailed view" or "LOD 1"

### Natural Language Patterns

- **Start broad, then narrow**: "Overview of AS15169" followed by "RPKI status for their prefixes"
- **Use comparison language**: "Compare", "versus", "differences between" for temporal analysis
- **Express investigation intent**: "Investigate", "analyze", "audit" for comprehensive checks
- **Request verification**: "Verify", "validate", "confirm" for compliance checks

### Advanced Techniques

- **Historical snapshots**: Compare current state with specific past dates using `query_time`
- **Regional analysis**: Request BGP data from specific RIPE RRC collectors
- **Bulk operations**: Process multiple prefixes or ASNs in a single investigation
- **Cross-validation**: Use multiple tools to verify the same information from different angles

### Integration Strategies

- **Start with RIPEstat**: Use network discovery as foundation for other MCP server queries
- **Chain efficiently**: Pass RIPEstat results (IP ranges, ASNs) as input to other servers
- **Correlate data**: Match timestamps between different data sources for event correlation
- **Automate workflows**: Create repeatable investigation patterns for common use cases
