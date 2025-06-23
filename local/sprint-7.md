Implement RIPEStat Endpoint Abuse‑Contact‑Finder  ☞ /data/abuse-contact-finder/data.json

Read the local/requirements.md for general guidance for implementation.

### Capability Check
• Example: 193.0.0.0/21 → `data.abuse_contacts`.

### Route Design
Expose abuse contacts for any resource.  Likely pattern:
    GET /abuse-contact-finder?resource=<value>
Return `{ contacts: [...], fetched_at }`.

### Implementation Notes
• Validate `resource`.
• Cache for ~1 hour.
• Empty contact list is a valid, 200‑OK answer.