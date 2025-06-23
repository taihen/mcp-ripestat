Implement RIPEStat Endpoint Whois  ☞ /data/whois/data.json

Read the local/requirements.md for general guidance for implementation.

### Capability Check
• Request example:  `GET /data/whois/data.json?resource=AS3333`  → ensure `records` present.

### Route Design
Design a route that exposes Whois data for any IP, prefix, or ASN.  Suggested shape (discover and confirm):
    GET /whois?resource=<value>
Return the RIPEstat `records` list inside our standard response envelope `{ data, cached, fetched_at }`.

### Implementation Notes
• Validate the `resource` parameter (ASN or IPv4/IPv6, single or range).
• Cache successful answers for ~5 minutes.
• Translate upstream errors into a 5xx response that includes the RIPEstat status message.