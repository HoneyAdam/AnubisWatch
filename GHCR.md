# AnubisWatch Container Images

**AnubisWatch** is published on **GitHub Container Registry (GHCR)** at:

```
ghcr.io/anubiswatch/anubiswatch
```

Docker Hub is **not** used. All official images are available exclusively via GHCR.

---

## Quick Start

```bash
# Pull latest release
docker pull ghcr.io/anubiswatch/anubiswatch:latest

# Pull specific version
docker pull ghcr.io/anubiswatch/anubiswatch:v1.0.0

# Run single node
docker run -d \
  --name anubis \
  -p 8443:8443 \
  -v anubis-data:/var/lib/anubis \
  ghcr.io/anubiswatch/anubiswatch:latest
```

---

## Available Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `v1.0.0` | Specific version (semver) |
| `dev` | Development build (main branch) |

---

## Multi-Architecture Support

Images are built for the following architectures:

- `linux/amd64` - x86_64 (Intel/AMD)
- `linux/arm64` - ARM 64-bit (Apple Silicon, Raspberry Pi 4, AWS Graviton)
- `linux/arm/v7` - ARM v7 (Raspberry Pi 3, older ARM devices)

Docker will automatically pull the correct image for your architecture.

---

## Kubernetes Helm Chart

The Helm chart uses GHCR by default:

```bash
# Add the Helm repository
helm repo add anubiswatch https://anubiswatch.github.io/helm-charts
helm repo update

# Install AnubisWatch
helm install anubis anubiswatch/anubiswatch \
  --namespace monitoring \
  --create-namespace
```

### Custom Image Configuration

```yaml
# values.yaml
image:
  repository: ghcr.io/anubiswatch/anubiswatch
  tag: "v1.0.0"
  pullPolicy: IfNotPresent
```

---

## Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  anubis:
    image: ghcr.io/anubiswatch/anubiswatch:latest
    container_name: anubis
    restart: unless-stopped
    ports:
      - "8443:8443"
    volumes:
      - anubis-data:/var/lib/anubis
      - ./anubis.yaml:/etc/anubis/anubis.yaml:ro
    environment:
      - ANUBIS_LOG_LEVEL=info
```

---

## Image Labels

The image includes OCI-compliant labels:

```dockerfile
LABEL org.opencontainers.image.title="AnubisWatch"
LABEL org.opencontainers.image.description="The Judgment Never Sleeps — Zero-dependency uptime monitoring"
LABEL org.opencontainers.image.vendor="ECOSTACK TECHNOLOGY OÜ"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/AnubisWatch/anubiswatch"
```

Inspect labels:

```bash
docker inspect ghcr.io/anubiswatch/anubiswatch:latest
```

---

## Security

- **Rootless**: Image runs as non-root user (UID/GID: 65534)
- **Scratch base**: Built from `scratch` (no OS, minimal attack surface)
- **Read-only filesystem**: Container filesystem is read-only
- **No shell**: No shell or package manager in the image

### Recommended Security Context

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  runAsGroup: 65534
  fsGroup: 65534
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

---

## Volumes

| Mount Point | Purpose | Recommended |
|-------------|---------|-------------|
| `/var/lib/anubis` | Data storage (CobaltDB) | **Required** |
| `/etc/anubis/anubis.yaml` | Configuration file | Recommended |

---

## Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 8443 | TCP | HTTPS API + Dashboard |
| 7946 | TCP | Raft consensus |
| 7946 | UDP | Gossip discovery |
| 9090 | TCP | gRPC API (optional) |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ANUBIS_CONFIG` | `/etc/anubis/anubis.yaml` | Config file path |
| `ANUBIS_DATA_DIR` | `/var/lib/anubis/data` | Data directory |
| `ANUBIS_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `ANUBIS_LOG_FORMAT` | `json` | Log format (json, text) |
| `ANUBIS_NODE_ID` | Auto-generated | Cluster node ID |
| `ANUBIS_CLUSTER_SECRET` | - | Raft cluster shared secret |

---

## Building Locally

```bash
# Build image
docker build -t ghcr.io/anubiswatch/anubiswatch:dev .

# Test locally
docker run -d -p 8443:8443 ghcr.io/anubiswatch/anubiswatch:dev
```

---

## GitHub Actions (CI/CD)

AnubisWatch uses GitHub Actions for automated builds:

```yaml
# .github/workflows/release.yml
name: Build and Push

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ghcr.io/anubiswatch/anubiswatch:${{ github.ref_name }}
            ghcr.io/anubiswatch/anubiswatch:latest
          platforms: linux/amd64,linux/arm64,linux/arm/v7
```

---

## Troubleshooting

### Pull Permission Denied

GHCR packages are public. If you get permission errors:

```bash
# Ensure you're not logged in with wrong credentials
docker logout ghcr.io

# Or use anonymous pull
docker pull --platform linux/amd64 ghcr.io/anubiswatch/anubiswatch:latest
```

### Image Not Found

Verify the tag exists:

```bash
# List available tags (requires GitHub CLI)
gh api /users/anubiswatch/packages/container/anubiswatch/versions
```

---

## Support

- **Documentation:** https://github.com/AnubisWatch/anubiswatch
- **Issues:** https://github.com/AnubisWatch/anubiswatch/issues
- **Discussions:** https://github.com/AnubisWatch/anubiswatch/discussions

---

**⚖️ The Judgment Never Sleeps**
