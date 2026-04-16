# SC-WebSocket: WebSocket Security Check

**Target:** `internal/api/websocket.go`
**Date:** 2026-04-16
**Skill:** sc-websocket

---

## Summary

| Finding | Severity | Confidence | Status |
|---------|----------|------------|--------|
| WS-001 | None | 100 | Protected |
| WS-002 | None | 100 | Protected |
| WS-003 | None | 100 | Protected |
| WS-004 | Low | 85 | Informational |

---

## Finding WS-001: Bearer Token Authentication Required

**Severity:** None (Protected)
**Confidence:** 100
**File:** `internal/api/websocket.go:94-128`

### Details

WebSocket connections require Bearer token authentication via the `Authorization` header. Query parameter tokens are explicitly rejected to prevent token leakage.

### Evidence

**Query param token rejected (websocket.go:97-103):**
```go
// SECURITY: Reject query parameter tokens to prevent token leakage in
// access logs, browser history, and Referer headers. (HIGH-03 fix)
if r.URL.Query().Get("token") != "" {
    s.logger.Warn("WebSocket connection rejected: token via query parameter is not allowed",
        "remote_addr", r.RemoteAddr)
    http.Error(w, "Unauthorized: token must be provided via Authorization header, not query parameter", http.StatusUnauthorized)
    return
}
```

**Bearer token required (websocket.go:105-128):**
```go
authHeader := r.Header.Get("Authorization")
if !strings.HasPrefix(authHeader, "Bearer ") {
    s.logger.Warn("WebSocket connection rejected: missing Bearer token", "remote_addr", r.RemoteAddr)
    http.Error(w, "Unauthorized: missing Bearer token in Authorization header", http.StatusUnauthorized)
    return
}
token := strings.TrimPrefix(authHeader, "Bearer ")

// Validate token
if token == "" {
    s.logger.Warn("WebSocket connection rejected: empty token", "remote_addr", r.RemoteAddr)
    http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
    return
}

// Authenticate the token
user, err := s.authenticator.Authenticate(token)
if err != nil {
    s.logger.Warn("WebSocket connection rejected: invalid token", ...)
    http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
    return
}
```

---

## Finding WS-002: Origin Validation via OriginPatterns

**Severity:** None (Protected)
**Confidence:** 100
**File:** `internal/api/websocket.go:139-147`

### Details

The `coder/websocket` library uses `OriginPatterns` for origin validation on WebSocket upgrade.

### Evidence

```go
// Upgrade HTTP to WebSocket with origin checking
opts := &websocket.AcceptOptions{
    OriginPatterns: s.allowedOrigins,  // Origin validation
}
conn, err := websocket.Accept(w, r, opts)
```

**Allowed origins initialization (websocket.go:43-61):**
```go
func NewWebSocketServer(logger *slog.Logger, authenticator Authenticator, allowedOrigins []string) *WebSocketServer {
    if len(allowedOrigins) == 0 {
        // Default origins for development
        allowedOrigins = []string{
            "http://localhost:3000",
            "http://localhost:8080",
            "http://127.0.0.1:3000",
            "http://127.0.0.1:8080",
        }
    }
    // ...
}
```

**Priority order for allowed origins:**
1. Config file: `s.config.AllowedOrigins` (rest.go:1613-1614)
2. Environment: `ANUBIS_CORS_ORIGINS` (rest.go:1617-1618)
3. Default localhost (websocket.go:46-50)

---

## Finding WS-003: Message Size Limit Enforced

**Severity:** None (Protected)
**Confidence:** 100
**File:** `internal/api/websocket.go:238`

### Details

A 512KB maximum message size limit prevents DoS via large messages.

### Evidence

```go
func (c *WSClient) readPump(ctx context.Context) {
    defer func() {
        c.server.removeClient(c.ID)
        c.Conn.CloseNow()
    }()

    c.Conn.SetReadLimit(512 * 1024) // 512KB max message size  // DOS PROTECTION
    // ...
}
```

---

## Finding WS-004: Workspace Switching Forbidden

**Severity:** Low (Informational)
**Confidence:** 85
**File:** `internal/api/websocket.go:292-294`

### Details

The `join_workspace` message type explicitly rejects workspace switching, binding users to their authenticated workspace.

### Evidence

```go
case "join_workspace":
    // Reject workspace switching - users are bound to their authenticated workspace
    c.send <- c.createErrorMessage("forbidden", "workspace switching not supported")
```

---

## Additional Security Controls

| Control | Location | Assessment |
|---------|----------|------------|
| Token via Authorization header only | websocket.go:97-110 | Prevents token leakage |
| Read limit (512KB) | websocket.go:238 | DoS protection |
| Message type validation | websocket.go:269-298 | Rejects unknown types |
| Read timeout (60s) | websocket.go:241 | Prevents slow-read DoS |
| Write timeout (10s) | websocket.go:318-320 | Prevents blocking writes |
| Ping/pong heartbeat (30s) | websocket.go:303-333 | Connection liveness |
| Per-client send buffer (256) | websocket.go:159 | Backpressure handling |

---

## Architecture Summary

```
WebSocket Connection Flow:
1. Reject query-param tokens (prevents log leakage)
2. Require Bearer Authorization header
3. Validate token via Authenticator.Authenticate()
4. Get user's workspace from authenticated user
5. Check Origin against allowedOrigins (OriginPatterns)
6. Upgrade to WebSocket
7. Subscribe to workspace room
8. Handle only allowed message types (subscribe/unsubscribe/ping)
```

---

## Assessment

**WebSocket Security: SECURE**

| Vulnerability | Status |
|---------------|--------|
| Cross-site WebSocket hijacking | Protected - OriginPatterns + Bearer auth |
| Missing authentication | Protected - Token required |
| Token leakage via URL | Protected - Query params rejected |
| Message injection | Protected - JSON validation on allowed types |
| Rate limiting | Partial - Per-client buffer (256), read limits |
| Unencrypted (ws://) | Config-dependent - Should use wss:// in production |

---

## Recommendations

1. **Ensure TLS in production** - WebSocket should use `wss://` not `ws://` in production
2. **Configure allowed origins** - Set `ANUBIS_CORS_ORIGINS` or config `AllowedOrigins` in production (not localhost defaults)
3. **No immediate action required** - Security controls are properly implemented

---

## References

- CWE-1385: Missing Origin Validation in WebSocket (https://cwe.mitre.org/data/definitions/1385.html)
- CWE-306: Missing Authentication (https://cwe.mitre.org/data/definitions/306.html)
- OWASP WebSocket Security Cheat Sheet