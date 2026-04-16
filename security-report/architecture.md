# AnubisWatch Security Architecture Report — Phase 1a: Reconnaissance

**Target:** AnubisWatch Uptime Monitoring Platform
**Date:** 2026-04-16
**Languages:** Go 1.26.2 (backend), TypeScript/React 19 (dashboard)
**Classification:** Internal Security Assessment

---

## 1. Technology Stack Detection

### Backend (Go)
| Component | Version/Technology | File Location |
|-----------|-------------------|---------------|
| Language | Go 1.26.2 | `go.mod:3` |
| HTTP Framework | Custom router (stdlib `net/http`) | `internal/api/rest.go` |
| WebSocket | `github.com/coder/websocket v1.8.14` | `go.mod:6` |
| gRPC | `google.golang.org/grpc v1.80.0` | `go.mod:10` |
| LDAP Auth | `github.com/go-ldap/ldap/v3 v3.4.13` | `go.mod:7` |
| Cryptography | `golang.org/x/crypto v0.49.0` | `go.mod:8` |
| Networking | `golang.org/x/net v0.52.0` | `go.mod:9` |
| Config | `gopkg.in/yaml.v3 v3.0.1` | `go.mod:12` |

### Frontend (React Dashboard)
| Component | Version | File Location |
|-----------|---------|---------------|
| Framework | React 19 | `internal/dashboard/package.json:12` |
| Build Tool | Vite 6.2 | `internal/dashboard/package.json:24` |
| Styling | Tailwind CSS 4.1 | `internal/dashboard/package.json:23` |
| Charts | Recharts 2.15 | `internal/dashboard/package.json:14` |

### Build & Deployment
- **Build:** `CGO_ENABLED=0 go build` (static binary)
- **Docker:** Multi-stage build supported
- **Helm Charts:** Kubernetes deployment via `deployments/charts/anubiswatch/`

---

## 2. Application Type Classification

**Primary Classification:** Uptime/Synthetic Monitoring Platform
**Sub-classification:** Multi-tenant SaaS-ready monitoring with embedded dashboard

| Characteristic | Value |
|----------------|-------|
| Architecture | Single-binary with embedded React dashboard |
| Deployment Mode | On-premise or cloud (zero-dependency goal) |
| Cluster Mode | Raft-based distributed (optional) |
| Storage | Embedded B+Tree (CobaltDB) with WAL |
| Multi-tenancy | Workspace isolation (HIGH-09) |

**Egyptian Mythology Theming:**
- **Soul** = Monitored target (HTTP/TCP/etc.)
- **Judgment** = Single check result
- **Verdict** = Alert decision
- **Jackal** = Probe node
- **Pharaoh** = Raft leader
- **Necropolis** = Cluster
- **Feather** = B+Tree storage engine

---

## 3. Entry Points Mapping

### 3.1 CLI Commands (`cmd/anubis/main.go:21-65`)

| Command | Function | Risk |
|---------|----------|------|
| `serve` | Start server (main entry) | High (exposes all APIs) |
| `init` | Initialize config | Medium (file write) |
| `watch <target>` | Quick-add monitor | Medium (creates resources) |
| `judge [--all]` | Force check | Medium (triggers probes) |
| `summon <addr>` | Add cluster node | High (network join) |
| `banish <id>` | Remove cluster node | High (network ops) |
| `backup/restore` | Data management | High (file/directory access) |

### 3.2 REST API Routes (`internal/api/rest.go:295-450`)

**Public Endpoints (no auth):**
| Method | Path | Handler | Risk |
|--------|------|---------|------|
| GET | `/health` | `handleHealth` | Low |
| GET | `/ready` | `handleReady` | Low |
| GET | `/api/openapi.json` | `handleOpenAPIJSON` | Low |
| GET | `/api/docs` | `handleOpenAPIDocs` | Low |
| GET | `/status`, `/status.html` | `handleStatusPage` | Low |
| GET | `/public/status` | `handlePublicStatus` | Low |
| POST | `/api/v1/auth/login` | `handleLogin` | Medium |
| POST | `/api/v1/auth/logout` | `handleLogout` | Medium |
| POST | `/api/v1/auth/reset-password` | `handleRequestPasswordReset` | Medium |
| POST | `/api/v1/auth/reset-password/confirm` | `handleConfirmPasswordReset` | Medium |

**OIDC Endpoints (conditional):**
| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/v1/auth/oidc/login` | `handleOIDCLogin` |
| GET | `/api/v1/auth/oidc/callback` | `handleOIDCCallback` |

**Authenticated Endpoints (Bearer token):**
| Method | Path | Handler | IDOR Protection |
|--------|------|---------|-----------------|
| GET | `/api/v1/souls` | `handleListSouls` | Workspace filter |
| POST | `/api/v1/souls` | `handleCreateSoul` | — |
| GET | `/api/v1/souls/:id` | `handleGetSoul` | Line 828-829 |
| PUT | `/api/v1/souls/:id` | `handleUpdateSoul` | Line 847-848 |
| DELETE | `/api/v1/souls/:id` | `handleDeleteSoul` | Line 871-872 |
| POST | `/api/v1/souls/:id/check` | `handleForceCheck` | Line 889-890 |
| GET | `/api/v1/channels` | `handleListChannels` | HIGH-09 (Line 945) |
| GET | `/api/v1/rules` | `handleListRules` | HIGH-09 (Line 1074) |
| GET | `/ws` | `handleWebSocket` | Token auth |
| GET | `/api/v1/events` | `handleSSE` | Token auth |

**Role-Based Endpoints (require specific permissions):**
```
souls:*  → Create/update/delete souls
channels:* → Manage alert channels
rules:*  → Manage alert rules
settings:write → Config changes
```

### 3.3 gRPC API (`internal/grpcapi/server.go`)

| Service | Port | Auth | Key Methods |
|---------|------|------|-------------|
| `AnubisWatchService` | 9090 (default) | Bearer token interceptor (Line 1305-1337) | ListSouls, CreateSoul, ListChannels, CreateChannel |

gRPC Interceptors:
- `authInterceptor` (Line 1305) — Unary authentication
- `authStreamInterceptor` (Line 1341) — Stream authentication

### 3.4 WebSocket (`internal/api/websocket.go`)

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `/ws` | Token via query param or initial message | Real-time judgment streaming |

**Origin validation:** `allowedOrigins` whitelist (Line 44-51)

### 3.5 Internal Ports

| Port | Protocol | Purpose | Config Key |
|------|----------|---------|------------|
| 8443 | HTTP/HTTPS | REST API (default) | `server.port` |
| 9090 | gRPC | gRPC API | `server.grpc_port` |
| 7946 | Raft | Cluster communication | `necropolis.bind_addr` |
| 7947 | HTTP | Raft transport | Dynamic |

---

## 4. Data Flow Map

### 4.1 Monitoring Data Flow

```
[Target] → [Probe Engine (Jackal)] → [Judgment] → [Storage (Feather/CobaltDB)]
                                    ↓
                            [Alert Manager (Ma'at)]
                                    ↓
                    [Dispatchers] → [Notification Channels]
```

### 4.2 API Request Flow

```
[Client] → [REST Server] → [Middleware Chain]
                              ↓
                    1. loggingMiddleware (Line 1566)
                    2. securityHeadersMiddleware (Line 1731)
                    3. corsMiddleware (Line 1582)
                    4. recoveryMiddleware (Line 1642)
                    5. validateJSONMiddleware (Line 1655)
                    6. validatePathParams (Line 1755)
                    7. rateLimitMiddleware (Line 1780)
                              ↓
                    [requireAuth / requireRole]
                              ↓
                    [Handler] → [Storage Adapter] → [CobaltDB]
```

### 4.3 Authentication Flow

```
[Login Request] → [handleLogin] → [auth.Login(email, pass)]
                                        ↓
                    [LocalAuth] ──or── [LDAPAuth] ──or── [OIDCAuth]
                            ↓              ↓              ↓
                    [bcrypt hash]    [LDAP bind]    [OIDC callback]
                            ↓              ↓              ↓
                    [Session Token ← CSPRNG generated]
                            ↓
                    [Token stored in memory map]
                            ↓
                    [Return {user, token}]
```

### 4.4 Cluster Data Flow

```
[Raft Node] ←→ [TCP Transport with optional TLS]
                    ↓
[StorageFSM] → [CobaltDB] → [WAL]
                    ↓
[Distributor] → [Probe Engine] (work distribution)
```

---

## 5. Trust Boundaries

### 5.1 Authentication Boundary

| Boundary | Protection | Location |
|----------|------------|----------|
| REST API | Bearer token + workspace isolation | `rest.go:1514-1548` (requireAuth) |
| gRPC API | Bearer token + metadata | `grpcapi/server.go:1305-1337` |
| WebSocket | Token validation on connect | `websocket.go:43-61` |
| Config file | Environment variable expansion | `core/config.go:68-80` |
| OIDC state | HMAC-signed state token (cluster-compatible) | `auth/oidc.go:125-150` |

### 5.2 Input Validation

| Layer | Mechanism | Location |
|-------|-----------|----------|
| JSON body | `maxRequestBodySize` (1MB) + depth limit (32) | `rest.go:20-24` |
| Content-Type | Must be `application/json` for POST/PUT | `rest.go:1658` |
| Path params | Length limit (256) + path traversal check | `rest.go:1755-1776` |
| SQL/Injection | Pattern detection in `containsInjectionPatterns` | `rest.go:1702-1728` |
| Rate limiting | Per-IP and per-user limits | `rest.go:1779-1902` |

### 5.3 CORS Configuration (`rest.go:1610-1627`)

```go
// Priority: config.AllowedOrigins → ANUBIS_CORS_ORIGINS env → localhost defaults
Allowed Origins:
- http://localhost:3000
- http://localhost:8080
- http://127.0.0.1:3000
- http://127.0.0.1:8080
```

### 5.4 Security Headers (`rest.go:1731-1751`)

| Header | Value |
|--------|-------|
| X-Content-Type-Options | `nosniff` |
| X-Frame-Options | `DENY` |
| X-XSS-Protection | `1; mode=block` |
| Referrer-Policy | `strict-origin-when-cross-origin` |
| Content-Security-Policy | `default-src 'self'` |
| Strict-Transport-Security | `max-age=31536000; includeSubDomains; preload` |

### 5.5 Workspace Isolation (HIGH-09)

Verified in handlers:
- `handleListChannels` (Line 945): `ListChannelsByWorkspace(ctx.Workspace)`
- `handleListRules` (Line 1074): `ListRulesByWorkspace(ctx.Workspace)`
- `handleGetSoul` (Line 828-829): workspace check
- `handleUpdateSoul` (Line 847-848): workspace check
- `handleGetChannel` (Line 999-1002): workspace check
- `handleGetRule` (Line 1128-1130): workspace check
- gRPC `ListSouls` (Line 488-495): uses authenticated user's workspace

### 5.6 Rate Limiting (`rest.go:1779-1902`)

| Endpoint Category | Limit |
|-------------------|-------|
| Auth endpoints | 10 req/min |
| Sensitive ops (delete/update) | 20 req/min |
| Default | 100 req/min |

---

## 6. External Integrations

### 6.1 Databases/Storage

| Integration | Type | Location |
|-------------|------|----------|
| CobaltDB (Feather) | Embedded B+Tree | `internal/storage/engine.go` |
| Write-Ahead Log | Crash recovery | `internal/storage/engine.go:59-64` |

### 6.2 Alert Dispatchers (`internal/alert/dispatchers.go`)

| Dispatcher | Type | Config Fields |
|------------|------|---------------|
| Slack | Webhook | `webhook_url` |
| Discord | Webhook | `webhook_url` |
| Email | SMTP | `smtp_host`, `smtp_port`, `username`, `password`, `from`, `to` |
| PagerDuty | API | `integration_key` |
| NTFY | HTTP | `server`, `topic` |
| Webhook | HTTP(S) | `url`, `method`, `headers`, `secret` |

**SSRF Protection:** HTTP redirects disabled (Line 37-39)

### 6.3 Authentication Providers

| Provider | Package | Config |
|----------|---------|--------|
| Local | In-memory + bcrypt | `auth.local.admin_email`, `auth.local.admin_password` |
| OIDC | Custom implementation | `auth.oidc.issuer`, `auth.oidc.client_id`, `auth.oidc.client_secret` |
| LDAP | `go-ldap/ldap/v3` | `auth.ldap.url`, `auth.ldap.bind_dn`, `auth.ldap.bind_password` |

### 6.4 TLS/ACME

| Feature | Implementation | Location |
|---------|----------------|----------|
| TLS Server | stdlib `crypto/tls` | `rest.go:470-471` |
| Auto-cert (Let's Encrypt) | `internal/acme/` | `server.go:945-967` |
| mTLS for Raft | Optional | `cluster/manager.go:17-65` |

---

## 7. Authentication Architecture

### 7.1 Local Auth (`internal/auth/local.go`)

**Password Storage:** bcrypt (cost 12, Line 376)

**Security Controls:**
- Brute force protection (Line 376-380): 5 attempts, 15-min lockout
- Timing attack prevention (Line 294-298): dummy bcrypt on wrong email
- CSPRNG tokens (Line 355-363)
- Password policy: min 12 chars, 3 of 4 classes (Line 384-418)
- Session persistence to disk with atomic write (Line 158-198)
- Session files: mode 0600 (Line 184-197)

**Password Reset Flow:**
1. `RequestPasswordReset` (Line 531-559): constant-time email comparison, token logged to stdout
2. `ConfirmPasswordReset` (Line 563-596): validate token, enforce password policy, invalidate all sessions

**Change Password Flow (HIGH-04, Line 493-528):**
1. Validate current token
2. Verify current password
3. Validate new password policy
4. Invalidate all existing sessions

### 7.2 OIDC Auth (`internal/auth/oidc.go`)

**JWT Validation:**
- Algorithm whitelist (Line 551-559): RS256/384/512, ES256/384/512 — rejects "none"
- JWK signature verification via JWKS URI (Line 372-421)
- Claims validation: iss, aud, sub, nonce, exp, nbf (Line 617-696)

**State Parameter:**
- HMAC-signed state (cluster-compatible, Line 125-150)
- CSRF protection via nonce binding (Line 161-162, 207-214)

**Session Management:**
- 24-hour tokens stored in-memory (Line 264)
- JWK cache TTL: 24 hours (Line 121)

### 7.3 LDAP Auth (`internal/auth/ldap.go`)

**Flow:**
1. Connect with timeout (10s)
2. StartTLS if not ldaps://
3. Bind with user credentials
4. Re-bind with service account for user lookup
5. Search with escaped filter (LDAP injection prevention)

**Fallback:** Falls back to local auth on LDAP failure

### 7.4 gRPC Auth

**Interceptor (Line 1305-1337):**
```go
// Extract Bearer token from metadata
// Validate via authenticator.Authenticate(token)
// Add user to context
```

---

## 8. File Structure Analysis

### 8.1 Sensitive Files

| File | Sensitivity | Notes |
|------|-------------|-------|
| `anubis.yaml` / `anubis.json` | **CRITICAL** | Contains credentials, TLS keys, cluster secrets |
| `go.mod` | Low | Dependency manifest |
| `internal/dashboard/dist/` | Medium | Compiled React (embedded in binary) |
| `sessions.json` (runtime) | **HIGH** | Session tokens, password hashes |
| `wal.log` (runtime) | Medium | Write-ahead log |
| `.env` | **CRITICAL** | If present, contains secrets |

### 8.2 Deployment Files

| File | Purpose |
|------|---------|
| `Makefile` | Build automation |
| `Dockerfile` | Container image |
| `deployments/charts/anubiswatch/` | Helm charts for K8s |
| `deploy/k8s/base.yaml` | Kubernetes manifests |
| `buf.yaml`, `buf.gen.yaml` | gRPC code generation |

### 8.3 Configuration Loading Order (`cmd/anubis/config.go`)

1. `./anubis.json`
2. `./anubis.yaml`
3. `~/.config/anubis/anubis.json`
4. `/etc/anubis/anubis.json`

### 8.4 Directory Structure

```
D:\CODEBOX\PROJECTS\AnubisWatch\
├── cmd/anubis/          # CLI entry point
│   ├── main.go          # Command routing
│   ├── server.go        # Server bootstrap
│   ├── init.go          # Config initialization
│   └── config.go        # Config discovery
├── internal/
│   ├── api/             # REST, WebSocket, gRPC, MCP
│   ├── auth/            # local.go, oidc.go, ldap.go
│   ├── probe/           # Check engine (Jackal)
│   ├── storage/         # CobaltDB (Feather)
│   ├── alert/           # Alert manager (Ma'at)
│   ├── cluster/         # Raft coordination (Necropolis)
│   ├── core/            # Domain types, config
│   ├── dashboard/       # Embedded React
│   ├── grpcapi/         # gRPC service
│   └── ...              # Other modules
├── web/                 # React dashboard source
├── configs/             # Example configs
├── deployments/         # K8s/Helm manifests
└── docs/                # OpenAPI spec
```

---

## 9. Detected Security Controls

### 9.1 Implemented Controls

| Control | Implementation | Location |
|---------|---------------|----------|
| Input validation | JSON depth/body limits, path param validation | `rest.go:20-24`, `1655-1698`, `1755-1776` |
| Output encoding | JSON-only responses | `rest.go:2045-2049` |
| Authentication | Multi-provider (local, OIDC, LDAP) | `auth/` |
| Session management | CSPRNG tokens, expiration, persistence | `local.go:355-373` |
| Password policy | MIN 12 chars, 3/4 classes | `local.go:384-418` |
| Brute force protection | 5 attempts, 15-min lockout | `local.go:376-380`, `422-444` |
| Timing attack prevention | Constant-time comparisons | `local.go:292-298` |
| IDOR protection | Workspace isolation on all resources | `rest.go:828-829, 847-848, 871-872, etc.` |
| CORS protection | Whitelist-based origin validation | `rest.go:1582-1607`, `1610-1640` |
| Rate limiting | Per-IP and per-user limits | `rest.go:1779-1902` |
| Security headers | HSTS, CSP, X-Frame-Options, etc. | `rest.go:1731-1751` |
| JWT validation | Algorithm whitelist, signature verification | `auth/oidc.go:551-559`, `529-614` |
| OIDC state CSRF | HMAC-signed state parameter | `auth/oidc.go:125-150` |
| Workspace isolation | Multi-tenant data separation (HIGH-09) | Throughout handlers |
| SSRF protection | HTTP redirect disabled in alert dispatchers | `alert/dispatchers.go:37-39` |
| TLS configuration | Min TLS 1.2, secure cipher suites | `cluster/manager.go:32-38` |
| Panic recovery | Deferred recovery middleware | `rest.go:1642-1651` |
| Error masking | Internal errors don't leak details | `rest.go:2056-2060` |
| Mass assignment prevention | Whitelist-based config field updates | `grpcapi/server.go:891-919` |
| WAL for crash recovery | Write-ahead log in storage | `storage/engine.go:59-64` |
| Atomic file writes | Temp file + rename for sessions | `local.go:182-195` |

### 9.2 Security-Related Recent Commits

| Commit | Fix |
|--------|-----|
| `a846bbd` | HIGH-04: password change and reset mechanisms |
| `d24a250` | HIGH-15: gorilla/websocket → coder/websocket |
| `d61f8a4` | HIGH-14: Go 1.26.1 → 1.26.2 (CVE patch) |
| `3b01f23` | MED-06, MED-09, MED-11, MED-12, MED-18: password policy, JWT alg, workspace isolation |

---

## 10. Detected Languages (for Phase 2 Skill Activation)

| Language | Files | Primary Use |
|----------|-------|-------------|
| **Go** | `**/*.go` | Backend server, API, auth, storage, probe engine |
| **TypeScript** | `web/src/**/*.ts`, `web/src/**/*.tsx` | React dashboard |
| **YAML** | `configs/*.yaml`, `deployments/**/*.yaml` | Configuration, K8s manifests |
| **JSON** | `go.sum`, `package.json`, `buf.gen.yaml` | Build configs, lock files |

**Phase 2 Skill Activation Recommendations:**

| Language | Recommended Scanners |
|----------|---------------------|
| Go | `scan-go`, `scan-supply-chain`, `scan-config` |
| TypeScript/React | `scan-code`, `scan-xss`, `scan-dependencies` |
| YAML/K8s | `scan-iac` (for Helm/K8s manifests) |
| General | `scan-auth`, `scan-idor`, `scan-cors`, `scan-ratelimit` |

---

## Appendix A: Key File Locations

| Component | Primary File |
|-----------|-------------|
| REST API | `internal/api/rest.go` |
| gRPC API | `internal/grpcapi/server.go` |
| WebSocket | `internal/api/websocket.go` |
| Auth (local) | `internal/auth/local.go` |
| Auth (OIDC) | `internal/auth/oidc.go` |
| Auth (LDAP) | `internal/auth/ldap.go` |
| Config | `internal/core/config.go` |
| Storage | `internal/storage/engine.go` |
| Cluster | `internal/cluster/manager.go` |
| Probe | `internal/probe/engine.go` |
| Alert Manager | `internal/alert/manager.go` |
| Dispatchers | `internal/alert/dispatchers.go` |
| Server Bootstrap | `cmd/anubis/server.go` |
| CLI Entry | `cmd/anubis/main.go` |
| Web Dashboard | `internal/dashboard/` |
| React Source | `web/src/` |

---

## Appendix B: Environment Variables

| Variable | Purpose | Security Impact |
|----------|---------|-----------------|
| `ANUBIS_CONFIG` | Config file path | Low |
| `ANUBIS_DATA_DIR` | Data directory | Medium |
| `ANUBIS_ENCRYPTION_KEY` | Storage encryption | **HIGH** |
| `ANUBIS_CLUSTER_SECRET` | Raft authentication | **HIGH** |
| `ANUBIS_ADMIN_PASSWORD` | Initial admin password | **HIGH** |
| `ANUBIS_LOG_LEVEL` | Logging verbosity | Low |
| `ANUBIS_CORS_ORIGINS` | CORS whitelist override | Medium |
| `OIDC_CLIENT_ID` | OIDC authentication | **HIGH** |
| `OIDC_CLIENT_SECRET` | OIDC authentication | **HIGH** |

---

*End of Phase 1a: Reconnaissance Report*