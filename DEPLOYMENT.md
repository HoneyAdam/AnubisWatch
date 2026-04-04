# AnubisWatch — Deployment Guide

> **Deployment Options & Configuration**
> **Version:** 1.0.0

---

## Table of Contents

1. [Linux Binary Installation](#linux-binary-installation)
2. [Docker Deployment](#docker-deployment)
3. [Docker Compose](#docker-compose)
4. [systemd Service](#systemd-service)
5. [Kubernetes](#kubernetes)
6. [Environment Variables](#environment-variables)
7. [Configuration Reference](#configuration-reference)

---

## Linux Binary Installation

### Install Script (Recommended)

```bash
# Install latest release
curl -fsSL https://anubis.watch/install.sh | sh

# Install specific version
curl -fsSL https://anubis.watch/install.sh | sh -s -- --version v1.0.0
```

### Manual Installation

```bash
# Download binary
wget https://github.com/AnubisWatch/anubiswatch/releases/latest/download/anubis-linux-amd64

# Verify checksum (optional but recommended)
sha256sum anubis-linux-amd64
# Compare with checksum from releases page

# Install
sudo chmod +x anubis-linux-amd64
sudo mv anubis-linux-amd64 /usr/local/bin/anubis

# Verify installation
anubis version
```

### Create User and Directories

```bash
# Create dedicated user
sudo useradd --system --no-create-home --shell /usr/sbin/nologin anubis

# Create data directory
sudo mkdir -p /var/lib/anubis/data
sudo chown anubis:anubis /var/lib/anubis

# Create config directory
sudo mkdir -p /etc/anubis
```

---

## Docker Deployment

### Pull Image

```bash
# Pull from GHCR (GitHub Container Registry)
docker pull ghcr.io/anubiswatch/anubiswatch:latest

# Pull specific version
docker pull ghcr.io/anubiswatch/anubiswatch:v1.0.0
```

### Run Single Node

```bash
docker run -d \
  --name anubis \
  --restart unless-stopped \
  -p 8443:8443 \
  -v anubis-data:/var/lib/anubis \
  -v $(pwd)/anubis.yaml:/etc/anubis/anubis.yaml:ro \
  -e ANUBIS_LOG_LEVEL=info \
  ghcr.io/anubiswatch/anubiswatch:latest
```

### Run with Custom Network

```bash
# Create network
docker network create monitoring

# Run AnubisWatch
docker run -d \
  --name anubis \
  --network monitoring \
  -p 8443:8443 \
  -v anubis-data:/var/lib/anubis \
  ghcr.io/anubiswatch/anubiswatch:latest
```

### Volume Permissions

If you encounter permission issues:

```bash
# Option 1: Run as root (not recommended for production)
docker run -d --user 0:0 ...

# Option 2: Set correct ownership
docker volume create anubis-data
docker run --rm -v anubis-data:/data alpine chown -R 65534:65534 /data
```

---

## Docker Compose

### Single Node

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
      - ANUBIS_CONFIG=/etc/anubis/anubis.yaml

volumes:
  anubis-data:
    driver: local
```

```bash
docker-compose up -d
```

### Multi-Node Cluster

```yaml
# docker-compose.cluster.yml
version: '3.8'

services:
  anubis-1:
    image: ghcr.io/anubiswatch/anubiswatch:latest
    container_name: anubis-1
    hostname: anubis-1
    restart: unless-stopped
    ports:
      - "8443:8443"
      - "7946:7946"
    volumes:
      - data-1:/var/lib/anubis
    environment:
      - ANUBIS_NODE_ID=anubis-1
      - ANUBIS_CLUSTER_SECRET=your-secret-key
    command: ["serve", "--cluster", "--bootstrap"]
    networks:
      - anubis-net

  anubis-2:
    image: ghcr.io/anubiswatch/anubiswatch:latest
    container_name: anubis-2
    hostname: anubis-2
    restart: unless-stopped
    ports:
      - "8444:8443"
    volumes:
      - data-2:/var/lib/anubis
    environment:
      - ANUBIS_NODE_ID=anubis-2
      - ANUBIS_CLUSTER_SECRET=your-secret-key
    command: ["serve", "--cluster", "--join", "anubis-1:7946"]
    networks:
      - anubis-net
    depends_on:
      - anubis-1

  anubis-3:
    image: ghcr.io/anubiswatch/anubiswatch:latest
    container_name: anubis-3
    hostname: anubis-3
    restart: unless-stopped
    ports:
      - "8445:8443"
    volumes:
      - data-3:/var/lib/anubis
    environment:
      - ANUBIS_NODE_ID=anubis-3
      - ANUBIS_CLUSTER_SECRET=your-secret-key
    command: ["serve", "--cluster", "--join", "anubis-1:7946"]
    networks:
      - anubis-net
    depends_on:
      - anubis-1

volumes:
  data-1:
  data-2:
  data-3:

networks:
  anubis-net:
    driver: bridge
```

```bash
docker-compose -f docker-compose.cluster.yml up -d
```

---

## systemd Service

### Create Service File

```bash
sudo tee /etc/systemd/system/anubis.service > /dev/null <<'EOF'
[Unit]
Description=AnubisWatch Uptime Monitoring
Documentation=https://github.com/AnubisWatch/anubiswatch
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=anubis
Group=anubis

# Environment
Environment="ANUBIS_CONFIG=/etc/anubis/anubis.yaml"
Environment="ANUBIS_DATA_DIR=/var/lib/anubis/data"
Environment="ANUBIS_LOG_LEVEL=info"

# Executable
ExecStart=/usr/local/bin/anubis serve
ExecReload=/bin/kill -HUP $MAINPID

# Restart policy
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/anubis

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
```

### Enable and Start

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable anubis

# Start service
sudo systemctl start anubis

# Check status
sudo systemctl status anubis

# View logs
sudo journalctl -u anubis -f
```

### Configuration Locations

| File/Directory | Purpose |
|----------------|---------|
| `/etc/anubis/anubis.yaml` | Main configuration |
| `/var/lib/anubis/data/` | Data storage (CobaltDB) |
| `/var/log/anubis/` | Log files (if not using journal) |

---

## Kubernetes

### Namespace and ConfigMap

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
  labels:
    name: monitoring
```

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: anubis-config
  namespace: monitoring
data:
  anubis.yaml: |
    server:
      host: "0.0.0.0"
      port: 8443
    
    storage:
      path: "/var/lib/anubis/data"
      retention_days: 90
    
    souls:
      - name: "Kubernetes API"
        type: http
        target: "https://kubernetes.default.svc/healthz"
        weight: 30s
        http:
          method: GET
          valid_status: [200]
```

### StatefulSet

```yaml
# statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: anubis
  namespace: monitoring
spec:
  serviceName: anubis
  replicas: 3
  selector:
    matchLabels:
      app: anubis
  template:
    metadata:
      labels:
        app: anubis
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534
      containers:
        - name: anubis
          image: ghcr.io/anubiswatch/anubiswatch:latest
          imagePullPolicy: Always
          args:
            - "serve"
            - "--cluster"
            - "--join"
            - "anubis-0.anubis.monitoring.svc.cluster.local:7946"
          ports:
            - name: http
              containerPort: 8443
              protocol: TCP
            - name: raft
              containerPort: 7946
              protocol: TCP
            - name: gossip
              containerPort: 7946
              protocol: UDP
          volumeMounts:
            - name: data
              mountPath: /var/lib/anubis
            - name: config
              mountPath: /etc/anubis
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
      volumes:
        - name: config
          configMap:
            name: anubis-config
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
```

### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: anubis
  namespace: monitoring
spec:
  clusterIP: None
  selector:
    app: anubis
  ports:
    - name: http
      port: 8443
      targetPort: http
    - name: raft
      port: 7946
      targetPort: raft
```

### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: anubis
  namespace: monitoring
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - anubis.example.com
      secretName: anubis-tls
  rules:
    - host: anubis.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: anubis
                port:
                  name: http
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ANUBIS_CONFIG` | Path to config file | `./anubis.yaml` |
| `ANUBIS_HOST` | Server bind host | `0.0.0.0` |
| `ANUBIS_PORT` | Server bind port | `8443` |
| `ANUBIS_DATA_DIR` | Data directory | `./data` |
| `ANUBIS_LOG_LEVEL` | Log level | `info` |
| `ANUBIS_LOG_FORMAT` | Log format | `json` |
| `ANUBIS_NODE_ID` | Cluster node ID | Auto-generated |
| `ANUBIS_CLUSTER_SECRET` | Raft cluster secret | - |
| `ANUBIS_ENCRYPTION_KEY` | Storage encryption key | - |

---

## Configuration Reference

### Server

```yaml
server:
  host: "0.0.0.0"      # Bind address
  port: 8443           # Bind port
  tls:
    enabled: false     # Enable TLS
    cert: ""           # Certificate path
    key: ""            # Key path
    auto_cert: false   # Auto-cert via ACME
    acme_email: ""     # ACME registration email
```

### Storage

```yaml
storage:
  path: "./data"       # Data directory
  retention_days: 90   # Data retention
  encryption:
    enabled: false     # Enable encryption
    key: ""            # Encryption key (32 bytes)
```

### Cluster (Necropolis)

```yaml
necropolis:
  enabled: false       # Enable clustering
  node_name: "jackal-1"
  region: "default"
  bind_addr: "0.0.0.0:7946"
  cluster_secret: ""   # Shared secret
  discovery:
    mode: "mdns"       # mdns, gossip, or manual
    seeds: []          # Seed nodes for manual
```

### Souls (Monitors)

```yaml
souls:
  - name: "Example"
    type: http         # http, tcp, udp, dns, icmp, smtp, grpc, ws, tls
    target: "https://example.com"
    weight: 60s        # Check interval
    timeout: 10s       # Timeout
    enabled: true
    tags:
      - production
      - api
    http:
      method: GET
      valid_status: [200, 201]
      feather: 500ms   # Max latency budget
```

### Channels (Alerts)

```yaml
channels:
  - name: "ops-slack"
    type: slack
    enabled: true
    slack:
      webhook_url: "${SLACK_WEBHOOK}"
      channel: "#ops"
```

### Verdicts (Alert Rules)

```yaml
verdicts:
  rules:
    - name: "Service Down"
      enabled: true
      condition:
        type: consecutive_failures
        threshold: 3
      severity: critical  # critical, high, medium, low
      channels: ["ops-slack"]
      cooldown: 5m
```

---

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
sudo lsof -i :8443
sudo systemctl stop anubis
# Change port in config
```

**Permission denied:**
```bash
sudo chown -R anubis:anubis /var/lib/anubis
sudo chmod 750 /var/lib/anubis
```

**Cluster not forming:**
```bash
# Check firewall rules
sudo ufw allow 7946/tcp
sudo ufw allow 7946/udp

# Verify network connectivity
nc -zv node1 7946
```

### Logs

```bash
# systemd
journalctl -u anubis -f

# Docker
docker logs -f anubis

# Kubernetes
kubectl logs -n monitoring -l app=anubis -f
```

### Health Check

```bash
curl http://localhost:8443/health
curl http://localhost:8443/ready
```

---

## Support

- **Documentation:** https://github.com/AnubisWatch/anubiswatch
- **Issues:** https://github.com/AnubisWatch/anubiswatch/issues
- **Discussions:** https://github.com/AnubisWatch/anubiswatch/discussions

---

**⚖️ The Judgment Never Sleeps**
