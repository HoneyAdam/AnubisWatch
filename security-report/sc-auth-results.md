# sc-auth: Authentication Flaws — Results

**Skill:** sc-auth | **Severity:** Mixed | **Confidence:** High
**Scan Date:** 2026-04-16 | **Files Scanned:** internal/auth/, internal/api/
**Focus:** internal/auth/local.go, oidc.go, ldap_test.go, internal/api/rest.go

---

## Summary: 1 MEDIUM FINDING, 4 POSITIVE SECURITY PRACTICES

### Findings

#### AUTH-001: Password Reset Token Logged to Server.stdout
- **Severity:** Medium
- **Confidence:** 100
- **File:** `internal/auth/local.go:556`
- **Vulnerability Type:** CWE-532 (Information Exposure Through Log Files)
- **Description:** The `RequestPasswordReset` function generates a secure random token but immediately prints it to stdout via `fmt.Printf`. Since AnubisWatch runs as a server process, stdout is typically captured by the process supervisor/systemd, exposing reset tokens in log files.
- **Impact:** Anyone with access to server logs (system administrators, log aggregation systems) can obtain valid password reset tokens, leading to account takeover.
- **Remediation:** Remove the `fmt.Printf` call entirely. In production, the reset token should be sent via email to the admin's registered email address. The current comment acknowledges this ("no email infrastructure") — at minimum, log tokens at DEBUG level (not Info/Print), or remove the output entirely until email is implemented.
- **References:** https://cwe.mitre.org/data/definitions/532.html

---

## Positive Security Practices Verified

### Strengths

1. **Strong Password Hashing** — `internal/auth/local.go:76`
   - Uses `bcrypt.GenerateFromPassword` with cost factor 12 (`bcryptCost = 12`)
   - Default cost is `bcrypt.DefaultCost` (10) in dummy comparison, but main password uses cost 12

2. **Brute Force Protection** — `internal/auth/local.go:375-382`
   - Max 5 login attempts (`maxLoginAttempts = 5`)
   - 15-minute lockout (`lockoutDuration = 15 * time.Minute`)
   - Attempt counter resets after 30 minutes (`attemptResetWindow`)
   - Implemented via `checkBruteForceProtection`, `recordFailedAttempt`, `clearFailedAttempts`

3. **Timing-Safe Comparison** — `internal/auth/local.go:292`
   - Uses `subtle.ConstantTimeCompare` for email comparison
   - Dummy bcrypt comparison prevents timing-based user enumeration

4. **Password Policy Enforcement** — `internal/auth/local.go:384-419`
   - Minimum 12 characters (`minPasswordLength = 12`)
   - Requires 3 of 4 character classes (upper, lower, digit, special)
   - Enforced before password hashing and during password change

5. **Session Token Generation** — `internal/auth/local.go:355-364`
   - 32 bytes from `crypto/rand` (256-bit entropy)
   - Fail-closed: panics on CSPRNG failure rather than returning predictable token

6. **Generic Error Messages** — `internal/auth/local.go:288,298,304`
   - Returns "invalid credentials" for both wrong email and wrong password
   - Dummy bcrypt comparison run to normalize timing

7. **Token Expiration** — `internal/auth/local.go:261-265`
   - Sessions expire after 24 hours (`time.Now().Add(24 * time.Hour)`)
   - Expired sessions are deleted on access

8. **Rate Limiting on Auth Endpoints** — `internal/api/rest.go:1799-1801`
   - Auth endpoints limited to 10 requests/minute (`authLimit = 10`)
   - Separate per-IP and per-user rate limiting

---

## Verdict

The authentication system is well-designed with strong security practices. The single finding (password reset token logged to stdout) is a medium-severity issue that should be addressed by removing the print statement or routing it to a secure audit log.

---