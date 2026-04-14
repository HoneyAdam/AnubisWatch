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

<!-- rtk-instructions v2 -->
# RTK (Rust Token Killer) - Token-Optimized Commands

## Golden Rule

**Always prefix commands with `rtk`**. If RTK has a dedicated filter, it uses it. If not, it passes through unchanged. This means RTK is always safe to use.

**Important**: Even in command chains with `&&`, use `rtk`:
```bash
# ❌ Wrong
git add . && git commit -m "msg" && git push

# ✅ Correct
rtk git add . && rtk git commit -m "msg" && rtk git push
```

## RTK Commands by Workflow

### Build & Compile (80-90% savings)
```bash
rtk cargo build         # Cargo build output
rtk cargo check         # Cargo check output
rtk cargo clippy        # Clippy warnings grouped by file (80%)
rtk tsc                 # TypeScript errors grouped by file/code (83%)
rtk lint                # ESLint/Biome violations grouped (84%)
rtk prettier --check    # Files needing format only (70%)
rtk next build          # Next.js build with route metrics (87%)
```

### Test (90-99% savings)
```bash
rtk cargo test          # Cargo test failures only (90%)
rtk vitest run          # Vitest failures only (99.5%)
rtk playwright test     # Playwright failures only (94%)
rtk test <cmd>          # Generic test wrapper - failures only
```

### Git (59-80% savings)
```bash
rtk git status          # Compact status
rtk git log             # Compact log (works with all git flags)
rtk git diff            # Compact diff (80%)
rtk git show            # Compact show (80%)
rtk git add             # Ultra-compact confirmations (59%)
rtk git commit          # Ultra-compact confirmations (59%)
rtk git push            # Ultra-compact confirmations
rtk git pull            # Ultra-compact confirmations
rtk git branch          # Compact branch list
rtk git fetch           # Compact fetch
rtk git stash           # Compact stash
rtk git worktree        # Compact worktree
```

Note: Git passthrough works for ALL subcommands, even those not explicitly listed.

### GitHub (26-87% savings)
```bash
rtk gh pr view <num>    # Compact PR view (87%)
rtk gh pr checks        # Compact PR checks (79%)
rtk gh run list         # Compact workflow runs (82%)
rtk gh issue list       # Compact issue list (80%)
rtk gh api              # Compact API responses (26%)
```

### JavaScript/TypeScript Tooling (70-90% savings)
```bash
rtk pnpm list           # Compact dependency tree (70%)
rtk pnpm outdated       # Compact outdated packages (80%)
rtk pnpm install        # Compact install output (90%)
rtk npm run <script>    # Compact npm script output
rtk npx <cmd>           # Compact npx command output
rtk prisma              # Prisma without ASCII art (88%)
```

### Files & Search (60-75% savings)
```bash
rtk ls <path>           # Tree format, compact (65%)
rtk read <file>         # Code reading with filtering (60%)
rtk grep <pattern>      # Search grouped by file (75%)
rtk find <pattern>      # Find grouped by directory (70%)
```

### Analysis & Debug (70-90% savings)
```bash
rtk err <cmd>           # Filter errors only from any command
rtk log <file>          # Deduplicated logs with counts
rtk json <file>         # JSON structure without values
rtk deps                # Dependency overview
rtk env                 # Environment variables compact
rtk summary <cmd>       # Smart summary of command output
rtk diff                # Ultra-compact diffs
```

### Infrastructure (85% savings)
```bash
rtk docker ps           # Compact container list
rtk docker images       # Compact image list
rtk docker logs <c>     # Deduplicated logs
rtk kubectl get         # Compact resource list
rtk kubectl logs        # Deduplicated pod logs
```

### Network (65-70% savings)
```bash
rtk curl <url>          # Compact HTTP responses (70%)
rtk wget <url>          # Compact download output (65%)
```

### Meta Commands
```bash
rtk gain                # View token savings statistics
rtk gain --history      # View command history with savings
rtk discover            # Analyze Claude Code sessions for missed RTK usage
rtk proxy <cmd>         # Run command without filtering (for debugging)
rtk init                # Add RTK instructions to CLAUDE.md
rtk init --global       # Add RTK to ~/.claude/CLAUDE.md
```

## Token Savings Overview

| Category | Commands | Typical Savings |
|----------|----------|-----------------|
| Tests | vitest, playwright, cargo test | 90-99% |
| Build | next, tsc, lint, prettier | 70-87% |
| Git | status, log, diff, add, commit | 59-80% |
| GitHub | gh pr, gh run, gh issue | 26-87% |
| Package Managers | pnpm, npm, npx | 70-90% |
| Files | ls, read, grep, find | 60-75% |
| Infrastructure | docker, kubectl | 85% |
| Network | curl, wget | 65-70% |

Overall average: **60-90% token reduction** on common development operations.
<!-- /rtk-instructions -->