# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

AnubisWatch is a zero-dependency, single-binary uptime and synthetic monitoring platform written in Go. It uses Egyptian mythology theming throughout the codebase.

## Common Commands

### Prerequisites

- Go 1.26+
- Node.js 22+ (for dashboard)
- Make (optional)

### Build
```bash
# Build the binary (requires dashboard built first)
make build
# Or directly: CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/anubis ./cmd/anubis

# Build dashboard (React 19 + Tailwind 4.1, embedded in binary)
make dashboard
# Or directly: cd web && npm ci && npm run build

# Dashboard dev server (hot reload)
make dashboard-dev
# Or directly: cd web && npm run dev

# Build everything
make all

# Cross-compile for all platforms
make build-all

# Build Docker image
make docker
```

### Test
```bash
# Run all tests with race detection and coverage
make test
# Or directly: go test -race -coverprofile=coverage.out ./...

# Run short tests only
make test-short

# Run a single test
go test -race -run TestName ./path/to/package

# Run integration tests (requires running server)
go test -v -tags=integration ./...
```

### Development
```bash
# Run in development mode (single node, uses ./anubis.yaml)
make dev
# Or directly: go run ./cmd/anubis serve --single --config ./anubis.yaml

# Initialize default config
anubis init

# Run with custom config
anubis serve --config ./anubis.yaml

# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Download dependencies
make deps

# Tidy go modules
make tidy
```

### CLI Commands
```bash
# Show version
anubis version
anubis version --json    # JSON output

# Initialize configuration
anubis init

# Quick-add a monitor
anubis watch https://example.com --name "Example"

# Show current judgments
anubis judge

# Server management
anubis serve --single                    # Single node mode
anubis serve --config ./anubis.yaml      # Custom config
anubis status                            # Show server status
anubis logs --follow                     # View logs
anubis config validate                   # Validate config
anubis config show                       # Show current config
anubis export --format json              # Export data

# Backup & Restore
anubis backup --output ./backup.tar.gz
anubis restore --input ./backup.tar.gz

# Cluster management
anubis necropolis              # Show cluster status
anubis summon 10.0.0.2:7946    # Add node to cluster
anubis banish jackal-02        # Remove node from cluster
```

## Architecture

### Egyptian Mythology Theming

| Term | Meaning | File Location |
|------|---------|---------------|
| **Soul** | A monitored target (HTTP endpoint, TCP port, etc.) | `internal/core/soul.go` |
| **Judgment** | A single check execution result | `internal/core/judgment.go` |
| **Verdict** | An alert decision based on judgment patterns | `internal/core/verdict.go` |
| **Jackal** | A probe node that performs health checks | `internal/probe/` |
| **Pharaoh** | The Raft leader in a cluster | `internal/raft/` |
| **Necropolis** | The distributed cluster | `internal/cluster/` |
| **Feather** | The embedded B+Tree storage engine (CobaltDB) | `internal/storage/engine.go` |
| **Ma'at** | The alert engine (goddess of truth) | `internal/alert/` |
| **Duat** | The WebSocket real-time layer | `internal/api/websocket.go` |
| **Journey** | Multi-step synthetic monitoring | `internal/journey/` |

### Project Structure

```
cmd/anubis/              # CLI entry point
├── main.go              # Command routing, CLI handling, server bootstrap
├── server.go            # Server initialization and dependency injection
├── init.go              # Config initialization (interactive and simple)
└── config.go            # Config file discovery and loading

internal/
├── core/                # Domain types (Soul, Judgment, Verdict, Config, Feather)
├── api/                 # REST API, WebSocket (Duat), MCP server, metrics, audit
├── probe/               # Protocol checkers (Jackal engine)
├── storage/             # CobaltDB B+Tree storage engine (Feather) with WAL
├── alert/               # Alert engine (Ma'at) and dispatchers
├── raft/                # Raft consensus implementation (Pharaoh)
├── cluster/             # Cluster coordination (Necropolis) and node distribution
├── journey/             # Multi-step synthetic monitoring executor
├── auth/                # Local authentication
├── acme/                # Let's Encrypt / ZeroSSL ACME integration
├── statuspage/          # Public status page handler
├── dashboard/           # Embedded React 19 dashboard (compiled into binary)
├── region/              # Multi-region support with health and replication
├── backup/              # Backup/restore functionality with compression
├── profiling/           # Built-in pprof performance profiling
├── secrets/             # Secrets management
├── metrics/             # Prometheus-compatible metrics endpoint
├── tracing/             # Distributed tracing support
├── cache/               # Caching layer
├── version/             # Version management
└── release/             # Release preparation and changelog generation

web/                     # React 19 + Tailwind 4.1 dashboard source (built into internal/dashboard/)
```

### Key Components

#### API Layer (`internal/api/`)
- `rest.go` - REST API server with custom router and handlers
- `websocket.go` - Real-time updates via WebSocket (Duat)
- `mcp.go` - Model Context Protocol server for AI integration
- `metrics.go` - Prometheus-compatible metrics endpoint
- `audit.go` - Audit logging for API operations
- `handlers_extra.go` - Additional API handlers (backup, restore, status, export)

#### Raft Consensus (`internal/raft/`)
- `node.go` - Raft node implementation
- `fsm.go` - Finite state machine for log application
- `transport.go` - HTTP transport for Raft RPC
- `distributor.go` - Work distribution across nodes
- `discovery.go` - Node discovery via mDNS/gossip

#### Probe Engine (`internal/probe/`)
- `checker.go` - Checker interface and registry
- `engine.go` - Scheduling and execution
- `http.go`, `tcp.go`, `dns.go`, etc. - Protocol implementations
- All checkers implement the `Checker` interface with `Type()`, `Judge()`, and `Validate()` methods
- Supports 10 protocols: HTTP/HTTPS, TCP, UDP, DNS, ICMP, SMTP, IMAP, gRPC, WebSocket, TLS

#### Multi-Region (`internal/region/`)
- `manager.go` - Region lifecycle management
- `routing.go` - Geographic routing rules
- `replication.go` - Cross-region data replication
- `health.go` - Region health monitoring

#### Storage (`internal/storage/`)
- `engine.go` - CobaltDB B+Tree implementation with configurable order (default 32)
- `judgments.go` - Time-series judgment storage
- `timeseries.go` - General time-series data support
- `raft_log.go` - Raft log storage adapter
- `retention.go` - Data retention and cleanup policies
- `engine_journey.go` - Journey-specific storage
- Uses WAL for crash recovery

## Domain Types

### Soul Status Values
- `alive` - Service healthy
- `dead` - Service failing
- `degraded` - Performance issues
- `unknown` - Not yet checked
- `embalmed` - Maintenance mode

### Check Types
`http`, `tcp`, `udp`, `dns`, `smtp`, `imap`, `icmp`, `grpc`, `websocket`, `tls`

## Configuration

Config files support JSON or YAML format. Default locations checked in order:
1. `./anubis.json`
2. `./anubis.yaml`
3. `~/.config/anubis/anubis.json`
4. `/etc/anubis/anubis.json`

Environment variables:
- `ANUBIS_CONFIG` - Config file path
- `ANUBIS_DATA_DIR` - Data directory (default: `/var/lib/anubis`)
- `ANUBIS_LOG_LEVEL` - Log level (debug, info, warn, error)

## Testing Guidelines

- All packages should maintain >80% test coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies (network calls, time)
- Run with `-race` flag to detect race conditions
- Integration tests use `//go:build integration` tag
- Chaos testing available in `internal/raft/chaos_test.go`
- Benchmark tests available in probe, storage, and API packages

## Dependencies

Minimal external dependencies (zero-dependency goal):
- `golang.org/x/net` - Extended networking
- `gopkg.in/yaml.v3` - YAML parsing
- `golang.org/x/sys` - System calls (indirect)
- `golang.org/x/text` - Text processing (indirect)
- `github.com/gorilla/websocket` - WebSocket support (indirect)

Dashboard (Node.js):
- React 19, Tailwind 4.1, Vite 6
- Recharts for visualizations, Zustand for state

Module: `github.com/AnubisWatch/anubiswatch`
Go version: 1.26+