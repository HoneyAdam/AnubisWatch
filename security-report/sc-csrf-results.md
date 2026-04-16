# SC-CSRF: Cross-Site Request Forgery Security Check

**Target:** `internal/api/rest.go`
**Date:** 2026-04-16
**Skill:** sc-csrf

---

## Summary

| Finding | Severity | Confidence | Status |
|---------|----------|------------|--------|
| CSRF-001 | Low | 95 | Informational |

---

## Finding CSRF-001: Bearer Token Auth Provides CSRF Immunity

**Severity:** Low
**Confidence:** 95
**File:** `internal/api/rest.go:1533-1548`

### Details

The AnubisWatch REST API uses **Bearer token authentication** (not cookie-based sessions) for all state-changing operations. Per OWASP guidelines, APIs using `Authorization: Bearer <token>` header are **CSRF-immune** because:

1. Browsers do not automatically include the `Authorization` header in cross-origin form submissions
2. Attackers cannot force the victim's browser to send Bearer tokens cross-origin

### Evidence

**Token extraction (rest.go:1533-1534):**
```go
token := ctx.Request.Header.Get("Authorization")
token = strings.TrimPrefix(token, "Bearer ")
```

**Login endpoint returns token (rest.go:548-567):**
```go
func (s *RESTServer) handleLogin(ctx *Context) error {
    // ...
    user, token, err := s.auth.Login(req.Email, req.Password)
    return ctx.JSON(http.StatusOK, map[string]interface{}{
        "user": user,
        "token": token,  // Token returned to client, stored and sent via Authorization header
    })
}
```

**No session cookies for API auth** - The API does not use cookie-based authentication for its endpoints. Tokens are stored client-side (localStorage/sessionStorage) and sent explicitly.

### OIDC Cookie Protection (rest.go:674-750)

For OIDC authentication flows, cookies are properly protected:

```go
// rest.go:676-684 - OIDC nonce cookie
http.SetCookie(ctx.Response, &http.Cookie{
    Name: "oidc_nonce",
    Value: nonce,
    Path: "/",
    HttpOnly: true,
    Secure: true,
    SameSite: http.SameSiteStrictMode,  // Strict CSRF protection
    MaxAge: 600,
})

// rest.go:723-726 - Nonce validation on callback
nonceCookie, err := ctx.Request.Cookie("oidc_nonce")
if err != nil {
    s.logger.Error("OIDC callback missing nonce cookie", "error", err)
    return ctx.Error(http.StatusBadRequest, "missing nonce cookie: possible CSRF attack")
}
```

### Content-Type Validation (rest.go:1654-1699)

The `validateJSONMiddleware` enforces `application/json` Content-Type on POST/PUT/PATCH:

```go
// rest.go:1658-1661
if !strings.HasPrefix(contentType, "application/json") {
    return ctx.Error(http.StatusBadRequest, "Content-Type must be application/json")
}
```

This prevents HTML form-based attacks that would send `application/x-www-form-urlencoded`.

### Why Low (Not None)

- **Informational**: The architecture is CSRF-safe by design, but no explicit CSRF token middleware exists for defense-in-depth
- If cookie-based auth is added in the future without CSRF protection, it would be vulnerable
- The OIDC flow is properly protected with SameSite=Strict + nonce validation

---

## Remediation

No immediate action required. The Bearer token pattern is OWASP-recommended and CSRF-immune.

**Optional defense-in-depth:** Add explicit CSRF token middleware for state-changing API endpoints if cookie-based auth is ever introduced.

---

## References

- CWE-352: Cross-Site Request Forgery (https://cwe.mitre.org/data/definitions/352.html)
- OWASP CSRF Prevention Cheat Sheet