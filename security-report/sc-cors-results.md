# SC-CORS: CORS Misconfiguration Security Check

**Target:** `internal/api/rest.go`
**Date:** 2026-04-16
**Skill:** sc-cors

---

## Summary

| Finding | Severity | Confidence | Status |
|---------|----------|------------|--------|
| CORS-001 | Medium | 85 | Needs Review |

---

## Finding CORS-001: Reflected Origin with Credentials on SSE Endpoint

**Severity:** Medium
**Confidence:** 85
**File:** `internal/api/rest.go:1476-1489`

### Details

The Server-Sent Events (SSE) endpoint handler sets `Access-Control-Allow-Credentials: true` while reflecting the client-provided `Origin` header. While the origin is validated against an allowlist, the combination of reflected origin + credentials warrants review.

### Evidence

**SSE handler (rest.go:1476-1489):**
```go
func (s *RESTServer) handleSSE(ctx *Context) error {
    // Set SSE headers
    w := ctx.Response
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // CORS headers - use same logic as middleware
    origin := ctx.Request.Header.Get("Origin")
    allowedOrigins := s.getAllowedOrigins()
    if s.isOriginAllowed(origin, allowedOrigins) {
        w.Header().Set("Access-Control-Allow-Origin", origin)  // Reflected
        w.Header().Set("Access-Control-Allow-Credentials", "true")  // Credentials
    }
    // ...
}
```

**Note:** The origin IS validated against an allowlist via `isOriginAllowed()` (line 1630-1640), so arbitrary origin reflection does NOT occur. However, the reflected-origins + credentials pattern is flagged per SKILL.md guidelines.

---

## CORS Middleware Analysis

The main `corsMiddleware` (rest.go:1582-1608) properly validates origins:

```go
func (s *RESTServer) corsMiddleware(handler Handler) Handler {
    return func(ctx *Context) error {
        origin := ctx.Request.Header.Get("Origin")
        allowedOrigins := s.getAllowedOrigins()

        if s.isOriginAllowed(origin, allowedOrigins) {
            ctx.Response.Header().Set("Access-Control-Allow-Origin", origin)
            ctx.Response.Header().Set("Access-Control-Allow-Credentials", "true")
        }
        // ...
    }
}
```

**Origin validation (rest.go:1630-1640):**
```go
func (s *RESTServer) isOriginAllowed(origin string, allowed []string) bool {
    if origin == "" {
        return false
    }
    for _, allowed := range allowed {
        if strings.EqualFold(origin, allowed) {  // Exact match (case-insensitive)
            return true
        }
    }
    return false
}
```

### Allowed Origins Priority (rest.go:1611-1627)

```
1. Config file: s.config.AllowedOrigins
2. Environment: ANUBIS_CORS_ORIGINS
3. Default localhost: http://localhost:3000, http://localhost:8080, etc.
```

---

## Positive Security Controls

| Control | Location | Assessment |
|---------|----------|------------|
| Origin allowlist (not wildcard) | rest.go:1630-1640 | Correctly implements exact-match allowlist |
| Vary: Origin header | rest.go:1594-1595, 1489 | Prevents cache poisoning |
| No null origin acceptance | rest.go:1631-1632 | Empty origin returns false |
| No regex bypass | rest.go:1635 | Uses EqualFold (exact match), not regex |
| OIDC cookies: Secure + HttpOnly + SameSite | rest.go:680-682, 691-693 | Properly secured |
| CORS preflight handling | rest.go:1929-1962 | Properly handles OPTIONS |

---

## CORS Configuration Flow

**Preflight (rest.go:1929-1962):**
```go
if method == "OPTIONS" {
    origin := req.Header.Get("Origin")
    // ... allowlist validation ...
    if originAllowed {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Credentials", "true")
    }
    w.Header().Set("Vary", "Origin")
    // ...
}
```

---

## Assessment

**CORS Configuration is Correctly Implemented.** The reflected-origin pattern is only used when the origin is validated against a strict allowlist. The architecture:

1. Uses exact-match origin validation (no wildcards)
2. Sets Vary: Origin on all responses to prevent cache poisoning
3. Rejects empty origins
4. Does not use regex (no subdomain takeover risk)

**Finding CORS-001 is informational** - flagged only because the reflected-origins + credentials pattern exists, but is safe given the allowlist validation.

---

## Recommendations

1. **No immediate action required** - CORS implementation is secure
2. **Optional:** Consider using `X-Content-Type-Options: nosniff` consistently across all responses (already set in `securityHeadersMiddleware`)
3. **Monitor:** Ensure `ANUBIS_CORS_ORIGINS` or config `AllowedOrigins` is set in production (not relying on localhost defaults)

---

## References

- CWE-942: Permissive Cross-domain Policy (https://cwe.mitre.org/data/definitions/942.html)
- MDN: CORS best practices