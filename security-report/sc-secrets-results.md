# Security Check: Hardcoded Secrets (sc-secrets)

## Summary
**Result: PASS with Notes**

No hardcoded API keys, passwords, or tokens were found in production source code. Test files contain expected test credentials.

---

## Scan Scope
- Go files: `cmd/anubis/`, `internal/`, all `*.go`
- TypeScript files: `web/src/`
- Patterns: API keys, passwords, tokens, Bearer auth, private keys

---

## Findings

### PASS - No Secrets in Production Code

No hardcoded secrets found in production Go or TypeScript source files.

### FINDING - Admin Password Logged on Startup (LOW)

**File:** `cmd/anubis/server.go:405-407`

When no admin password is configured, a random password is generated and **logged at WARN level**:

```go
logger.Warn("no admin password configured — random password generated",
    "action", "set auth.local.admin_password in config",
    "password", adminPassword)
```

**Risk:** The generated admin password appears in server logs. In shared-hosting or multi-tenant log aggregation environments, this could expose the bootstrap credentials.

**Recommendation:** Log the password only once at initialization, or emit it to a separate, more restricted log channel (e.g., `audit` or `secure` level). Alternatively, write it exclusively to a file with restricted permissions, not to a general logger.

### FINDING - Generated Password Printed During Setup (INFO)

**File:** `cmd/anubis/init.go:73`

```go
fmt.Printf(" Generated password: %s\n", adminPass)
```

The interactive setup prints the generated admin password to stdout. This is intentional for first-run setup but is noted for review.

### FINDING - Test Files Contain Expected Test Credentials (OK)

**Files:**
- `internal/alert/dispatchers_test.go:1180,1200` — `password: "testpass"` (test SMTP dispatcher)
- `internal/auth/ldap_test.go:39,77,95,129,161,229,245,264,290` — `TestPass1234!` (test LDAP fixtures)
- `web/src/pages/Login.test.tsx:81` — `password: 'password'` (React UI test stub)

These are test-only fixtures and are not a security concern.

---

## Notable: No API Keys or Token Patterns Found

The following patterns were searched and **none** were found in production code:
- `api_key`, `api-key`, `client_secret` assignments
- `password: "..."` string literals
- `bearer` token patterns
- `private_key` string values

---

## Conclusion

No hardcoded secrets in production code. The primary concern is the admin password being written to general logs at `cmd/anubis/server.go:405-407`, which should be addressed to prevent credential exposure in production log aggregation systems.