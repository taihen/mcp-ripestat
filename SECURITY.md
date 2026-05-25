# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in mcp-ripestat, please report it responsibly.

### How to Report

Please report security vulnerabilities by creating a private security advisory on GitHub:

1. Go to the [Security tab](https://github.com/taihen/mcp-ripestat/security/advisories)
2. Click "Report a vulnerability"
3. Provide detailed information about the vulnerability

Alternatively, you can email security concerns to the maintainer.

### What to Include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if applicable)

## Supported Versions

We support the following versions of mcp-ripestat (three last minor releases):

| Version | Supported |
| ------- | --------- |
| 2.10.x  | ✅        |
| 2.9.x   | ✅        |
| 2.8.x   | ✅        |
| < 2.8   | ❌        |

## Security Considerations

**Important:** At current stage this MCP server does not provide authentication.
Use firewall or other L3 networking to restrict access to the server.

### Network Security

- Deploy behind a firewall
- Use network-level access controls
- Consider VPN or private network deployment
- Monitor access logs

### Dependencies and Supply Chain

- Go dependencies are regularly updated
- Automated vulnerability scanning with govulncheck
- Container images scanned with Trivy
- Release container images are published to `ghcr.io/taihen/mcp-ripestat`
- Container builds generate BuildKit provenance and SBOM attestations
- Release images receive signed GitHub Artifact Attestations for SLSA build provenance

Verify a release image provenance attestation with:

```bash
gh attestation verify oci://ghcr.io/taihen/mcp-ripestat:<version> \
  --repo taihen/mcp-ripestat
```

The release workflow targets SLSA Build Level 2 for container images. This means
published images are built by GitHub Actions and have signed provenance bound to
the image digest. Higher SLSA levels require additional repository and workflow
hardening, such as protected release tags, stricter workflow review, and a
hardened build platform.

### Development Security

- Code is scanned with CodeQL
- Linting includes security checks
- Branch protection rules

## Responsible Disclosure

We follow responsible disclosure practices:

- We will acknowledge receipt of vulnerability reports within 72 hours
- We will provide regular updates on the status of fixes
- We will credit reporters appropriately (unless they prefer to remain anonymous)
- We aim to fix critical vulnerabilities within 30 days
- We will coordinate with reporters on disclosure timing
