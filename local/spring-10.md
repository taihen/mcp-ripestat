Implement RIPEStat Endpoint Looking Glass  ☞ /data/looking-glass/data.json

Read the local/requirements.md for general guidance for implementation.

### Capability Check
• Example: 140.78.0.0/16, look_back_limit=3600 → list of RRCs.

### Route Design
Surface near‑real‑time BGP observations for a resource.  Tentative layout:
    GET /looking-glass?resource=<value>&look_back_limit=<secs>
(Enforce a reasonable upper bound on `look_back_limit`, e.g. 48 hours.)

### Implementation Notes
• Streaming or chunked JSON may be required for large responses.
• No caching.