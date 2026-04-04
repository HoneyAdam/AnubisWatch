# Changelog

All notable changes to AnubisWatch will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2026-04-04

### Added
- Initial release of AnubisWatch
- 10 protocol checkers: HTTP/HTTPS, TCP, UDP, DNS, ICMP, SMTP, IMAP, gRPC, WebSocket, TLS
- Embedded B+Tree storage (CobaltDB) with WAL and MVCC
- Raft consensus for distributed clustering
- Probe engine with adaptive intervals
- Alert engine with compound conditions and rate limiting
- REST API, WebSocket, and gRPC interfaces
- React 19 + Tailwind 4.1 dashboard
- Single binary deployment with zero dependencies
- Multi-tenancy support with workspaces and role-based access control
- Public status pages with custom domains, password protection, and uptime history
- 9 alert notification channels (Slack, Discord, Telegram, Email, PagerDuty, OpsGenie, SMS, Ntfy, Webhook)
- MCP (Model Context Protocol) server for AI integration
- Time-series storage with automatic downsampling
- Dashboard embedding in single binary (React 19 + Tailwind 4.1)
- Status page REST API endpoints
- Kubernetes Helm chart for deployment
- Docker and docker-compose support (single-node and 3-node cluster)
- Installation script for easy Linux/macOS setup
- systemd service file for production deployments
- Homebrew formula for macOS/Linux
- GHCR-exclusive container images (multi-arch: amd64, arm64, arm/v7)

### Documentation
- README.md with features, quick start, and comparison table
- BRANDING.md with complete brand guidelines
- CONFIGURATION.md with full configuration reference
- DEPLOYMENT.md with deployment guides
- GHCR.md with container registry documentation
- openapi.yaml with OpenAPI 3.1.0 specification
- WEBSITE.md with anubis.watch landing page content
- CONTRIBUTING.md with contribution guidelines
- INDEX.md with documentation index
- RELEASE_TEMPLATE.md for GitHub Releases
- MARKETING.md with launch materials

---

[0.0.1]: https://github.com/AnubisWatch/anubiswatch/releases/tag/v0.0.1
