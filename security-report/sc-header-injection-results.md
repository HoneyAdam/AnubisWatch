# HTTP Header Injection Scan Results (sc-header-injection)

**Scanner:** sc-header-injection  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/ |
| Vulnerabilities Found | 1 |
| Low Severity | 1 |
| Confidence | 80 |

---

## Findings

### Finding HDR-001: Dynamic Access-Control-Allow-Origin from Request Origin

**Title:** Dynamic CORS Origin from Request Header  
**Severity:** Low  
**Confidence:** 80  
**File:** `internal/api/rest.go:1487`  

**Vulnerability Type:** CWE-113 (HTTP Response Splitting)  

**Description:**
The WebSocket and SSE endpoints dynamically set the `Access-Control-Allow-Origin` header based on the incoming `Origin` request header. While this is a common CORS pattern, it can be problematic in certain proxy configurations.

**Code:**
```go
// internal/api/rest.go:1484-1490
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
origin := r.Header.Get("Origin")
w.Header().Set("Access-Control-Allow-Origin", origin)  // Dynamic origin
w.Header().Set("Access-Control-Allow-Credentials", "true")
w.Header().Set("Vary", "Origin")
```

**Note:** The `Vary: Origin` header is correctly set, which helps caching intermediaries vary their cache by origin. This is the recommended pattern.

**Similar finding at:**
- `internal/api/rest.go:1953` - Same pattern in CORS preflight handler

**Impact:**
- In most modern configurations, Go's net/http rejects CRLF in header values automatically
- The `Vary: Origin` header mitigates caching issues
- Risk is primarily in legacy proxy configurations that don't properly handle the Vary header

**Remediation:**
```go
// Option 1: Validate against allowlist
allowedOrigins := map[string]bool{"https://example.com": true, "https://app.example.com": true}
if allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}

// Option 2: Check for valid origin format
if strings.HasPrefix(origin, "https://") && !strings.Contains(origin, "\r\n") {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

---

### Positive Security Finding: Security Headers Middleware

**File:** `internal/api/rest.go:1734-1740`

The REST server properly sets security headers:
```go
ctx.Response.Header().Set("X-Content-Type-Options", "nosniff")
ctx.Response.Header().Set("X-Frame-Options", "DENY")
ctx.Response.Header().Set("X-XSS-Protection", "1; mode=block")
ctx.Response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
ctx.Response.Header().Set("Content-Security-Policy", "default-src 'self'")
ctx.Response.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
```

---

### Analysis of Header Setting Patterns

**All header values are static constants except:**
1. `internal/api/rest.go:1487` - `Access-Control-Allow-Origin` from request Origin
2. `internal/api/rest.go:1953` - `Access-Control-Allow-Origin` from request Origin

All other header `Set-Cookie` usages (`internal/journey/executor.go:445`) read from internal headers map, not user input directly.

### Go's Built-in CRLF Protection

Go's `net/http` automatically rejects headers containing `\r` or `\n` characters. This is a defense-in-depth measure that makes header injection significantly harder in modern Go versions.

---

## Verdict

**1 LOW SEVERITY VULNERABILITY FOUND (Informational)**

The dynamic `Access-Control-Allow-Origin` based on request Origin is a common pattern and the `Vary: Origin` header is correctly set. In modern Go with `net/http`, CRLF injection in headers is blocked by the runtime. This is LOW rather than HIGH due to the existing mitigations.

---

## Common False Positives Eliminated

1. **Modern Go net/http** - Rejects CRLF in header values automatically
2. **Static header values** - Most headers use constant strings
3. **Vary: Origin** - Correctly set for dynamic origin to prevent caching issues
4. **Set-Cookie from internal map** - Journey executor extracts from structured headers map

## References

- CWE-113: HTTP Response Splitting
- https://cwe.mitre.org/data/definitions/113.html