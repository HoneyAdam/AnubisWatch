# AnubisWatch v0.0.1 — Aaru (Initial Release)

**Release Date:** April 4, 2026

## ⚖️ The Judgment Never Sleeps

We're thrilled to announce AnubisWatch v0.0.1, codename "Aaru" (paradise in Egyptian mythology).

This is the initial beta release of AnubisWatch. Built from the ground up with zero external dependencies, AnubisWatch judges your uptime with the precision of a god.

### 📦 Installation

```bash
# Homebrew (macOS/Linux)
brew install AnubisWatch/anubiswatch/anubiswatch

# Linux (curl-pipe-sh)
curl -fsSL https://anubis.watch/install.sh | sh

# Docker
docker pull ghcr.io/anubiswatch/anubiswatch:0.0.1

# Kubernetes Helm
helm repo add anubiswatch https://anubiswatch.github.io/helm-charts
helm install anubis anubiswatch/anubiswatch --version 0.0.1
```

---

## 🆕 What's New

### Eight Protocol Support

HTTP, TCP, UDP, DNS, SMTP, ICMP, gRPC, WebSocket, and TLS. If your infrastructure speaks it, we can judge it.

```yaml
souls:
  - name: "Production API"
    type: http
    target: "https://api.example.com/health"
    weight: 30s
    http:
      valid_status: [200]
      body_contains: '"status":"ok"'
  
  - name: "PostgreSQL"
    type: tcp
    target: "db.example.com:5432"
    tcp:
      banner_match: "PostgreSQL"
  
  - name: "DNS Propagation"
    type: dns
    target: "example.com"
    dns:
      record_type: A
      propagation_check: true
```

### Embedded React Dashboard

Beautiful React 19 + Tailwind 4.1 dashboard compiled into the binary. No separate frontend deployment. No `npm install`. No CORS nightmares.

- **Hall of Ma'at** — Global overview with uptime heatmap
- **Soul Detail** — Response time charts, EKG animation, TLS countdown
- **Custom Dashboards** — Grafana-style drag-and-drop widgets
- **Dark/Light Themes** — Tomb Interior and Desert Sun

### Raft Consensus Clustering

Built-in distributed clustering with automatic leader election. Every node is both a probe and a controller.

```bash
# Start cluster node
anubis serve --cluster

# Add node to cluster
anubis summon raft://node2.example.com:7946

# View cluster status
anubis necropolis status
```

### CobaltDB Storage

Embedded B+Tree database with:
- WAL (Write-Ahead Logging)
- MVCC (Multi-Version Concurrency Control)
- Encryption at rest
- Automatic time-series compaction
- Configurable retention policies

### Nine Alert Channels

Slack, Discord, Telegram, Email, PagerDuty, OpsGenie, SMS, Ntfy, and generic webhooks.

```yaml
channels:
  - name: "ops-slack"
    type: slack
    slack:
      webhook_url: "${SLACK_WEBHOOK_URL}"
      mention_on_critical: ["@oncall-team"]

  - name: "pagerduty"
    type: pagerduty
    pagerduty:
      integration_key: "${PAGERDUTY_KEY}"
      auto_resolve: true
```

### Public Status Pages

Custom domains, password protection, 90-day uptime history. Share your health with the world.

- Custom domain support (with ACME/Let's Encrypt)
- Password protection (public, protected, private)
- Uptime history (90-day GitHub-style heatmap)
- Embeddable badges/widgets

### Multi-Tenant Workspaces

SaaS-ready workspace isolation with RBAC:
- Admin, Editor, Viewer roles
- Per-workspace quotas
- Cross-workspace query isolation

### MCP Server

AI agent integration out of the box. Ask Claude, Cursor, or any MCP client:

```
"What's the status of production?"
→ "✅ 47/50 souls alive. ❌ 3 dead: API, DB, Cache."
```

---

## 📊 Key Statistics

| Metric | Value |
|--------|-------|
| Binary Size | ~24 MB |
| RAM Usage (idle) | < 64 MB |
| Protocols | 8 |
| Alert Channels | 9+ |
| Test Coverage | 80%+ |
| Architectures | amd64, arm64, arm/v7 |

---

## 📦 Binary Downloads

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | `anubiswatch_0.0.1_linux_amd64.tar.gz` |
| Linux | arm64 | `anubiswatch_0.0.1_linux_arm64.tar.gz` |
| Linux | arm/v7 | `anubiswatch_0.0.1_linux_armv7.tar.gz` |
| macOS | amd64 | `anubiswatch_0.0.1_darwin_amd64.tar.gz` |
| macOS | arm64 | `anubiswatch_0.0.1_darwin_arm64.tar.gz` |
| Windows | amd64 | `anubiswatch_0.0.1_windows_amd64.zip` |

---

## 📝 Verification

### Verify Docker Image

```bash
docker pull ghcr.io/anubiswatch/anubiswatch:0.0.1
docker inspect ghcr.io/anubiswatch/anubiswatch:0.0.1 | grep Digest
```

### Verify Homebrew

```bash
brew install AnubisWatch/anubiswatch/anubiswatch
anubis version
# AnubisWatch v0.0.1 (Aaru)
```

### Verify Binary

```bash
curl -LO https://github.com/AnubisWatch/anubiswatch/releases/download/v0.0.1/anubiswatch_0.0.1_linux_amd64.tar.gz
tar xzf anubiswatch_0.0.1_linux_amd64.tar.gz
./anubis version
```

---

## ⚠️ Breaking Changes

**None** — This is the initial release. No breaking changes.

---

## 🙏 Thank You

This release represents hundreds of hours of work. Thank you to everyone who contributed, tested, and provided feedback.

---

## 📚 Documentation

- [Quick Start](https://github.com/AnubisWatch/anubiswatch#quick-start)
- [Configuration Reference](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/CONFIGURATION.md)
- [Deployment Guide](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/DEPLOYMENT.md)
- [API Reference](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/openapi.yaml)

---

## 🔮 What's Next (v1.1.0)

- [ ] ACME/Let's Encrypt auto-certificate provisioning
- [ ] OIDC authentication (Google, GitHub, Okta)
- [ ] LDAP/Active Directory integration
- [ ] Full DNSSEC validation
- [ ] Anomaly detection for metrics
- [ ] PWA support for mobile dashboard

---

### Full Changelog

https://github.com/AnubisWatch/anubiswatch/compare/v0.0.1...HEAD

---

**⚖️ The Judgment Never Sleeps**

Built by ECOSTACK TECHNOLOGY OÜ  
Apache 2.0 License
