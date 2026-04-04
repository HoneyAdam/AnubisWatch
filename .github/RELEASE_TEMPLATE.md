# GitHub Release Template for AnubisWatch

Save this as `.github/RELEASE_TEMPLATE.md` or use when creating releases.

---

## Release Title
**AnubisWatch v{{VERSION}} — {{CODENAME}}**

Suggested codenames (Egyptian mythology theme):
- `Aaru` (paradise)
- `Duat` (underworld)
- `Ma'at` (truth/order)
- `Anubis` (jackal-headed god)
- `Osiris` (god of the dead)
- `Ra` (sun god)
- `Horus` (sky god)
- `Thoth` (god of wisdom)

---

## Release Notes

### ⚖️ The Judgment Never Sleeps

AnubisWatch v{{VERSION}} is now available. {{BRIEF_DESCRIPTION}}

**📦 Installation:**

```bash
# Homebrew (macOS/Linux)
brew install AnubisWatch/anubiswatch/anubiswatch

# Linux (curl-pipe-sh)
curl -fsSL https://anubis.watch/install.sh | sh

# Docker
docker pull ghcr.io/anubiswatch/anubiswatch:{{VERSION}}

# Kubernetes Helm
helm repo add anubiswatch https://anubiswatch.github.io/helm-charts
helm install anubis anubiswatch/anubiswatch --version {{VERSION}}
```

---

### 🆕 What's New

#### Feature 1: Title Here
Description of the feature. What problem does it solve? How does it work?

```yaml
# Example configuration
example_config: here
```

#### Feature 2: Title Here
Another feature description.

#### Feature 3: Title Here
Yet another feature.

---

### 📊 Key Improvements

| Category | Improvement |
|----------|-------------|
| **Performance** | X% faster startup, Y% lower memory |
| **Reliability** | Z bug fixes, improved error handling |
| **Security** | New security features, vulnerability fixes |
| **Usability** | Better UX, CLI improvements |

---

### 🐛 Bug Fixes

- Fixed issue where [#123](https://github.com/AnubisWatch/anubiswatch/issues/123)
- Resolved problem with [#456](https://github.com/AnubisWatch/anubiswatch/issues/456)
- Addressed edge case in [#789](https://github.com/AnubisWatch/anubiswatch/issues/789)

---

### ⚠️ Breaking Changes (if any)

**Note:** Describe any breaking changes and migration steps.

#### Migration Guide

1. Step one
2. Step two
3. Step three

If no breaking changes:
> **No breaking changes** — This is a backward-compatible release.

---

### 📦 Binary Downloads

| Platform | Architecture | Download | SHA256 |
|----------|--------------|----------|--------|
| Linux | amd64 | `anubiswatch_{{VERSION}}_linux_amd64.tar.gz` | `SHA256_HASH` |
| Linux | arm64 | `anubiswatch_{{VERSION}}_linux_arm64.tar.gz` | `SHA256_HASH` |
| Linux | arm/v7 | `anubiswatch_{{VERSION}}_linux_armv7.tar.gz` | `SHA256_HASH` |
| macOS | amd64 | `anubiswatch_{{VERSION}}_darwin_amd64.tar.gz` | `SHA256_HASH` |
| macOS | arm64 | `anubiswatch_{{VERSION}}_darwin_arm64.tar.gz` | `SHA256_HASH` |
| Windows | amd64 | `anubiswatch_{{VERSION}}_windows_amd64.zip` | `SHA256_HASH` |

---

### 📝 Verification

#### Verify Docker Image

```bash
# Pull image
docker pull ghcr.io/anubiswatch/anubiswatch:{{VERSION}}

# Verify digest
docker inspect ghcr.io/anubiswatch/anubiswatch:{{VERSION}} | grep Digest

# Expected: sha256:EXPECTED_DIGEST
```

#### Verify Binary

```bash
# Download
curl -LO https://github.com/AnubisWatch/anubiswatch/releases/download/v{{VERSION}}/anubiswatch_{{VERSION}}_linux_amd64.tar.gz

# Verify checksum
echo "SHA256_HASH anubiswatch_{{VERSION}}_linux_amd64.tar.gz" | sha256sum -c

# Should output: anubiswatch_{{VERSION}}_linux_amd64.tar.gz: OK
```

#### Verify Homebrew

```bash
brew install AnubisWatch/anubiswatch/anubiswatch
anubis version
# Should show: AnubisWatch v{{VERSION}}
```

---

### 🙏 Thank You

This release includes contributions from:
- @contributor1
- @contributor2

Want to contribute? Check out our [Contributing Guide](https://github.com/AnubisWatch/anubiswatch/blob/main/CONTRIBUTING.md).

---

### 📚 Documentation

- [Quick Start](https://github.com/AnubisWatch/anubiswatch#quick-start)
- [Configuration Reference](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/CONFIGURATION.md)
- [Deployment Guide](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/DEPLOYMENT.md)
- [API Reference](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/openapi.yaml)

---

### 🔮 What's Next

Sneak peek at what's coming in v{{NEXT_VERSION}}:

- Feature A
- Feature B
- Feature C

---

**⚖️ The Judgment Never Sleeps**

---

## Full Changelog

https://github.com/AnubisWatch/anubiswatch/compare/v{{PREV_VERSION}}...v{{VERSION}}

---

## Release Checklist (Internal)

- [ ] Update version in `cmd/anubis/main.go`
- [ ] Update CHANGELOG.md with release date
- [ ] Tag release: `git tag -a v{{VERSION}} -m "AnubisWatch v{{VERSION}}"`
- [ ] Push tag: `git push origin v{{VERSION}}`
- [ ] Wait for GitHub Actions to build binaries and Docker images
- [ ] Create GitHub Release with these notes
- [ ] Verify Docker image is available on GHCR
- [ ] Update Homebrew formula with new version and SHA256
- [ ] Post release announcement on X/Twitter
- [ ] Share on Hacker News, Reddit r/selfhosted, r/golang

---

## Example: v0.0.1 Release

```markdown
## AnubisWatch v0.0.1 — Aaru (Initial Release)

### ⚖️ The Judgment Never Sleeps

We're thrilled to announce AnubisWatch v0.0.1, codename "Aaru" (paradise in Egyptian mythology).

This is the initial beta release of AnubisWatch. Built from the ground up with zero external dependencies, AnubisWatch judges your uptime with the precision of a god.

**📦 Installation:**

```bash
# Homebrew
brew install AnubisWatch/anubiswatch/anubiswatch

# Linux
curl -fsSL https://anubis.watch/install.sh | sh

# Docker
docker pull ghcr.io/anubiswatch/anubiswatch:0.0.1
```

### 🆕 What's New

#### Eight Protocol Support
HTTP, TCP, UDP, DNS, SMTP, ICMP, gRPC, WebSocket, and TLS. If your infrastructure speaks it, we can judge it.

#### Embedded React Dashboard
Beautiful, responsive dashboard compiled into the binary. No separate frontend deployment.

#### Raft Consensus Clustering
Distributed monitoring with automatic leader election. The Necropolis never sleeps.

#### CobaltDB Storage
Embedded B+Tree database with encryption at rest. Time-series compaction built in.

#### Nine Alert Channels
Slack, Discord, Telegram, Email, PagerDuty, OpsGenie, SMS, Ntfy, and generic webhooks.

#### Public Status Pages
Custom domains, password protection, 90-day uptime history. Share your health with the world.

### 📊 Key Statistics

| Metric | Value |
|--------|-------|
| Protocols | 8 |
| Alert Channels | 9+ |
| Binary Size | <25MB |
| RAM Usage | <64MB |
| Test Coverage | 80%+ |

### 🙏 Thank You

This release represents hundreds of hours of work. Thank you to everyone who contributed, tested, and provided feedback.

### 📚 Documentation

- [Quick Start](https://github.com/AnubisWatch/anubiswatch#quick-start)
- [Configuration Reference](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/CONFIGURATION.md)
- [Deployment Guide](https://github.com/AnubisWatch/anubiswatch/blob/main/docs/DEPLOYMENT.md)

---

**⚖️ The Judgment Never Sleeps**
```
