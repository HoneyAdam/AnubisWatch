# AnubisWatch — Production Readiness Roadmap

> **Version:** 3.1.0
> **Date:** 2026-04-11
> **Based on:** ANALYSIS.md v3.1.0 findings
> **Status:** PRODUCTION READY — Polish and Enhancement Phase

---

## Executive Summary

AnubisWatch has achieved **production readiness (90/100)**. All critical issues are resolved. This roadmap focuses on **polish, enhancements, and enterprise features** rather than fixing blockers.

### Current State
- **Score:** 90/100
- **Test Coverage:** ~84% average
- **TODOs:** 1 remaining
- **Dependencies:** 3 direct (zero-dep goal achieved)
- **CI/CD:** Comprehensive (tests, lint, security, Docker, Helm)

### Goals for v4.0
1. Push coverage to 90%+ average
2. Implement the 1 remaining TODO (region conflict detection)
3. Add enterprise features (SSO, audit logs)
4. Improve CLI test coverage
5. Binary WAL format for storage performance

---

## Phase 1: Quality Polish (Week 1-2) — ~40h

### 1.1 Fix Flaky Test (2h) 🔴
- [ ] Fix `TestCobaltDB_ListJudgments_CorruptData` isolation issue under `-covermode=count`
- **Impact:** Eliminates false CI failures
- **Location:** `internal/storage/engine_test.go`

### 1.2 Improve CLI Coverage (8h)
- [ ] Add tests for `cmd/anubis/main.go` (currently 54.5%)
- [ ] Add tests for `cmd/anubis/server.go`
- [ ] Add tests for `cmd/anubis/init.go`
- **Target:** 80%+ coverage for cmd package
- **Impact:** Confidence in CLI entry point

### 1.3 Implement Region Conflict Detection (8h)
- [ ] Implement conflict detection in `internal/region/replication.go:552`
- [ ] Add last-write-wins or vector clock strategy
- **Impact:** Complete multi-region support
- **Location:** `internal/region/replication.go`

### 1.4 Improve Profiling Coverage (4h)
- [ ] Add tests for profiling handlers (currently 75%)
- **Target:** 90%+ coverage
- **Location:** `internal/profiling/`

### 1.5 Binary WAL Format (8h)
- [ ] Replace JSON-based WAL with binary format
- [ ] Add length-prefixed binary entries
- **Impact:** Better write performance, smaller WAL files
- **Location:** `internal/storage/engine.go:484-564`

### 1.6 Full JSONPath Support (4h)
- [ ] Replace simple string splitting with proper JSONPath parser
- [ ] Support array indexing (`$.items[0].name`)
- **Impact:** Better assertion capabilities
- **Location:** `internal/probe/http.go:383-433`

### 1.7 Persistent Rate Limiter (4h)
- [ ] Persist rate limit state to storage
- [ ] Survive server restarts
- **Impact:** Prevents abuse after restart
- **Location:** `internal/api/rest.go:1217-1340`

### 1.8 B+Tree Underflow Handling (8h)
- [ ] Implement node merging on deletion
- [ ] Handle underflow in B+Tree
- **Impact:** Correct B+Tree behavior for delete-heavy workloads
- **Location:** `internal/storage/engine.go`

---

## Phase 2: Enterprise Features (Week 3-6) — ~120h

### 2.1 SSO / OIDC (16h)
- [ ] Implement OIDC authentication provider
- [ ] Support Google, GitHub, Azure AD
- **Location:** `internal/auth/oidc.go`

### 2.2 LDAP Authentication (12h)
- [ ] Implement LDAP auth provider
- **Location:** `internal/auth/ldap.go`

### 2.3 Audit Logging (16h)
- [ ] Audit log for all API operations
- [ ] User action tracking
- [ ] Immutable audit trail
- **Location:** `internal/audit/`

### 2.4 API Versioning (8h)
- [ ] Implement API versioning strategy
- [ ] Backwards compatibility layer
- **Location:** `internal/api/`

### 2.5 Multi-Tenant Quotas (12h)
- [ ] Enforce workspace-level quotas
- [ ] Rate limits per workspace
- [ ] Storage limits per workspace
- **Location:** `internal/core/`

### 2.6 Webhook Validation (4h)
- [ ] Add actual test notification for webhook channels
- **Location:** `api/rest.go` (handleTestChannel)

### 2.7 Incident Management Enhancement (12h)
- [ ] Full incident lifecycle management
- [ ] Incident timeline
- **Location:** `internal/alert/`

### 2.8 Scheduled Maintenance Windows (8h)
- [ ] Maintenance mode scheduling
- [ ] Auto-suppress alerts during maintenance
- **Location:** `internal/alert/`

### 2.9 Alert Templates (8h)
- [ ] Customizable alert message templates
- [ ] Per-channel formatting
- **Location:** `internal/alert/`

### 2.10 gRPC Reflection (4h)
- [ ] Add gRPC service definitions
- [ ] Protobuf schema files
- **Location:** `internal/api/grpc.go`

### 2.11 Health Check Enhancement (8h)
- [ ] Deep health checks (storage, cluster, database)
- [ ] Dependency health reporting
- **Location:** `internal/api/rest.go` (handleReady)

### 2.12 Dashboard Improvements (12h)
- [ ] Add journey builder UI
- [ ] Add incident management UI
- [ ] Add maintenance calendar
- **Location:** `web/`

---

## Phase 3: Performance & Scale (Week 7-8) — ~60h

### 3.1 Load Testing Framework (16h)
- [ ] Automated load testing suite
- [ ] Baseline performance metrics
- [ ] Regression detection
- **Location:** `internal/probe/load_test.go`

### 3.2 Storage Optimization (12h)
- [ ] Time-series compaction
- [ ] Judgment data compression
- **Location:** `internal/storage/`

### 3.3 Connection Pool Tuning (8h)
- [ ] HTTP transport pool metrics
- [ ] Auto-tune based on load
- **Location:** `internal/probe/http.go`

### 3.4 Raft Performance (12h)
- [ ] Batch AppendEntries
- [ ] Snapshot streaming
- **Location:** `internal/raft/`

### 3.5 Memory Profiling (8h)
- [ ] Memory leak detection
- [ ] Allocation optimization
- **Location:** `internal/profiling/`

### 3.6 CDN Integration (4h)
- [ ] CDN support for status pages
- **Location:** `internal/statuspage/`

---

## Phase 4: Release Preparation (Week 9-10) — ~40h

### 4.1 Documentation (12h)
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Deployment guides
- [ ] Troubleshooting guide

### 4.2 Helm Chart Polish (8h)
- [ ] Helm chart v1.0
- [ ] Values customization
- **Location:** `deploy/helm/anubiswatch/`

### 4.3 Docker Optimization (4h)
- [ ] Multi-stage build optimization
- [ ] Distroless base image
- **Location:** `Dockerfile`

### 4.4 Release Automation (8h)
- [ ] Automated changelog generation
- [ ] Multi-platform binary release
- **Location:** `.github/workflows/release.yml`

### 4.5 Migration Tools (8h)
- [ ] Migration from Uptime Kuma
- [ ] Migration from UptimeRobot

### 4.6 Final Security Audit (4h)
- [ ] Third-party security review
- [ ] Penetration testing

---

## Go/No-Go Decision Points

### Go/No-Go #1: After Phase 1 (Week 2)
- **Criteria:** All technical debt resolved, coverage >85%
- **Decision:** Can we ship v3.5 as production release?

### Go/No-Go #2: After Phase 2 (Week 6)
- **Criteria:** Enterprise features working, tests pass
- **Decision:** Can we ship v4.0 as enterprise release?

### Go/No-Go #3: After Phase 4 (Week 10)
- **Criteria:** All phases complete, security audit passed
- **Decision:** Is v4.0 ready for GA?

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Test isolation issues | Medium | Low | Fix in Phase 1.1 |
| Storage performance at scale | Low | Medium | Binary WAL in Phase 1.5 |
| Enterprise feature scope creep | High | Medium | Strict scope in Phase 2 |
| Raft cluster scaling limits | Low | High | Performance testing in Phase 3 |

---

## Effort Summary

| Phase | Duration | Effort | Priority |
|-------|----------|--------|----------|
| Phase 1: Quality Polish | Week 1-2 | ~40h | 🔴 Critical |
| Phase 2: Enterprise | Week 3-6 | ~120h | 🟡 High |
| Phase 3: Performance | Week 7-8 | ~60h | 🟡 High |
| Phase 4: Release | Week 9-10 | ~40h | 🟢 Medium |
| **Total** | **10 weeks** | **~260h** | |

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| v2.0.0 | 2026-04-08 | Initial roadmap (focused on fixing critical issues) |
| v3.1.0 | 2026-04-11 | Updated — all critical issues fixed, focus shifted to polish and enterprise |

**Document End**
