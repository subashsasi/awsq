# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

Only the latest release receives security updates. We recommend always using the most recent version.

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email: subashsasi@gmail.com (or use GitHub's private vulnerability reporting)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Assessment**: Within 7 days
- **Fix release**: Within 14 days for critical issues

## Scope

Since `awsq` is a read-only CLI tool that runs locally, the primary security concerns are:

- Dependency vulnerabilities (AWS SDK, Cobra)
- Credential handling (ensuring AWS credentials are never logged or exposed)
- Supply chain integrity (binary checksums, module verification)

## Security Measures in Place

- `govulncheck` runs in CI to catch known dependency vulnerabilities
- `go mod verify` ensures dependency integrity
- Dependabot monitors for vulnerable dependencies weekly
- No secrets or credentials are ever logged or stored by the tool
- All AWS API communication uses TLS (enforced by the SDK)
