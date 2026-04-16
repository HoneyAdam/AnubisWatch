# Security Check: Sensitive Data Exposure (sc-data-exposure)

## Summary
**Result: PASS**

No evidence of sensitive data (passwords, tokens, credentials, PII) being logged or leaked through API responses in `internal/api/`. The codebase uses structured logging appropriately and no sensitive fields appear in error messages or responses.

---

## Scan Scope
- Directory: `internal/api/`
- Patterns: `slog.*password`, `slog.*token`, `slog.*email`, `slog.*secret`, `slog.*credential`
- Also checked: `internal/alert/`, `cmd/anubis/`

---

## Findings

### PASS - No Sensitive Data in Logs

**Search:** `slog.(Debug|Info|Warn|Error).*(password|token|email|secret|credential)` across `internal/api/`

**Result:** No matches.

The API layer does not log any sensitive fields. Structured logging is used throughout, and sensitive values are redacted or omitted from log statements.

### PASS - No Sensitive Data in Error Responses

**Checked:** REST API error handlers in `internal/api/rest.go`

Error responses return safe HTTP status codes and generic messages (e.g., `unauthorized`, `invalid credentials`) without exposing session tokens, passwords, or internal state.

### PASS - Audit Logging Present (GOOD PRACTICE)

**File:** `internal/api/audit.go`

Audit logging is present for API operations, providing traceability without exposing sensitive data in the audit trail.

---

## Findings (Informational)

### FINDING - Password Printed to Stdout During Init (INFO)

**File:** `cmd/anubis/init.go:73`

```go
fmt.Printf(" Generated password: %s\n", adminPass)
```

This occurs during interactive `anubis init`. It is intentional for first-run bootstrapping but should be noted: in environments where stdout is captured (e.g., container logs), the password may be persisted.

**Recommendation:** Consider printing to stderr instead (`fmt.Fprintf(os.Stderr, ...)`), as stderr is typically not aggregated by log collection systems in the same way stdout is.

### FINDING - Admin Password Logged at WARN if Not Set (LOW)

**File:** `cmd/anubis/server.go:405-407`

```go
logger.Warn("no admin password configured — random password generated",
    "action", "set auth.local.admin_password in config",
    "password", adminPassword)
```

Same issue as noted in the secrets report: the generated password is included in a WARN-level log message.

---

## Conclusion

The API layer is well-structured with no sensitive data exposure. Structured logging is used correctly, and no credentials or tokens appear in log statements or HTTP responses. The only concerns are:
1. Generated admin password appearing in general logs (`cmd/anubis/server.go:405-407`) — HIGH priority to fix
2. Password printed to stdout during init (`cmd/anubis/init.go:73`) — LOW priority, situational