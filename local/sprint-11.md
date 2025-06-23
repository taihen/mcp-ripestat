Implement RIPEStat Endpoint What’s‑My‑IP  ☞ /data/whats-my-ip/data.json

Read the local/requirements.md for general guidance for implementation.

### Capability Check
• Simple GET returns caller’s public IP.

### Route Design
Provide a direct proxy:
    GET /whats-my-ip
No parameters; return upstream payload as‑is inside the standard envelope.

### Implementation Notes
• Respect `X-Forwarded-For` when MCP is behind a proxy.
• Add flag to the mcp-server to allow disabling this feature (enabled by default) if run as a server and shared between team members