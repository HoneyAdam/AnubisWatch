# Cross-Site Scripting (XSS) Scan Results (sc-xss)

**Scanner:** sc-xss  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/, internal/dashboard/ |
| Vulnerabilities Found | 1 |
| Medium Severity | 1 |
| Confidence | Medium |

---

## Findings

### Finding XSS-001: Weak X-Frame-Options Header

**Title:** Clickjacking Risk via X-Frame-Options ALLOWALL
**Severity:** Medium
**Confidence:** 85
**File:** `internal/statuspage/handler.go:1010`

**Vulnerability Type:** CWE-79 (Cross-site Scripting)

**Description:**
The status page handler sets `X-Frame-Options: ALLOWALL` on HTML responses, which removes clickjacking protection. While this alone is not XSS, it enables clickjacking attacks where the status page can be embedded in an iframe and trick users into performing actions.

**Code:**
```go
// internal/statuspage/handler.go:1007-1010
w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Header().Set("Cache-Control", "public, max-age=60")
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("X-Frame-Options", "ALLOWALL")
```

**Impact:**
- Clickjacking attacks on status page
- UI redress attacks
- Phishing via embedded iframe

**Remediation:**
```go
// Replace ALLOWALL with DENY or SAMEORIGIN
w.Header().Set("X-Frame-Options", "DENY")
// Or use Content-Security-Policy frame-ancestors directive
w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
```

**References:**
- https://cwe.mitre.org/data/definitions/79.html
- https://owasp.org/Top10/A03_2021-Injection/

---

## Positive Security Findings

### Status Page HTML Output Uses html.EscapeString

**File:** `internal/statuspage/handler.go:271`
```go
html.EscapeString(page.Name),  // Properly escapes user-controlled page name
```

The status page HTML rendering uses `html.EscapeString()` for the page name, preventing reflected XSS.

### React Dashboard (web/src)

- No `dangerouslySetInnerHTML` usage found
- No `innerHTML` assignment found in source code
- React 19 with TypeScript provides auto-escaping in JSX expressions
- All dashboard code uses proper React patterns with TypeScript

### Security Headers Middleware

**File:** `internal/api/rest.go:1734-1738`
```go
ctx.Response.Header().Set("X-Content-Type-Options", "nosniff")
ctx.Response.Header().Set("X-Frame-Options", "DENY")  // Note: rest.go uses DENY
ctx.Response.Header().Set("X-XSS-Protection", "1; mode=block")
ctx.Response.Header().Set("Content-Security-Policy", "default-src 'self'")
```

The main API server sets proper XSS protection headers. However, the status page handler at `internal/statuspage/handler.go` overrides `X-Frame-Options` to `ALLOWALL`.

---

## Verdict

**1 MEDIUM VULNERABILITY FOUND**

The status page handler sets `X-Frame-Options: ALLOWALL` which weakens the clickjacking protection established by the main API security headers. This is not a direct XSS but enables clickjacking attacks.

---

## Recommendations

1. Change `X-Frame-Options: ALLOWALL` to `X-Frame-Options: DENY` or `SAMEORIGIN` in `internal/statuspage/handler.go:1010`
2. Consider adding `Content-Security-Policy: frame-ancestors 'none'` header
3. Audit all endpoints that override security headers

## References

- CWE-79: Cross-site Scripting
- https://owasp.org/Top10/A03_2021-Injection/