# sc-session: Session Management — Results

**Skill:** sc-session | **Severity:** Mixed | **Confidence:** High
**Scan Date:** 2026-04-16 | **Files Scanned:** internal/auth/, internal/api/
**Focus:** internal/auth/ for session handling, token storage; internal/api/rest.go for cookie configuration

---

## Summary: 1 MEDIUM FINDING, 3 POSITIVE PRACTICES

### Findings

#### SESS-001: Bearer Token Sent Via Authorization Header (No HttpOnly Cookie)
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/auth/local.go:333-343`, `internal/api/rest.go:569-577`
- **Vulnerability Type:** CWE-614 (Sensitive Cookie Without Secure Flag) / CWE-352 (CSRF)
- **Description:** Authentication tokens are returned in the JSON response body (`{"token": "..."}`) and passed back to the server via the `Authorization: Bearer <token>` header. This is standard practice for API authentication and is NOT inherently vulnerable. However, the token is vulnerable to:
  1. **CSRF attacks**: Since no cookie is used, traditional CSRF is not applicable. However, any cross-origin request that can set Authorization headers (e.g., via `fetch` with custom headers) could leak the token.
  2. **Token logging**: The token may appear in server access logs if using a reverse proxy that logs Authorization headers.
  3. **Browser history**: If the token is ever passed via URL query parameters (HIGH-03 was fixed to block this on WebSocket), it could appear in browser history.
- **Code (login response):**
  ```go
  // internal/api/rest.go:563-566
  return ctx.JSON(http.StatusOK, map[string]interface{}{
      "user": user,
      "token": token,  // Token in response body
  })
  ```
- **Code (token in Authorization header):**
  ```go
  // internal/api/rest.go:573-577
  token := ctx.Request.Header.Get("Authorization")
  token = strings.TrimPrefix(token, "Bearer ")
  if err := s.auth.Logout(token); err != nil { ... }
  ```
- **Impact:** The Bearer token approach is architecturally sound for API clients. The risk is that the token can be read by JavaScript (XSS) if any XSS vector exists, since it's not stored in an HttpOnly cookie. However, modern SPAs commonly use localStorage/sessionStorage for Bearer tokens.
- **Remediation:** The current approach is acceptable for API-based authentication. If browser-based XSS is a concern, consider:
  1. Using `httpOnly` cookies for token storage (requires SameSite=Strict + Secure)
  2. Implementing CSRF tokens for state-changing operations
  3. Ensuring `Content-Security-Policy` headers prevent XSS (already set in rest.go:1738)
- **Note:** The WebSocket endpoint explicitly blocks token-in-query-parameters (HIGH-03 fix at `internal/api/websocket.go:98`). OIDC cookies (`oidc_nonce`, `oidc_state`) properly use `HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode`.
- **References:** https://cwe.mitre.org/data/definitions/614.html, https://cwe.mitre.org/data/definitions/352.html

---

## Positive Security Practices Verified

### 1. Cryptographically Secure Token Generation
`internal/auth/local.go:355-364`:
```go
func generateToken() string {
    b := make([]byte, 32)  // 256 bits of entropy
    if _, err := rand.Read(b); err != nil {
        panic("CSPRNG failure: cannot generate secure token: " + err.Error())
    }
    return hex.EncodeToString(b)
}
```
- Uses `crypto/rand` (CSPRNG)
- 32 bytes = 256-bit entropy
- Fail-closed panic on CSPRNG failure
- Same implementation in `oidc.go:259` and `ldap.go:48`

### 2. Session Expiration
`internal/auth/local.go:336` — Sessions expire after 24 hours:
```go
ExpiresAt: time.Now().Add(24 * time.Hour),
```
Expired sessions are cleaned up on access (`local.go:261-265`) and by a background goroutine running every 5 minutes (`local.go:213`).

### 3. Session Invalidation on Logout
`internal/auth/local.go:346-353`:
```go
func (a *LocalAuthenticator) Logout(token string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    delete(a.tokens, token)
    a.saveSessionsLocked()
    return nil
}
```
Properly removes session from memory and persists the change.

### 4. OIDC Cookie Security
`internal/api/rest.go:676-695` — OIDC state and nonce cookies use all security attributes:
```go
http.SetCookie(ctx.Response, &http.Cookie{
    Name: "oidc_nonce",
    Value: nonce,
    Path: "/",
    HttpOnly: true,
    Secure: true,
    SameSite: http.SameSiteStrictMode,
    MaxAge: 600,
})
```
Same for `oidc_state` cookie. Cookies are properly cleared after use (`MaxAge: -1`).

### 5. Session Storage Atomicity
`internal/auth/local.go:182-198` — Sessions are written atomically using a temp file + rename:
```go
tmpPath := a.sessionPath + ".tmp"
os.WriteFile(tmpPath, jsonData, 0600)
os.Chmod(tmpPath, 0600)
os.Rename(tmpPath, a.sessionPath)
os.Chmod(a.sessionPath, 0600)
```
Prevents corruption from partial writes on crash.

### 6. Server-Side Session Store
All sessions are stored server-side in memory (with optional disk persistence). The token alone is insufficient — an attacker cannot forge a session without the server's memory state. No session fixation is possible since a new token is generated on each login.

### 7. No Token-in-URL on WebSocket
`internal/api/websocket.go:98` — WebSocket connections reject tokens passed as query parameters:
```go
if r.URL.Query().Get("token") != "" {
    s.logger.Warn("WebSocket connection rejected: token via query parameter is not allowed")
}
```

### 8. Password Change Invalidates All Sessions
`internal/auth/local.go:524-526` — On password change, all sessions are invalidated.

---

## Verdict

The session management is well-implemented. The Bearer token approach for API authentication is standard and acceptable. The most significant strength is server-side session storage with cryptographically secure tokens generated via CSPRNG, and proper OIDC cookie security attributes. No session fixation, predictable tokens, or improper session expiration was found.

---