# TypeScript/React Security Analysis Report
**AnubisWatch Dashboard** - `web/src/`

**Date:** 2026-04-16
**Skill:** sc-lang-typescript
**Checklist:** TypeScript Security Checklist (415+ items)

---

## Executive Summary

| Severity | Count |
|----------|-------|
| Critical | 3 |
| High | 8 |
| Medium | 6 |
| Low | 4 |

---

## Critical Findings

### C-1: JWT Stored in localStorage (SC-TS-030)
**Severity:** Critical | **CWE:** CWE-922

**Location:**
- `web/src/api/client.ts:24` - Token read from localStorage on initialization
- `web/src/api/client.ts:29` - Token stored in localStorage via `setToken()`
- `web/src/hooks/useWebSocket.tsx:46` - Token read for WebSocket auth
- `web/src/api/hooks.ts:374` - Token read for auth check

**Issue:** JWT tokens are stored in localStorage, making them vulnerable to XSS attacks. Any script injection can exfiltrate the token.

**Recommendation:** Use httpOnly cookies for JWT storage instead of localStorage. Implement token refresh mechanism with secure cookie settings.

---

### C-2: Client-Side Authorization Only (SC-TS-049)
**Severity:** Critical | **CWE:** CWE-602

**Location:** `web/src/App.tsx:20-37`

```typescript
function ProtectedRoute() {
  const { isAuthenticated, loading } = useAuth()
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}
```

**Issue:** Authorization is performed entirely on the client side by checking if a token exists in localStorage. Any API endpoint accessible to an authenticated user is also accessible to any JavaScript that can read localStorage. There is no server-side session validation on API calls.

**Recommendation:** All protected routes must be validated server-side. Consider implementing proper session-based authentication with httpOnly cookies.

---

### C-3: WebSocket Authentication Token Sent in Plain Text (SC-TS-060)
**Severity:** Critical | **CWE:** CWE-862

**Location:** `web/src/hooks/useWebSocket.tsx:61`

```typescript
ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'auth', token }))
  ws.send(JSON.stringify({ type: 'subscribe', channels: ['judgments', 'alerts', 'status'] }))
}
```

**Issue:** Authentication token is sent in the first WebSocket message after connection. This pattern is vulnerable if the WebSocket upgrade happens on an authenticated path without proper validation during handshake.

**Recommendation:** Implement WebSocket handshake with token in connection URL query parameter or use secure subprotocol authentication.

---

## High Findings

### H-1: API Key is Predictable (SC-TS-079)
**Severity:** High | **CWE:** CWE-330

**Location:** `web/src/pages/Settings.tsx:505`

```typescript
value={user ? `anb_live_${user.id}` : 'Not available'}
```

**Issue:** The API key is constructed from a predictable pattern (`anb_live_` + user ID). User IDs are likely sequential or guessable, making API key enumeration possible.

**Recommendation:** Generate cryptographically random API keys with sufficient entropy (32+ bytes) and store the hash for verification.

---

### H-2: SSRF Risk - Target URL Input Without Validation (SC-TS-021)
**Severity:** High | **CWE:** CWE-918

**Location:** `web/src/pages/Souls.tsx:531-538`

```typescript
<input
  type="text"
  required
  value={formData.target}
  onChange={(e) => setFormData({ ...formData, target: e.target.value })}
  placeholder="https://api.example.com/health"
/>
```

**Issue:** User-provided URLs are accepted without validation. While this is for monitoring targets, it could be exploited to probe internal network resources if the backend doesn't validate target URLs.

**Recommendation:** Implement URL scheme allowlisting (http/https only), hostname validation, and block private IP ranges on the server side.

---

### H-3: No Content Security Policy Configured (SC-TS-264)
**Severity:** High | **CWE:** CWE-1021

**Location:** `web/vite.config.ts`

**Issue:** Vite configuration lacks any Content-Security-Policy headers. This makes XSS attacks more severe as any injected script can load additional resources.

**Recommendation:** Add CSP headers in production deployment:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'
```

---

### H-4: No Security Headers Configured (SC-TS-169, SC-TS-170)
**Severity:** High | **CWE:** CWE-16

**Location:** `web/vite.config.ts`

**Issue:** Missing security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Strict-Transport-Security`
- `Permissions-Policy`

**Recommendation:** Configure security headers in the production server or use a Vite plugin for development.

---

### H-5: WebSocket Reconnection Could Be Abused (SC-TS-187)
**Severity:** High | **CWE:** CWE-400

**Location:** `web/src/hooks/useWebSocket.tsx:91-98`

```typescript
if (reconnectAttemptsRef.current < maxReconnectAttempts) {
  reconnectAttemptsRef.current++
  const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000)
  reconnectTimeoutRef.current = setTimeout(() => {
    connect()
  }, delay)
}
```

**Issue:** WebSocket reconnection logic could be exploited for DoS if an attacker can trigger reconnections repeatedly.

**Recommendation:** Add jitter to reconnection delays and implement connection state tracking for abuse detection.

---

### H-6: Service Worker Without Integrity Check (SC-TS-189)
**Severity:** High | **CWE:** CWE-353

**Location:** `web/src/main.tsx:10`

```typescript
navigator.serviceWorker.register('/sw.js')
```

**Issue:** Service worker is registered without Subresource Integrity (SRI) verification. If `/sw.js` is compromised, malicious code could execute with full page access.

**Recommendation:** Implement SRI for service worker or ensure `/sw.js` is served with strong cache validation.

---

### H-7: Hardcoded Demo Credentials (SC-TS-035)
**Severity:** High | **CWE:** CWE-598

**Location:** `web/src/pages/Login.tsx:196-197`

```typescript
<p className="text-xs text-gray-500 font-cormorant text-center italic">
  Demo offering: <span className="text-[#D4AF37]">admin@anubis.watch</span> / <span className="text-[#D4AF37]">admin</span>
</p>
```

**Issue:** Default credentials displayed in the login page. Security risk if production deployments don't change defaults.

**Recommendation:** Remove hardcoded credentials from production code. Use environment variables for demo mode only.

---

### H-8: Potential XSS via Dynamic Content Rendering (SC-TS-017)
**Severity:** High | **CWE:** CWE-79

**Location:** `web/src/components/widgets/TableWidget.tsx:65`

```typescript
{typeof row[col] === 'boolean' ? (...) : String(row[col] ?? '-')}
```

**Issue:** While React escapes content by default, String() conversion should be verified safe for all API-returned data.

**Recommendation:** Use DOMPurify or similar library to sanitize any data that might contain HTML.

---

## Medium Findings

### M-1: No CSRF Protection on Forms (SC-TS-059)
**Severity:** Medium | **CWE:** CWE-352

**Location:** All form submissions in `web/src/pages/`

**Issue:** Forms performing state-changing operations lack CSRF tokens.

**Recommendation:** Implement CSRF tokens for state-changing operations, especially those using cookie authentication.

---

### M-2: Missing Input Length Validation (SC-TS-002)
**Severity:** Medium | **CWE:** CWE-20

**Location:** Multiple form inputs

**Issue:** Form inputs don't have explicit length limits.

**Recommendation:** Add `maxLength` attributes to text inputs and validate on server side.

---

### M-3: No Rate Limiting Feedback on Login (SC-TS-027)
**Severity:** Medium | **CWE:** CWE-307

**Location:** `web/src/pages/Login.tsx`

**Issue:** Login provides no feedback about remaining attempts or lockout status.

**Recommendation:** Show rate limit feedback and implement exponential backoff server-side.

---

### M-4: WebSocket URL Derived from window.location (SC-TS-016)
**Severity:** Medium | **CWE:** CWE-79

**Location:** `web/src/hooks/useWebSocket.tsx:49-50`

```typescript
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const wsUrl = `${protocol}//${window.location.host}/ws`
```

**Issue:** WebSocket URL derived from user-controlled `window.location` properties.

**Recommendation:** Use a fixed WebSocket endpoint configured at build time.

---

### M-5: Widget Query Accepts Arbitrary Source/Metric (SC-TS-020)
**Severity:** Medium | **CWE:** CWE-20

**Location:** `web/src/pages/DashboardDetail.tsx:218-222`

```typescript
query: { source, metric, time_range: timeRange },
```

**Issue:** Widget configuration accepts arbitrary source/metric values without validation.

**Recommendation:** Validate widget query configuration against an allowlist.

---

### M-6: Missing Error Boundaries (SC-TS-326)
**Severity:** Medium | **CWE:** CWE-209

**Location:** `web/src/App.tsx`

**Issue:** No error boundaries visible. Unhandled errors could expose stack traces.

**Recommendation:** Implement error boundaries that log errors securely and show user-friendly messages.

---

## Low Findings

### L-1: Missing Strict TypeScript Configuration
**Severity:** Low | **CWE:** CWE-704

**Issue:** TypeScript configuration may not have full strict mode enabled.

**Recommendation:** Enable `strict: true` in tsconfig.json.

---

### L-2: No Request Timeout on API Client (SC-TS-185)
**Severity:** Low | **CWE:** CWE-400

**Location:** `web/src/api/client.ts`

**Issue:** fetch() calls don't specify a timeout.

**Recommendation:** Add AbortController with timeout for all API requests.

---

### L-3: Predictable Widget IDs (SC-TS-317)
**Severity:** Low | **CWE:** CWE-20

**Location:** `web/src/pages/DashboardDetail.tsx:64`

```typescript
id: `w_${Date.now()}`,
```

**Issue:** Widget IDs use timestamp pattern which is predictable.

**Recommendation:** Use cryptographically random IDs.

---

### L-4: Vite Source Maps in Production (SC-TS-253)
**Severity:** Low | **CWE:** CWE-540

**Location:** `web/vite.config.ts`

**Issue:** Config doesn't explicitly disable source map generation.

**Recommendation:** Ensure `build.sourcemap` is `false` or `'hidden'` in production.

---

## Positive Findings (Secure Patterns)

1. **React default escaping** - No dangerouslySetInnerHTML usage found
2. **Proper TypeScript interfaces** - Good use of typed interfaces
3. **Zustand store pattern** - Centralized state management with proper typing
4. **Proper error handling** - Most async operations have try/catch
5. **No eval() usage** - No dynamic code execution with user input
6. **Proper cleanup in hooks** - useEffect cleanup functions present
7. **Authentication redirect** - 401 responses properly redirect to login
8. **No SQL injection** - Using typed queries, not string concatenation

---

## Recommendations Summary

| Priority | Action | Impact |
|----------|--------|--------|
| 1 | Move JWT from localStorage to httpOnly cookies | Critical |
| 2 | Implement server-side session validation | Critical |
| 3 | Add CSP headers | High |
| 4 | Implement CSRF protection | Medium |
| 5 | Add rate limiting feedback on login | Medium |
| 6 | Generate secure random API keys | High |
| 7 | Add security headers (X-Frame-Options, etc.) | High |
| 8 | Implement URL validation on server for SSRF | High |

---

## Test Coverage Notes

Security-focused testing should include:
- Authentication bypass attempts
- XSS payload injection in all input fields
- CSRF token validation testing
- WebSocket connection manipulation
- Session timeout and expiry testing

---

*Report generated using sc-lang-typescript security checklist (415+ items)*