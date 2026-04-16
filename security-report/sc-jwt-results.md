# JWT Implementation Flaws — Security Check Results

**Scanner:** sc-jwt v1.0.0  
**Target:** `internal/auth/oidc.go`  
**Findings:** 6

---

## Finding: JWT-001

- **Title:** ID token audience validation accepts missing audience claim in some branches
- **Severity:** High
- **Confidence:** 85
- **File:** `internal/auth/oidc.go:636`
- **Vulnerability Type:** CWE-345 (Insufficient Verification of Data Authenticity)
- **Description:** `parseIDToken` at line 636 has audience validation logic with three branches:
  - If `aud` is a string and matches `o.config.ClientID` — accepted
  - If `aud` is a list and contains `o.config.ClientID` — accepted  
  - **Else: `return nil, fmt.Errorf("missing audience claim")`**

  However, there is a subtle issue: at line 636, if `aud, ok := claims["aud"].(string); ok && aud != ""` is false (aud is not a non-empty string), but `audList, ok := claims["aud"].([]interface{}); ok` is true (it IS a list), and the list doesn't contain the client ID, it falls through to the else and returns an error. This is correct.

  But at line 651, the `else` branch (no aud at all) returns `fmt.Errorf("missing audience claim")` — this is correct.

  The actual issue is that if `aud` is a list containing client ID (found=true), but the list also contains OTHER audiences (e.g., an attacker-controlled audience), the token is still accepted. The OIDC spec recommends validating that the audience list contains ONLY the expected audience, not just that it contains it.

- **Impact:** If an OIDC provider issues tokens with multiple audiences (e.g., both the AnubisWatch client ID and an attacker-controlled client ID), the token would be accepted. An attacker who registers their own OAuth client with the same OIDC provider could potentially receive tokens meant for AnubisWatch.
- **Remediation:** After finding a match in the audience list, additionally verify that no unexpected audiences are present, or use an exact match when `aud` is a string (no list allowed). If multiple audiences are expected, document and validate against an explicit allowlist.
- **References:** https://cwe.mitre.org/data/definitions/345.html

---

## Finding: JWT-002

- **Title:** OIDC access token is not cryptographically verified — only decoded
- **Severity:** Medium
- **Confidence:** 90
- **File:** `internal/auth/oidc.go:231`
- **Vulnerability Type:** CWE-347 (Improper Verification of Cryptographic Signature)
- **Description:** At line 231, `o.getUserInfo(tokenResp.AccessToken)` fetches user info from the OIDC provider using the access token. The `getUserInfo` function sends the access token in the `Authorization: Bearer` header and trusts the response without verifying the access token's signature or claims. The access token is a JWT (or opaque token) issued by the OIDC provider, but its signature is never verified by AnubisWatch.

While `parseIDToken` at line 225 verifies the ID token's signature and claims, the access token is used directly to fetch user info. If the OIDC provider uses opaque access tokens, this is fine. But if access tokens are signed JWTs, their signature and claims (exp, aud, iss) should be validated.

- **Impact:** If the OIDC provider issues access tokens as signed JWTs and AnubisWatch blindly trusts them without verification, a tampered or expired access token could be used to fetch user info. However, the actual authentication relies on the ID token verification at line 225, so the impact is limited to potential information disclosure via userinfo endpoint.
- **Remediation:** If access tokens are JWTs, verify their signature using the same JWK set and validate exp/aud/iss claims. Alternatively, use only the ID token for user identification and skip the userinfo fetch entirely (the ID token already contains email, name, and sub).
- **References:** https://cwe.mitre.org/data/definitions/347.html

---

## Finding: JWT-003

- **Title:** JWT `kid` parameter not validated against a trusted key ID allowlist
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/auth/oidc.go:424`
- **Vulnerability Type:** CWE-345 (Insufficient Verification of Data Authenticity)
- **Description:** `findKeyForJWT` at line 424 matches keys from the JWK set by `kid` if present in the JWT header, or returns the first usable key if no `kid` is present. However, there is no validation that the `kid` in the JWT header corresponds to a key that is authorized for this client. An attacker could craft a JWT with a `kid` that matches a different key in the JWKS (e.g., a key used for signing other things), potentially allowing algorithm confusion or key confusion attacks.

Additionally, if `kid != ""` and no matching key is found, the function returns an error at line 447. But if `kid == ""`, it returns the first usable key at line 450-453 — this is a fallback that could use the wrong key if the JWKS contains multiple keys for different purposes.

- **Impact:** If the OIDC provider's JWKS contains keys meant for different audiences or purposes, an attacker could potentially trick AnubisWatch into using the wrong key to verify a token. Combined with algorithm confusion (not directly exploitable given the alg restriction at line 552), this could lead to token forgery.
- **Remediation:** 
  1. Validate that the `kid` corresponds to a key intended for this client (check `kid` against a filtered subset of the JWKS meant for this application's use)
  2. Reject tokens with no `kid` if multiple keys are present in the JWKS (require `kid` for security)
  3. Document which key types (RSA vs EC) and algorithms are expected for this client
- **References:** https://cwe.mitre.org/data/definitions/345.html

---

## Finding: JWT-004

- **Title:** JWT `nbf` (not-before) claim tolerance not configurable — strict rejection
- **Severity:** Low
- **Confidence:** 90
- **File:** `internal/auth/oidc.go:676`
- **Vulnerability Type:** CWE-345 (Insufficient Verification of Data Authenticity)
- **Description:** `parseIDToken` at line 676 checks `if nbf, ok := claims["nbf"].(float64); ok` and rejects the token if `time.Now().Before(time.Unix(int64(nbf), 0))`. There is no clock skew tolerance — a token with `nbf` up to a few seconds in the future (due to clock drift between servers) would be rejected. OIDC providers and clients typically allow a small tolerance (e.g., 30-60 seconds) for `nbf` validation to account for clock synchronization issues.

- **Impact:** In environments with slight clock drift between the AnubisWatch server and the OIDC provider, valid tokens could be incorrectly rejected, causing authentication failures. This could result in service disruption or users being locked out of their accounts.
- **Remediation:** Add a configurable `nbfTolerance` duration (e.g., 30 seconds) to the OIDC config. When checking `nbf`, use `time.Now().Add(-nbfTolerance).Before(time.Unix(...))` to allow a small clock skew window.
- **References:** https://cwe.mitre.org/data/definitions/345.html

---

## Finding: JWT-005

- **Title:** JWK set fetched on every token parse when cache is empty — no persistent caching
- **Severity:** Low
- **Confidence:** 75
- **File:** `internal/auth/oidc.go:372`
- **Vulnerability Type:** CWE-770 (Allocation Without Limits)
- **Description:** `fetchJWKs` at line 372 uses a 24-hour TTL for the JWK cache (`jwksTTL: 24 * time.Hour`). However, if a new JWT arrives and the cache is empty (e.g., after a server restart or cache eviction), every incoming token would trigger a JWK fetch. With many concurrent users, this could result in a thundering herd problem where all concurrent requests during a cache miss simultaneously fetch the JWKS.

The double-checked locking pattern at lines 384-386 is correct, but under high concurrency with cache misses, multiple goroutines could still perform simultaneous JWKS fetches before any of them populates the cache.

- **Impact:** Under high concurrency with cache misses, the JWKS endpoint could be overwhelmed with requests, potentially causing the OIDC provider to rate-limit or block AnubisWatch's JWKS fetching. This could cause authentication failures for all users.
- **Remediation:** Add a retry with jitter and exponential backoff if JWKS fetch fails. Consider using a shared/fetching cache (e.g., `singleflight`) to ensure only one goroutine fetches the JWKS while others wait, even with cache misses.
- **References:** https://cwe.mitre.org/data/definitions/770.html

---

## Finding: JWT-006

- **Title:** State parameter HMAC uses only SHA-256 without algorithm agility
- **Severity:** Low
- **Confidence:** 85
- **File:** `internal/auth/oidc.go:128`
- **Vulnerability Type:** CWE-347 (Improper Verification of Cryptographic Signature)
- **Description:** `signState` at line 128 uses `hmac.New(sha256.New, o.stateHMAC)` with a hardcoded SHA-256. The HMAC key is a 32-byte random value generated at startup (`stateHMAC := make([]byte, 32)` at line 109). While this is cryptographically sound for a symmetric MAC, there is no mechanism to rotate the HMAC key or change the hash function without restarting the server.

If the HMAC key is compromised (e.g., through memory disclosure), all state tokens signed with that key become forgeable. There is no key rotation mechanism or way to invalidate old state tokens.

- **Impact:** If the HMAC key is compromised (e.g., via a memory dump or log exposure), an attacker could forge valid OIDC state parameters, enabling CSRF attacks against the OIDC login flow. The state parameter binds the authorization request to the callback, so forging it could allow an attacker to intercept the ID token.
- **Remediation:** 
  1. Implement key rotation — store the current key ID (kid) in the signed state, and maintain a list of valid keys for verification
  2. Add a `key_version` field to the state token, allowing rotation without invalidating all existing sessions
  3. Store HMAC keys in the config file or environment variable (not just in memory) with support for multiple keys during rotation
- **References:** https://cwe.mitre.org/data/definitions/347.html

---

## Positive Findings (Secure Patterns)

- Algorithm restriction at oidc.go:551 — Rejects `alg: none` and only allows RS256/384/512 and ES256/384/512
- Issuer validation at oidc.go:630 — Strict issuer match required
- Nonce validation at oidc.go:662 — Binds ID token nonce to callback session
- Expiration check at oidc.go:669 — Validates `exp` claim
- Constant-time state signature comparison at oidc.go:146 — `hmac.Equal` used for MAC comparison
- HMAC key is a 32-byte cryptographically random value generated via `io.ReadFull(rand.Reader)` at startup
- CSRF protection via state+nonce dual binding
- Token expiration checked on authentication at oidc.go:704
- Session tokens are 32-byte random hex strings generated via `crypto/rand.Read` — not guessable