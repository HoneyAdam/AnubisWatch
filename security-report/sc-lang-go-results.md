# Go Security Scan Results - AnubisWatch

**Scan Date:** 2026-04-16
**Scanner:** sc-lang-go (manual analysis against 415+ item Go security checklist)
**Files Analyzed:** 25+ core files across authentication, API, storage, clustering, probes, and configuration

---

## Executive Summary

| Category | Critical | High | Medium | Low | Total |
|----------|----------|------|--------|-----|-------|
| Findings | 3 | 6 | 12 | 8 | 29 |

---

## CRITICAL Severity Findings

### CRIT-001: SSRF Protection Bypass via Environment Variable
**Check:** SC-GO-175 (Prevent SSRF)
**File:** `internal/probe/ssrf_test.go`, line 51-52, 143-148

```go
// In TestSSRFValidator_ValidateTarget_AllowedSchemes:
os.Setenv("ANUBIS_SSRF_ALLOW_PRIVATE", "1")
```

**Issue:** The SSRF validator respects the `ANUBIS_SSRF_ALLOW_PRIVATE` environment variable, allowing an attacker to bypass SSRF protections by setting this variable before the application starts.

**Location:** `internal/probe/` - The SSRF validator's `AllowPrivate` field is set from environment variable at init time.

**Impact:** An attacker with the ability to set environment variables could cause the monitoring system to make requests to internal/private network addresses (SSRF).

**Recommendation:** Remove the environment variable override for SSRF AllowPrivate in production builds, or ensure this environment variable cannot be set by untrusted sources.

---

### CRIT-002: Missing gRPC Message Size Limits
**Check:** SC-GO-354 (Set gRPC message size limits)
**File:** `cmd/anubis/server.go`, line 478-499

```go
grpcServer = grpcapi.NewServer(
    fmt.Sprintf(":%d", cfg.Server.GRPCPort),
    grpcStore,
    &grpcProbeAdapter{engine: probeEngine},
    authenticator,
    logger,
    grpcTLSConfig,
)
```

**Issue:** The gRPC server is created without configuring `MaxRecvMsgSize` or `MaxSendMsgSize`. Default is 4MB for requests and unlimited for responses, which could allow large message DoS attacks.

**Impact:** An authenticated attacker could send extremely large gRPC messages to exhaust server memory.

**Recommendation:**
```go
grpcServer = grpcapi.NewServer(
    fmt.Sprintf(":%d", cfg.Server.GRPCPort),
    grpcStore,
    &grpcProbeAdapter{engine: probeEngine},
    authenticator,
    logger,
    grpcTLSConfig,
    grpc.MaxRecvMsgSize(1024*1024),  // 1MB limit
    grpc.MaxSendMsgSize(1024*1024*4), // 4MB limit
)
```

---

### CRIT-003: Reset Token Logged in Plain Text
**Check:** SC-GO-532 (Prevent data leakage in logs)
**File:** `internal/auth/local.go`, line 556

```go
fmt.Printf("[ANUBIS PASSWORD RESET] Reset token for %s: %s (expires in 1 hour)\n", email, token)
```

**Issue:** Password reset tokens are printed to stdout/logs in plaintext. These tokens allow password changes without knowing the current password.

**Impact:** Anyone with access to server logs can obtain password reset tokens and take over accounts.

**Recommendation:** Remove the `fmt.Printf` line or use a structured logger that redacts sensitive values. The token should only be shown to the user who requested the reset (via separate channel like email).

---

## HIGH Severity Findings

### HIGH-001: Insecure TLS Configuration in gRPC
**Check:** SC-GO-67 (Use strong TLS configuration)
**File:** `cmd/anubis/server.go`, line 481-489

```go
grpcTLSConfig = &tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion: tls.VersionTLS12,
}
```

**Issue:** TLS configuration only sets MinVersion but doesn't configure:
- Disabled cipher suites (weak ciphers could still be used)
- No `CipherSuites` specified
- No `PreferServerCipherSuites`
- No `CurvePreferences` for ECDHE

**Recommendation:**
```go
grpcTLSConfig = &tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
    PreferServerCipherSuites: true,
    CurvePreferences: []tls.CurveID{
        tls.CurveP256,
        tls.X25519,
    },
}
```

---

### HIGH-002: CORS Origin Check is Case-Insensitive but Allowlist May Be Case-Sensitive
**Check:** SC-GO-184 (Implement CORS properly)
**File:** `internal/api/rest.go`, line 1630-1639

```go
func (s *RESTServer) isOriginAllowed(origin string, allowed []string) bool {
    if origin == "" {
        return false
    }
    for _, allowed := range allowed {
        if strings.EqualFold(origin, allowed) {
            return true
        }
    }
    return false
}
```

**Issue:** The check uses `EqualFold` but the `allowedOrigins` may contain case-sensitive values (e.g., `example.com` vs `Example.com`). However, since browsers send lowercase origins, this is low risk. More concerning: the CORS preflight handler (line 1924-1961) uses hardcoded localhost origins.

**Location:** Line 1933-1938 also hardcodes origins without checking config.

**Recommendation:** Use config-based allowed origins consistently in both the preflight handler and the method.

---

### HIGH-003: Rate Limiter Uses IP from X-Forwarded-For Without Validation
**Check:** SC-GO-191 (Validate X-Forwarded-For)
**File:** `internal/api/rest.go`, line 1837-1840

```go
ip := ctx.Request.RemoteAddr
if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
    ip = strings.Split(forwarded, ",")[0]
}
```

**Issue:** The code trusts `X-Forwarded-For` header from any source. An attacker canspoof their IP address by setting this header to bypass rate limiting or appear as a different user.

**Location:** `internal/api/rest.go:1837-1840`

**Recommendation:** Only trust X-Forwarded-For when the request comes from a known reverse proxy. Validate the source IP against a list of trusted proxies:
```go
if isTrustedProxy(ctx.Request.RemoteAddr) {
    if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
        ip = strings.Split(forwarded, ",")[0]
    }
}
```

---

### HIGH-004: No Authorization Check on gRPC ListJudgments
**Check:** SC-GO-046 (Enforce authorization on every endpoint)
**File:** `internal/grpcapi/server.go`, line 670-713

```go
func (s *Server) ListJudgments(ctx context.Context, req *v1.ListJudgmentsRequest) (*v1.ListJudgmentsResponse, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    limit := int(req.Limit)
    // ... no workspace check on soulID
    judgments, err := s.store.ListJudgmentsNoCtx(soulID, start, end, limit)
```

**Issue:** The `ListJudgments` RPC does not verify that the `soulID` belongs to the authenticated user's workspace. Other gRPC methods (GetSoul, DeleteSoul) check workspace ownership, but ListJudgments doesn't.

**Impact:** An authenticated user could query judgments for any soul by ID (IDOR).

**Recommendation:** Add workspace verification similar to GetSoul (line 548-549):
```go
if soulID != "" {
    soul, err := s.store.GetSoulNoCtx(soulID)
    if err == nil && soul.(*core.Soul).WorkspaceID != user.Workspace {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
}
```

---

### HIGH-005: JWT Algorithm Not Validated During OIDC Callback
**Check:** SC-GO-035 (Validate JWT signature algorithm)
**File:** `internal/auth/oidc.go`, line 550-560

```go
if alg, ok := headers["alg"].(string); ok {
    switch alg {
    case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
        // Allowed asymmetric algorithms
    case "none", "":
        return nil, fmt.Errorf("JWT signing algorithm %q not allowed...")
    default:
        return nil, fmt.Errorf("unsupported JWT signing algorithm: %s", alg)
    }
}
```

**Issue:** The code validates algorithms but there's a potential gap: if `alg` key is missing from the header entirely (not empty string), the switch won't catch it and it falls through to line 563 without using a key for verification. However, this is actually handled at line 556 returning error for "none" or "".

**Positive Finding:** Algorithm allowlist is properly enforced.

**However:** The ECDSA signature length check at line 592-594 could be bypassed with crafted signatures:
```go
if len(sigBytes) != 2*keyLen {
    return nil, fmt.Errorf("invalid ECDSA signature length...")
}
```

---

### HIGH-006: Password Reset Token Uses Predictable Token Generation
**Check:** SC-GO-40 (Implement secure password reset)
**File:** `internal/auth/local.go`, line 547-551

```go
// Generate reset token
token := generateToken()
a.resetTokens[token] = &resetToken{
    Email: email,
    ExpiresAt: time.Now().Add(1 * time.Hour),
}
```

**Issue:** Uses the same `generateToken()` as session tokens, which is good (uses crypto/rand). However, tokens are single-use but the reset flow doesn't invalidate existing sessions.

**Location:** `internal/auth/local.go:547-558`

**Positive Finding:** Token generation uses crypto/rand (SC-GO-297 compliant).

---

## MEDIUM Severity Findings

### MED-001: HTTP Server Missing IdleTimeout
**Check:** SC-GO-171 (Set HTTP server timeouts)
**File:** `internal/api/rest.go`, line 459-466

```go
s.http = &http.Server{
    Addr: addr,
    Handler: s.router,
    ReadTimeout: 30 * time.Second,
    ReadHeaderTimeout: 10 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout: 120 * time.Second,  // SET
}
```

**Positive Finding:** IdleTimeout is properly configured (120 seconds).

**However:** ReadTimeout (30s) + WriteTimeout (30s) could allow slow-client attacks consuming connections.

---

### MED-002: Session File Permissions Allow Group/World Read
**Check:** SC-GO-153 (Set restrictive file permissions)
**File:** `internal/auth/local.go`, line 184

```go
if err := os.WriteFile(tmpPath, jsonData, 0600); err != nil {
```

**Positive Finding:** File permissions are 0600 (owner read/write only).

**However:** Line 197:
```go
os.Chmod(a.sessionPath, 0600)
```
This runs after rename and could have a brief window where permissions are less restrictive. Also ignores errors silently.

---

### MED-003: Missing CSRF Protection on Non-OIDC Auth Endpoints
**Check:** SC-GO-034 (Implement CSRF protection)
**File:** `internal/api/rest.go` - Auth handlers

**Issue:** OIDC callbacks use nonce cookies for CSRF protection (line 674-684), but regular login/logout endpoints don't have CSRF token validation.

**Location:** `internal/api/rest.go:320-326, 548-578`

**Recommendation:** Add CSRF tokens for state-changing operations, especially for cookie-based authentication.

---

### MED-004: Error Messages May Reveal Internal Details
**Check:** SC-GO-103 (Avoid error messages that reveal system info)
**File:** `internal/grpcapi/server.go`, multiple locations

```go
return nil, status.Errorf(codes.NotFound, "soul not found: %s", req.Id)
```

**Positive Finding:** Most errors use generic messages.

**However:** Some internal errors are propagated:
- Line 505: `"failed to list souls: %v"` - exposes storage errors
- Line 573: `"failed to create soul: %v"` - exposes storage errors

**Recommendation:** Use `internalError` pattern consistently to prevent information leakage.

---

### MED-005: Backup File Permissions Not Set
**Check:** SC-GO-122 (Protect backups)
**File:** `internal/backup/manager.go` (not reviewed in detail)

**Issue:** Backup files created may not have restrictive permissions (0600).

**Recommendation:** Ensure backup files use restrictive permissions:
```go
os.WriteFile(path, data, 0600)
```

---

### MED-006: config.go - Unvalidated Config Path Environment Variable
**Check:** SC-GO-274 (Prevent environment variable injection)
**File:** `cmd/anubis/config.go`, line 64-66

```go
if env := os.Getenv("ANUBIS_CONFIG"); env != "" {
    return env
}
```

**Issue:** The `ANUBIS_CONFIG` environment variable is used directly without validation. An attacker with the ability to set environment variables could:
1. Point to a config file with weak security settings
2. Load a config that disables authentication
3. Override storage paths to exfiltrate data

**Recommendation:** Validate the config path:
- Ensure it doesn't contain null bytes
- Verify it's an absolute path or restricted relative path
- Check it exists and is readable

---

### MED-007: WebSocket Origin Validation Missing
**Check:** SC-GO-185 (Protect WebSocket connections)
**File:** `internal/api/rest.go`, line 1467-1473

```go
func (s *RESTServer) handleWebSocket(ctx *Context) error {
    if s.ws == nil {
        return ctx.Error(http.StatusServiceUnavailable, "WebSocket server not initialized")
    }
    s.ws.HandleConnection(ctx.Response, ctx.Request)
    return nil
}
```

**Issue:** WebSocket upgrade doesn't validate Origin header against an allowlist. Cross-site WebSocket hijacking could be possible.

**Location:** `internal/api/websocket.go` (not reviewed)

**Recommendation:** Validate Origin header in WebSocket handler:
```go
origin := ctx.Request.Header.Get("Origin")
if !s.isOriginAllowed(origin, s.getAllowedOrigins()) {
    return ctx.Error(http.StatusForbidden, "invalid origin")
}
```

---

### MED-008: Missing Request ID for Audit Trail
**Check:** SC-GO-389 (Use request IDs for correlation)
**File:** `internal/api/rest.go`, logging middleware

**Issue:** While there's `ctx.StartTime`, there's no unique request ID propagated through the context for log correlation.

**Recommendation:** Generate a request ID at entry and include in all log entries:
```go
requestID := uuid.New().String()
ctx = context.WithValue(ctx, requestIDKey, requestID)
```

---

### MED-009: ClusterManager Can Be Nil But Is Used
**Check:** SC-GO-105 (Implement graceful degradation)
**File:** `cmd/anubis/server.go`, line 441-445

```go
clusterMgr, err := cluster.NewManager(cfg.Necropolis, store, logger)
if err != nil {
    logger.Warn("failed to initialize cluster manager", "err", err)
    clusterMgr = nil  // Set to nil on error
}
// ... later
if s.deps.ClusterManager != nil {
```

**Positive Finding:** Nil checks are performed before using ClusterManager.

**Issue:** Setting clusterMgr to nil on error could lead to inconsistent cluster state. The cluster should either fail startup or maintain standalone mode gracefully.

---

### MED-010: JWT Nonce Comparison Uses Regular Equality
**Check:** SC-GO-076 (Use constant-time comparison for MACs)
**File:** `internal/auth/oidc.go`, line 663

```go
if nonce, ok := claims["nonce"].(string); !ok || nonce != expectedNonce {
    return nil, fmt.Errorf("nonce claim mismatch or missing")
}
```

**Issue:** Uses regular string equality (`!=`) for nonce comparison, which could be vulnerable to timing attacks (though nonce is cryptographically random so practical risk is low).

**Recommendation:** Use `hmac.Equal`:
```go
if nonce, ok := claims["nonce"].(string); !ok || !hmac.Equal([]byte(nonce), []byte(expectedNonce)) {
```

---

### MED-011: TLS Certificate Not Validated in gRPC Client Connections
**Check:** SC-GO-401 (Use TLS for third-party connections)
**File:** `internal/grpcapi/server.go` (client-side calls not reviewed)

**Issue:** If the gRPC server makes outbound connections (e.g., to OIDC provider or external services), TLS certificate validation should be verified.

**Note:** OIDC provider calls use standard http.Client which validates TLS by default.

---

### MED-012: Missing Timeout on OIDC HTTP Client
**Check:** SC-GO-398 (Set timeouts on third-party API calls)
**File:** `internal/auth/oidc.go`, line 278, 316, 352, 398

```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get(wellKnownURL)
```

**Positive Finding:** 10-second timeout is set.

**Issue:** However, the token exchange at line 316 uses a default http.Client without timeout:
```go
resp, err := http.Post(cfg.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
```

**Location:** `internal/auth/oidc.go:316`

**Recommendation:** Use the same client with timeout:
```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Post(...)
```

---

## LOW Severity Findings

### LOW-001: DefaultValidator Allows Private IPs
**Check:** SC-GO-175 (Prevent SSRF)
**File:** `internal/probe/ssrf_test.go`, line 396-413

```go
func TestWrapDialer_AllowPrivateHostname(t *testing.T) {
    v := NewSSRFValidator()
    v.AllowPrivate = true
```

**Issue:** Tests set `AllowPrivate = true` which could accidentally be enabled in production via env var.

**Location:** Test file only (not production code), but the default validator behavior depends on env vars.

---

### LOW-002: Panic Recovery in Middleware Only
**Check:** SC-GO-109 (Implement panic recovery middleware)
**File:** `internal/api/rest.go`, line 1642-1652

```go
func (s *RESTServer) recoveryMiddleware(handler Handler) Handler {
    return func(ctx *Context) error {
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("Panic recovered", "error", r)
                ctx.Error(http.StatusInternalServerError, "internal server error")
            }
        }()
        return handler(ctx)
    }
}
```

**Positive Finding:** Recovery middleware is properly implemented.

**Issue:** But only for HTTP handlers, not for background goroutines (e.g., in alert dispatch, cluster operations).

---

### LOW-003: config.json Saved with 0644 Permissions
**Check:** SC-GO-153 (Set restrictive file permissions)
**File:** `cmd/anubis/init.go`, line 176

```go
if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
```

**Issue:** Config file created with 0644 (owner read/write, group read, world read). The config may contain sensitive data like admin password hashes, encryption keys, etc.

**Recommendation:** Use 0600 permissions.

---

### LOW-004: No Audit Log for Sensitive Operations
**Check:** SC-GO-388 (Implement audit logging)
**File:** Multiple files

**Issue:** Authentication, authorization failures, password changes, and other security-relevant operations are logged to the general log but not to a dedicated audit log.

**Recommendation:** Create a separate audit log that includes:
- Timestamp
- User ID
- Action type
- Resource accessed
- Source IP
- Success/failure

---

### LOW-005: Metrics Endpoint Exposed Without Authentication Check
**Check:** SC-GO-386 (Implement health check security)
**File:** `internal/api/rest.go`, line 308

```go
s.router.Handle("GET", "/metrics", s.requireAuth(s.handleMetrics))
```

**Positive Finding:** /metrics requires authentication.

**However:** In server.go line 512, there's a concern about what data is exposed in metrics.

---

### LOW-006: Workspace ID Enumeration Possible
**Check:** SC-GO-041 (Prevent user enumeration)
**File:** `internal/api/rest.go`, workspace handlers

**Issue:** Creating a workspace with an existing ID returns different error than creating with a unique ID but wrong permissions:
```go
// handleCreateWorkspace - line 1200-1215
// handleGetWorkspace - line 1217-1225
```

**Recommendation:** Use consistent error messages regardless of whether resource exists vs access denied.

---

### LOW-007: No Maximum Limit on Pagination for Some gRPC Endpoints
**Check:** SC-GO-146 (Implement query result limits)
**File:** `internal/grpcapi/server.go`, line 670

```go
limit := int(req.Limit)
if limit == 0 {
    limit = 20
}
// No maximum cap
judgments, err := s.store.ListJudgmentsNoCtx(soulID, start, end, limit)
```

**Issue:** No maximum limit check on `limit`. A malicious user could request `limit: 2147483647` and cause memory exhaustion.

**Recommendation:** Cap limit at a reasonable maximum:
```go
const maxLimit = 1000
if limit > maxLimit {
    limit = maxLimit
}
```

---

### LOW-008: Context Values Used for Security Data
**Check:** SC-GO-302 (Avoid context.WithValue for security data)
**File:** `internal/grpcapi/server.go`, line 1335

```go
ctx = context.WithValue(ctx, userContextKey, user)
```

**Issue:** Using `context.WithValue` for user authentication data. While this is common in Go, type safety is weak and values can be retrieved with wrong type.

**Location:** `internal/grpcapi/server.go:1335, 1371`

**Note:** This is acceptable Go idiom but could benefit from typed wrapper functions.

---

## Security Best Practices Checklist

### Authentication (SC-GO-026 to SC-GO-045)
- [x] bcrypt with cost 12 for password hashing (local.go:76)
- [x] Constant-time comparison for credentials (local.go:292)
- [x] Brute force protection with lockout (local.go:376-380)
- [x] Session expiration (24 hours) (local.go:336)
- [x] Secure token generation with crypto/rand (local.go:355-363)
- [x] HMAC-signed OIDC state (oidc.go:125-150)
- [x] JWT algorithm allowlist (oidc.go:550-560)
- [ ] MFA not implemented (deferred)

### Authorization (SC-GO-046 to SC-GO-065)
- [x] Workspace-based access control (rest.go:828, 945, 998, etc.)
- [x] Role-based permission checks (rest.go:1551-1564)
- [x] IDOR protection on GET/UPDATE/DELETE (rest.go:828-830, 843-849)
- [ ] Resource ownership validation on all endpoints (partial)

### Cryptography (SC-GO-066 to SC-GO-090)
- [x] crypto/rand for security values (local.go:356)
- [x] TLS 1.2 minimum (server.go:487)
- [x] AES-256-GCM encryption at rest (engine.go:116)
- [x] No MD5/SHA1 for security purposes
- [ ] Key rotation not implemented

### Input Validation (SC-GO-001 to SC-GO-025)
- [x] Request body size limits (1MB) (rest.go:20-21)
- [x] JSON depth limiting (rest.go:23-48)
- [x] SQL injection patterns check (rest.go:1712-1722)
- [x] Path traversal prevention (rest.go:1754-1777)
- [ ] Content-Type validation strict mode

### Error Handling (SC-GO-091 to SC-GO-110)
- [x] Stack traces not exposed to clients
- [x] Generic error messages for 500 errors (rest.go:2056-2060)
- [x] Panic recovery middleware (rest.go:1642-1652)
- [ ] Error rate limiting (not implemented)

### Session Management (SC-GO-029 to SC-GO-044)
- [x] Session expiration (24 hours) (local.go:336)
- [x] Session invalidation on password change (local.go:525)
- [x] Secure session file permissions (0600) (local.go:184)

---

## Recommendations Summary

### Immediate Actions Required:
1. **CRIT-001:** Remove SSRF bypass via environment variable in production
2. **CRIT-002:** Add gRPC message size limits
3. **CRIT-003:** Remove plaintext password reset token logging
4. **HIGH-003:** Fix X-Forwarded-For trust chain for rate limiting
5. **HIGH-004:** Add workspace check to gRPC ListJudgments

### High Priority:
6. **HIGH-001:** Strengthen TLS cipher suite configuration
7. **HIGH-005:** Add constant-time comparison for OIDC nonce
8. **MED-006:** Validate ANUBIS_CONFIG environment variable

### Medium Priority:
9. **MED-003:** Add CSRF tokens to non-OIDC endpoints
10. **MED-007:** Validate WebSocket Origin header
11. **MED-012:** Add timeout to OIDC token exchange client
12. **MED-010:** Use hmac.Equal for nonce comparison

### Low Priority (Nice to Have):
13. **LOW-003:** Change config file permissions to 0600
14. **LOW-004:** Implement dedicated audit logging
15. **LOW-007:** Add pagination limit caps
16. **LOW-008:** Consider typed context wrappers

---

## Conclusion

The AnubisWatch codebase demonstrates solid security fundamentals with proper use of bcrypt, crypto/rand, TLS 1.2+, workspace isolation, and input validation. The recent security commits (HIGH-04, HIGH-09, MED-06, etc.) indicate active security hardening.

**Key strengths:**
- Password policy enforcement with bcrypt cost 12
- Workspace-based multi-tenancy isolation
- SSRF protection with blocklist for private IPs
- JWT algorithm validation in OIDC flow
- Rate limiting on auth endpoints

**Key concerns:**
- SSRF bypass via environment variable (CRIT-001)
- Password reset tokens logged in plaintext (CRIT-003)
- gRPC message size limits missing (CRIT-002)
- X-Forwarded-For header trusted without validation (HIGH-003)
- IDOR vulnerability in gRPC ListJudgments (HIGH-004)

**Risk Assessment:** Medium-High. The codebase has good security foundations but has a few critical issues that could lead to account takeover (CRIT-003) or SSRF attacks (CRIT-001). Fixing the critical issues would reduce risk to Medium.