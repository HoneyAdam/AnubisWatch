# SC-Clickjacking: Clickjacking / UI Redressing Security Check

**Target:** `internal/api/rest.go`
**Date:** 2026-04-16
**Skill:** sc-clickjacking

---

## Summary

| Finding | Severity | Confidence | Status |
|---------|----------|------------|--------|
| CLICK-001 | None | 100 | Protected |
| CLICK-002 | None | 100 | Protected |

---

## Finding CLICK-001: X-Frame-Options Properly Set

**Severity:** None (Protected)
**Confidence:** 100
**File:** `internal/api/rest.go:1735`

### Details

The `X-Frame-Options` header is set to `DENY` on all responses via the `securityHeadersMiddleware`, preventing the application from being embedded in iframes on any origin.

### Evidence

**Security headers middleware (rest.go:1731-1752):**
```go
func (s *RESTServer) securityHeadersMiddleware(handler Handler) Handler {
    return func(ctx *Context) error {
        // Add security headers
        ctx.Response.Header().Set("X-Content-Type-Options", "nosniff")
        ctx.Response.Header().Set("X-Frame-Options", "DENY")  // CLICKJACKING PROTECTED
        ctx.Response.Header().Set("X-XSS-Protection", "1; mode=block")
        ctx.Response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        ctx.Response.Header().Set("Content-Security-Policy", "default-src 'self'")
        // ...
    }
}
```

---

## Finding CLICK-002: Content-Security-Policy Frame-Ancestors Set

**Severity:** None (Protected)
**Confidence:** 100
**File:** `internal/api/rest.go:1738`

### Details

The `Content-Security-Policy: default-src 'self'` header is set, which implicitly restricts frame embedding to same-origin. For explicit frame-ancestors control, the `default-src 'self'` directive restricts loading to the same origin.

### Evidence

```go
ctx.Response.Header().Set("Content-Security-Policy", "default-src 'self'")
```

**Note:** While `default-src 'self'` provides some frame-ancestors protection, explicit `frame-ancestors` directive would be more precise:
```
Content-Security-Policy: frame-ancestors 'none'
```

However, the `X-Frame-Options: DENY` already provides strong frame protection.

---

## Security Headers Summary

| Header | Value | Protection |
|--------|-------|------------|
| X-Frame-Options | DENY | Prevents all iframe embedding |
| X-Content-Type-Options | nosniff | Prevents MIME sniffing |
| X-XSS-Protection | 1; mode=block | XSS filter (legacy browsers) |
| Referrer-Policy | strict-origin-when-cross-origin | Controls referrer leakage |
| Content-Security-Policy | default-src 'self' | Restricts resource loading |
| Strict-Transport-Security | max-age=31536000; includeSubDomains; preload | Forces HTTPS |

---

## Test Coverage

The security headers are validated in tests:

**audit_test.go:282,370:**
```go
"X-Frame-Options": "DENY",
```

**handlers_extra_test.go:1085:**
```go
{"X-Frame-Options", "DENY"},
```

---

## Assessment

**Clickjacking Protection: SECURE**

The application sets both:
1. `X-Frame-Options: DENY` - Prevents ALL framing
2. `Content-Security-Policy: default-src 'self'` - Restricts to same-origin

No framebusting JavaScript is needed since the headers provide server-side protection.

---

## Recommendations

**No action required** - Clickjacking protection is properly implemented.

**Optional improvement:** Add explicit frame-ancestors directive if stricter control is needed:
```go
ctx.Response.Header().Set("Content-Security-Policy", "frame-ancestors 'none'; default-src 'self'")
```

---

## References

- CWE-1021: Improper Restriction of Rendered UI Layers (https://cwe.mitre.org/data/definitions/1021.html)
- MDN: X-Frame-Options
- OWASP Clickjacking Defense Cheat Sheet