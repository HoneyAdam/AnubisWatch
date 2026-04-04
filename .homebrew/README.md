# Homebrew Tap for AnubisWatch

[![Homebrew](https://img.shields.io/badge/dynamic/json?url=https://formulae.brew.sh/api/formula/anubiswatch.json&query=$.versions.stable&label=homebrew&logo=homebrew&logoColor=white&color=fbb040)](https://github.com/AnubisWatch/homebrew-anubiswatch)

Official Homebrew tap for [AnubisWatch](https://anubis.watch) — The Judgment Never Sleeps.

## Installation

### Install the Formula

```bash
# Add the tap
brew tap AnubisWatch/anubiswatch

# Install AnubisWatch
brew install anubiswatch
```

### One-Line Install

```bash
brew install AnubisWatch/anubiswatch/anubiswatch
```

## Usage

### Quick Start

```bash
# Generate default configuration
anubis init

# Start AnubisWatch
anubis serve

# Open dashboard in browser
open https://localhost:8443
```

### As a Service

```bash
# Start the service
brew services start anubiswatch

# Check status
brew services info anubiswatch

# View logs
brew services logs anubiswatch

# Stop the service
brew services stop anubiswatch

# Restart the service
brew services restart anubiswatch
```

### CLI Commands

```bash
# Version info
anubis version

# Health check
anubis health

# Generate config
anubis init --config ./anubis.yaml

# Single-node mode
anubis serve --config ./anubis.yaml

# Cluster mode
anubis serve --cluster --config ./anubis.yaml

# Cluster management
anubis summon raft://node.example.com:7946
anubis banish jackal-us-01
anubis necropolis status
```

## Configuration

### Default Config Location

- **macOS (Apple Silicon):** `/opt/homebrew/etc/anubis/anubis.yaml`
- **macOS (Intel):** `/usr/local/etc/anubis/anubis.yaml`
- **Linux:** `/etc/anubis/anubis.yaml`

### Data Directory

- **macOS (Apple Silicon):** `/opt/homebrew/var/lib/anubis/data`
- **macOS (Intel):** `/usr/local/var/lib/anubis/data`
- **Linux:** `/var/lib/anubis/data`

### Example Configuration

```yaml
# /opt/homebrew/etc/anubis/anubis.yaml
server:
  port: 8443
  tls:
    enabled: true
    auto_cert: true
    acme_email: "admin@example.com"
    acme_domains:
      - "anubis.local"

storage:
  path: /opt/homebrew/var/lib/anubis/data

souls:
  - name: "Google DNS"
    type: dns
    target: "google.com"
    weight: 60s
    dns:
      record_type: A

channels:
  - name: "ntfy-alerts"
    type: ntfy
    ntfy:
      server: "https://ntfy.sh"
      topic: "anubis-alerts"

verdicts:
  rules:
    - name: "DNS Down"
      scope: "type:dns"
      condition:
        type: consecutive_failures
        threshold: 3
      severity: critical
      channels: ["ntfy-alerts"]
      cooldown: 5m
```

## Development

### Build from Source

```bash
brew install --build-from-source anubiswatch
```

### Edit Formula

```bash
brew edit anubiswatch
```

### Run Tests

```bash
brew test anubiswatch
```

### Uninstall

```bash
# Stop service
brew services stop anubiswatch

# Uninstall
brew uninstall anubiswatch

# Remove config and data (optional)
rm -rf /opt/homebrew/etc/anubis
rm -rf /opt/homebrew/var/lib/anubis
```

## Troubleshooting

### Permission Denied

If you encounter permission errors:

```bash
# Fix permissions on data directory
sudo chown -R $(whoami) /opt/homebrew/var/lib/anubis/data
```

### Port Already in Use

If port 8443 is in use:

```bash
# Edit config to use different port
# server.port: 8444
anubis serve --config /opt/homebrew/etc/anubis/anubis.yaml
```

### Service Won't Start

Check logs:

```bash
brew services logs anubiswatch
```

Validate config:

```bash
anubis validate --config /opt/homebrew/etc/anubis/anubis.yaml
```

## Requirements

- macOS 10.15+ or Linux with Homebrew
- Go 1.26+ (for building from source)

## License

Apache 2.0 — See [LICENSE](https://github.com/AnubisWatch/anubiswatch/blob/main/LICENSE)

## Links

- **Homepage:** https://anubis.watch
- **GitHub:** https://github.com/AnubisWatch/anubiswatch
- **Documentation:** https://github.com/AnubisWatch/anubiswatch/tree/main/docs
- **Issues:** https://github.com/AnubisWatch/anubiswatch/issues

---

**⚖️ The Judgment Never Sleeps**
