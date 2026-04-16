# Rate Limiting & DoS — Security Check Results

**Scanner:** sc-rate-limiting v1.0.0  
**Target:** `internal/api/rest.go`  
**Findings:** 5

---

## Finding: RATE-001

- **Title:** Rate limit detection uses `strings.Contains` which is imprecise and could miss endpoints
- **Severity:** High
- **Confidence:** 80
- **File:** `internal/api/rest.go:1799`
- **Vulnerability Type:** CWE-799 (Improper Control of Interaction Frequency)
- **Description:** `getLimit` at line 1799 uses `strings.Contains(path, "delete")` and `strings.Contains(path, "update")` to detect sensitive endpoints. This approach is fragile because:
  1. A path like `/api/v1/souls/updated` would incorrectly match the "update" detection
  2. DELETE requests to `/api/v1/souls/:id` (the HTTP method IS delete) would NOT match because the string "delete" doesn't appear in the path
  3. Custom endpoint paths containing these substrings would incorrectly get `sensitiveLimit`
  
  The function relies on the path string containing "delete" or "update", but the actual HTTP method is available in `ctx.Request.Method` and should be used instead.

- **Impact:** DELETE requests to soul/rule/channel endpoints would receive `defaultLimit` (100 req/min) instead of `sensitiveLimit` (20 req/min), allowing 5x more requests than intended. This could enable faster brute-force enumeration of resources for deletion.
- **Remediation:** Use `ctx.Request.Method == http.MethodDelete` and `ctx.Request.Method == http.MethodPut` / `ctx.Request.Method == http.MethodPatch` to detect mutating operations, regardless of path content. Apply `sensitiveLimit` for all mutating methods on sensitive resources.
- **References:** https://cwe.mitre.org/data/definitions/799.html

---

## Finding: RATE-002

- **Title:** Rate limit state maps (`ipClients`, `userClients`) grow unbounded during cleanup race
- **Severity:** Medium
- **Confidence:** 70
- **File:** `internal/api/rest.go:1810`
- **Vulnerability Type:** CWE-770 (Allocation Without Limits)
- **Description:** The cleanup goroutine at line 1810 runs every 5 minutes and removes expired entries under `mu.Lock()`. However, if the server is under attack with many unique IPs, there's a window between when entries are created and when cleanup runs where the maps could grow to consume significant memory. Additionally, the cleanup uses `mu.Lock()` which temporarily blocks all rate limit checks during the 5-minute cleanup cycle — with many entries, this could cause request latency spikes.

The maps store `*clientState` structs which are relatively small, but under a sustained high-IP spoofing attack (using random X-Forwarded-For IPs), the maps could accumulate thousands of entries between cleanup cycles.

- **Impact:** Memory exhaustion DoS via rate limit map growth. An attacker with ability to send requests with varying source IPs (or spoofing X-Forwarded-For) could exhaust server memory with unique IP entries. Additionally, the cleanup lock acquisition could cause latency spikes.
- **Remediation:** 
  1. Add a maximum size cap for the rate limit maps with a simple eviction policy (e.g., LRU or random eviction when cap is reached)
  2. Use a sharded map design (multiple sub-maps with independent locks) to reduce lock contention
  3. Move cleanup to use a read-mostly lock (RWLock) or use a sync.Map which handles concurrent access better
- **References:** https://cwe.mitre.org/data/definitions/770.html

---

## Finding: RATE-003

- **Title:** Rate limit uses X-Forwarded-For without validation, susceptible to Spoofing
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/api/rest.go:1837`
- **Vulnerability Type:** CWE-345 (Insufficient Verification of Data Authenticity)
- **Description:** At line 1837, the rate limiter extracts the client IP from `X-Forwarded-For` header: `if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" { ip = strings.Split(forwarded, ",")[0] }`. This is standard, but the code does not verify whether the request actually came through a trusted proxy that set this header. If the AnubisWatch server is directly exposed to the internet (not behind a trusted reverse proxy), an attacker can spoof the `X-Forwarded-For` header to bypass per-IP rate limiting entirely.

Note: This is only exploitable if AnubisWatch is NOT behind a trusted proxy that sanitizes X-Forwarded-For.

- **Impact:** If not behind a trusted proxy, an attacker could bypass IP-based rate limiting by sending requests with crafted `X-Forwarded-For: <random-ip>` headers, allowing unlimited requests to hit the server and bypass brute-force protections.
- **Remediation:** 
  1. Only use `X-Forwarded-For` when the request comes from a known/trusted proxy IP (check `ctx.Request.RemoteAddr` against a list of trusted proxy IPs)
  2. Alternatively, always use `RemoteAddr` as the rate limit key and only trust `X-Forwarded-For` when the connection comes from a trusted proxy
  3. Add a configuration option to disable X-Forwarded-For processing entirely
- **References:** https://cwe.mitre.org/data/definitions/345.html

---

## Finding: RATE-004

- **Title:** Request body injection pattern check uses simple string contains — bypassable
- **Severity:** Low
- **Confidence:** 75
- **File:** `internal/api/rest.go:1701`
- **Vulnerability Type:** CWE-1333 (ReDoS / Pattern Bypass)
- **Description:** `containsInjectionPatterns` at line 1701 uses simple case-insensitive substring matching for SQL injection, XSS, and path traversal patterns. This approach can be bypassed:
  - SQL: `UNion SELect` or `UniOn SeLECT` would bypass the check since only exact lowercase matches are checked
  - Path traversal: URL-encoded `..%2F` or `..%5C` bypasses the plain `../` check
  - XSS: `<scr%69pt>` (URL-encoded `i`) bypasses `<script` check

The function also only checks for literal string patterns, not for obfuscated variants.

- **Impact:** Low — This is a defense-in-depth measure (JSON body validation), and the actual database/storage layer likely has its own parameterization. However, an attacker could potentially inject these patterns through the REST API if the downstream systems don't properly sanitize.
- **Remediation:** 
  1. URL-decode the body before checking for path traversal patterns
  2. Use a proper SQL injection detection library or parameterized queries for any string interpolation
  3. Remove the string-based injection check and rely solely on schema validation (`json.Unmarshal` into a typed struct) which naturally rejects unexpected structures
- **References:** https://cwe.mitre.org/data/definitions/1333.html

---

## Finding: RATE-005

- **Title:** ReDoS vulnerability in `containsInjectionPatterns` with many pattern checks
- **Severity:** Low
- **Confidence:** 60
- **File:** `internal/api/rest.go:1717`
- **Vulnerability Type:** CWE-1333 (ReDoS)
- **Description:** `containsInjectionPatterns` iterates over 18 SQL injection patterns and calls `strings.Contains` for each one on the full input string. While each individual `strings.Contains` call is O(n), the nested loop makes the total complexity O(n * patterns). For a maliciously large request body (close to the 1MB limit), this could cause measurable CPU usage per request. The function converts input to lowercase on each iteration (`lowerInput := strings.ToLower(input)`) and then checks all 18 patterns.

An attacker could send many requests with payloads designed to maximize the time spent in this function, contributing to CPU exhaustion.

- **Impact:** Low — The function runs on every POST/PUT request with a body, but the input is bounded to 1MB and the pattern checks are simple substring operations. The impact would require a very high request rate to cause meaningful CPU exhaustion.
- **Remediation:** Remove the loop entirely. Schema validation (unmarshaling into a typed struct) already prevents injection attacks by rejecting strings that don't match the expected structure. If additional checks are needed, use a single pass with a regex that is known to be safe (no nested quantifiers).
- **References:** https://cwe.mitre.org/data/definitions/1333.html

---

## Positive Findings (Secure Patterns)

- `maxRequestBodySize = 1 << 20` (1MB) at rest.go:21 — Prevents large payload DoS
- `maxJSONDepth = 32` at rest.go:23 — Prevents stack overflow via deeply nested JSON
- `maxDepthReader` struct at rest.go:29 — Custom reader enforcing depth limits
- Rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`) set at lines 1895-1897
- Per-user rate limit (2x IP limit) at line 1871 — Allows legitimate multi-IP users while limiting abuse
- Retry-After header set at line 1862 when rate limit exceeded
- `ReadTimeout: 30 * time.Second` and `ReadHeaderTimeout: 10 * time.Second` at rest.go:462-464