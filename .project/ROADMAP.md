# Project Roadmap: AnubisWatch

> Based on comprehensive codebase analysis performed on 2026-04-16
> This roadmap prioritizes work needed for v1.0.0 stable release and beyond

## Current State Assessment

AnubisWatch is **production-ready** at v0.1.2 with:
- ✅ 100% of specified features implemented
- ✅ 83.8% test coverage (exceeds 80% threshold)
- ✅ All critical and high severity security issues resolved
- ✅ Comprehensive CI/CD pipeline
- ✅ Single-binary deployment model

**Key Blockers**: None. The project is ready for production deployment.

---

## Phase 1: Pre-Release Hardening (Week 1-2)

### Must-Fix Items
- [ ] **Exclude gRPC generated code from coverage**
  - Files: `internal/grpcapi/v1/*.pb.go`
  - Effort: 30 minutes
  - Add `//go:build ignore` or codecov.yml exclusion

- [ ] **Expand frontend test coverage**
  - Current: 40 tests
  - Target: 60 tests (50% increase)
  - Priority components: Dashboard widgets, error boundaries
  - Effort: 2-3 days

- [ ] **Add E2E test for critical user flow**
  - Flow: Create soul → Wait for judgment → Verify alert
  - Tool: Playwright
  - Effort: 1 day

### Documentation Polish
- [ ] **API versioning strategy document**
  - Location: `docs/adr/adr-009-api-versioning.md`
  - Effort: 2 hours

- [ ] **Production deployment guide**
  - Location: `docs/deployment/production.md`
  - Cover: HA setup, backup automation, monitoring
  - Effort: 1 day

---

## Phase 2: Performance Optimization (Week 3-4)

### Database Performance
- [ ] **Add query result caching**
  - Package: `internal/storage/`
  - Implement: LRU cache for frequent judgment queries
  - TTL: 30 seconds for latest judgments
  - Effort: 2 days

- [ ] **Optimize time-series downsampling**
  - Current: Background goroutine
  - Improvement: Batch multiple soul downsampling
  - Effort: 3 days

- [ ] **Add storage performance metrics**
  - Metrics: Query latency, compaction duration, WAL size
  - Package: `internal/metrics/`
  - Effort: 1 day

### Frontend Performance
- [ ] **Implement virtual scrolling for large soul lists**
  - Library: `@tanstack/react-virtual`
  - Target: Support 1000+ souls smoothly
  - Effort: 2 days

- [ ] **Add service worker for offline dashboard**
  - Cache: Static assets, recent judgments
  - Effort: 2 days

---

## Phase 3: Enterprise Features (Week 5-8)

### Authentication & Authorization
- [ ] **SAML 2.0 support**
  - Location: `internal/auth/saml.go`
  - Use: `crewjam/saml` (first external auth dep)
  - Effort: 1 week

- [ ] **RBAC permissions system**
  - Roles: viewer, operator, admin, superadmin
  - Package: `internal/auth/rbac.go`
  - Effort: 3 days

- [ ] **Audit log export**
  - Format: JSON Lines, CSV
  - Filter: Date range, user, action type
  - Effort: 2 days

### Observability
- [ ] **Structured health check endpoint**
  - Current: Simple HTTP 200
  - Add: Database connectivity, Raft status, probe queue depth
  - Effort: 1 day

- [ ] **OpenTelemetry tracing**
  - Package: `internal/tracing/` (exists but minimal)
  - Spans: API requests, probe checks, alert dispatches
  - Effort: 3 days

- [ ] **Custom metrics endpoint improvements**
  - Add: Soul count by status, alert rate by channel
  - Effort: 1 day

---

## Phase 4: Reliability & Operations (Week 9-12)

### Backup & Disaster Recovery
- [ ] **Automated backup scheduling**
  - Config: Cron expression, retention policy
  - Location: `internal/backup/scheduler.go`
  - Effort: 3 days

- [ ] **Point-in-time recovery**
  - Requires: WAL archival
  - Effort: 1 week

- [ ] **Cross-region replication**
  - Use: Raft for metadata, async for judgments
  - Effort: 2 weeks

### Alerting Reliability
- [ ] **Alert deduplication by fingerprint**
  - Current: Per-rule dedup
  - Improvement: Content-based fingerprinting
  - Effort: 2 days

- [ ] **Alert throttling per destination**
  - Prevent: Email/SMS flooding
  - Effort: 1 day

- [ ] **Dead letter queue for failed alerts**
  - Retry: Exponential backoff
  - Storage: Local file queue
  - Effort: 3 days

---

## Phase 5: User Experience (Week 13-16)

### Dashboard Enhancements
- [ ] **Dark/light mode toggle**
  - Location: Theme provider in `web/src/`
  - Effort: 2 days

- [ ] **Custom dashboard templates**
  - Presets: API monitoring, infrastructure, synthetic
  - Effort: 3 days

- [ ] **Real-time log viewer**
  - WebSocket stream of judgments
  - Filter: Soul, status, time range
  - Effort: 3 days

### Mobile Experience
- [ ] **Responsive design audit**
  - Target: Full functionality on mobile
  - Effort: 2 days

- [ ] **Touch-optimized status page**
  - Location: `internal/statuspage/`
  - Effort: 2 days

---

## Phase 6: Integrations (Week 17-20)

### Monitoring Integrations
- [ ] **Prometheus remote write**
  - Export: Judgments as metrics
  - Effort: 3 days

- [ ] **Datadog integration**
  - Metrics: Custom metrics API
  - Events: Alert webhooks
  - Effort: 2 days

### Notification Channels
- [ ] **Microsoft Teams webhook**
  - Location: `internal/alert/teams.go`
  - Effort: 1 day

- [ ] **AWS SNS integration**
  - Location: `internal/alert/sns.go`
  - Effort: 2 days

- [ ] **Generic webhooks with templates**
  - Template engine: Go templates or JSONPath
  - Effort: 3 days

---

## Phase 7: Security Hardening (Week 21-24)

### Advanced Security
- [ ] **Secret rotation automation**
  - Encryption keys, API tokens
  - Zero-downtime rotation
  - Effort: 1 week

- [ ] **Certificate transparency log monitoring**
  - Alert: Unauthorized certificates for monitored domains
  - Effort: 3 days

- [ ] **Security scanning automation**
  - Tool: Trivy, Nancy, gosec in CI
  - Currently: ✅ Already implemented
  - Improvement: Fail build on HIGH+ findings
  - Effort: 1 day

### Compliance
- [ ] **SOC 2 Type II preparation**
  - Document: Controls, evidence collection
  - Effort: 2 weeks (documentation)

- [ ] **GDPR data export/deletion**
  - Feature: User data export, right to erasure
  - Effort: 1 week

---

## Phase 8: Release Preparation (Week 25-26)

### Release v1.0.0
- [ ] **Final security audit**
  - External penetration testing
  - Effort: 1 week (external)

- [ ] **Performance benchmarking**
  - Load: 10K souls, 100 concurrent
  - Report: Published in docs/
  - Effort: 3 days

- [ ] **Documentation freeze**
  - Review: All docs for accuracy
  - Effort: 2 days

- [ ] **Release notes**
  - Location: `RELEASE_NOTES_v1.0.0.md`
  - Effort: 1 day

---

## Beyond v1.0: Future Enhancements

### Phase 9: AI-Powered Features (Q3 2026)
- [ ] **Anomaly detection**
  - ML-based alert threshold recommendation
  - Baseline learning from historical data

- [ ] **Incident correlation**
  - Group related alerts into incidents
  - Root cause suggestion

- [ ] **Natural language queries**
  - "Show me all failing souls in us-east"
  - MCP expansion

### Phase 10: Global Scale (Q4 2026)
- [ ] **Federation**
  - Multi-cluster coordination
  - Global status aggregation

- [ ] **Edge deployment**
  - Lightweight probe-only mode
  - IoT/small device support

### Phase 11: Ecosystem (2027)
- [ ] **Plugin system**
  - Custom checkers via WebAssembly
  - Custom alert channels

- [ ] **Terraform provider**
  - Infrastructure as code for souls

- [ ] **CLI plugins**
  - Extensible command system

---

## Effort Summary

| Phase | Duration | Effort | Priority | Dependencies |
|-------|----------|--------|----------|--------------|
| Phase 1 | Week 1-2 | 1 week | 🔴 Critical | None |
| Phase 2 | Week 3-4 | 1.5 weeks | 🟡 High | Phase 1 |
| Phase 3 | Week 5-8 | 3 weeks | 🟡 High | None |
| Phase 4 | Week 9-12 | 4 weeks | 🟡 High | Phase 2 |
| Phase 5 | Week 13-16 | 2 weeks | 🟢 Medium | None |
| Phase 6 | Week 17-20 | 2 weeks | 🟢 Medium | None |
| Phase 7 | Week 21-24 | 2 weeks | 🟡 High | Phase 3 |
| Phase 8 | Week 25-26 | 1 week | 🔴 Critical | All above |
| **Total** | **26 weeks** | **~16.5 person-weeks** | | |

**Minimum to v1.0.0**: Phase 1 + Phase 8 = **2 weeks**
**Full v1.0.0**: All phases = **6 months**

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| SAML integration complexity | Medium | Medium | Spike implementation first |
| Cross-region replication bugs | Medium | High | Extensive chaos testing |
| Mobile responsiveness issues | Low | Low | Early testing in Phase 5 |
| Dependency update breaks build | Low | Medium | Pin versions, CI matrix |
| Performance degradation | Low | High | Benchmark tests in CI |

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-16 | Target v1.0.0 in 2 weeks (minimal) | Current state is production-ready |
| 2026-04-16 | Defer SAML to Phase 3 | Enterprise feature, not blocking |
| 2026-04-16 | Prioritize observability | Operations-critical for production |

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| v0.1.2 FINAL | 2026-04-14 | Phase 9 complete, all tests passing |
| v1.0.0-RC | 2026-04-16 | **New** — Comprehensive roadmap for stable release |

---

*The Road to Judgment* ⚖️
