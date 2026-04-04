# AnubisWatch — Documentation Index

> **Complete guide to AnubisWatch documentation**
> **Version:** 1.0.0

---

## Quick Links

| Document | Description | Location |
|----------|-------------|----------|
| **README.md** | Getting started, features, quick start | [Root](../README.md) |
| **CHANGELOG.md** | Release notes and version history | [Root](../CHANGELOG.md) |
| **CONTRIBUTING.md** | How to contribute to AnubisWatch | [Root](../CONTRIBUTING.md) |
| **DEPLOYMENT.md** | Deployment guides (Docker, K8s, systemd) | [Root](../DEPLOYMENT.md) |
| **GHCR.md** | Container registry and image docs | [Root](../GHCR.md) |

---

## Core Documentation

### For Users

| Document | Description | Location |
|----------|-------------|----------|
| **CONFIGURATION.md** | Complete `anubis.yaml` reference | [docs/](CONFIGURATION.md) |
| **openapi.yaml** | REST API specification (OpenAPI 3.1.0) | [docs/](openapi.yaml) |
| **DEPLOYMENT.md** | Production deployment guide | [Root](../DEPLOYMENT.md) |
| **GHCR.md** | Docker image and container docs | [Root](../GHCR.md) |

### For Developers

| Document | Description | Location |
|----------|-------------|----------|
| **CONTRIBUTING.md** | Contribution guidelines | [Root](../CONTRIBUTING.md) |
| **RELEASE_TEMPLATE.md** | GitHub Release template | [.github/](../.github/RELEASE_TEMPLATE.md) |
| **openapi.yaml** | API specification for client generation | [docs/](openapi.yaml) |

### For Design/Branding

| Document | Description | Location |
|----------|-------------|----------|
| **BRANDING.md** | Brand identity, colors, typography | [.project/](../.project/BRANDING.md) |
| **WEBSITE.md** | anubis.watch landing page content | [docs/](WEBSITE.md) |
| **ANUBISWATCH_LOGO_PROMPTS.md** | Logo generation prompts | [.project/](../.project/ANUBISWATCH_LOGO_PROMPTS.md) |

### For Architecture/Planning

| Document | Description | Location |
|----------|-------------|----------|
| **SPECIFICATION.md** | Technical specification | [.project/](../.project/SPECIFICATION.md) |
| **IMPLEMENTATION.md** | Implementation plan | [.project/](../.project/IMPLEMENTATION.md) |
| **TASKS.md** | Development task breakdown | [.project/](../.project/TASKS.md) |
| **BRANDING.md** | Brand guidelines | [.project/](../.project/BRANDING.md) |

---

## Documentation by Topic

### Getting Started

1. **README.md** — Quick start and overview
2. **DEPLOYMENT.md** — Installation options
3. **CONFIGURATION.md** — Configuration guide
4. **GHCR.md** — Docker images

### Configuration

1. **CONFIGURATION.md** — Full configuration reference
2. **configs/anubis.example.yaml** — Example configuration
3. **openapi.yaml** — API schema definitions

### Deployment

1. **DEPLOYMENT.md** — Deployment methods
2. **GHCR.md** — Container images
3. **docker-compose.yml** — Docker Compose examples
4. **deployments/charts/anubiswatch/** — Helm chart

### Development

1. **CONTRIBUTING.md** — How to contribute
2. **SPECIFICATION.md** — Technical specs
3. **openapi.yaml** — API specification
4. **RELEASE_TEMPLATE.md** — Release process

### Operations

1. **CONFIGURATION.md** — Configuring AnubisWatch
2. **DEPLOYMENT.md** — Production deployments
3. **GHCR.md** — Container management
4. **CHANGELOG.md** — Version history

---

## Documentation Standards

### Format

- **Markdown** for all documentation
- **YAML** for configuration examples
- **OpenAPI 3.1.0** for API specification

### Style Guide

- Use **bold** for emphasis
- Use `code` for inline code, commands, and file paths
- Use code blocks with language specification
- Use tables for structured data
- Use links for cross-references

### Naming Conventions

- **File names:** UPPERCASE.md for root docs, lowercase.md for subdirectories
- **Directories:** lowercase with hyphens (e.g., `alert-channels`)
- **Code files:** snake_case.go for Go, kebab-case.ts for TypeScript

---

## Missing Documentation (TODO)

| Document | Priority | Description |
|----------|----------|-------------|
| CLI Reference | High | Complete CLI command reference |
| Protocol Guides | Medium | Deep-dive into each protocol checker |
| Troubleshooting Guide | Medium | Common issues and solutions |
| FAQ | Low | Frequently asked questions |
| Tutorial Series | Medium | Step-by-step tutorials |
| Architecture Diagrams | High | Visual architecture docs |

---

## Documentation Maintenance

### When to Update

- **On every feature addition:** Update README, CONFIGURATION, openapi.yaml
- **On every bug fix:** Update CHANGELOG if user-impacting
- **On every release:** Update CHANGELOG, create GitHub Release
- **On every breaking change:** Update CHANGELOG, add migration guide

### Review Cycle

- **Monthly:** Review all docs for accuracy
- **Per-release:** Full documentation audit
- **As-needed:** Fix typos, clarify confusing sections

---

## Support

- **GitHub Issues:** https://github.com/AnubisWatch/anubiswatch/issues
- **GitHub Discussions:** https://github.com/AnubisWatch/anubiswatch/discussions
- **Discord:** https://discord.gg/anubiswatch
- **X/Twitter:** https://x.com/AnubisWatch

---

**⚖️ The Judgment Never Sleeps**
