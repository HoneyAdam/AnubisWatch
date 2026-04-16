# SQL Injection Scan Results (sc-sqli)

**Scanner:** sc-sqli  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/ |
| Vulnerabilities Found | 0 |
| Informational Findings | 0 |
| Confidence | High |

---

## Analysis

### Storage Architecture

The application uses **CobaltDB**, a custom **B+Tree embedded storage engine** (`internal/storage/engine.go`). It is NOT a SQL database (no PostgreSQL, MySQL, SQLite). All data access is through Go method calls on the B+Tree structure, not SQL queries.

Key storage files:
- `internal/storage/engine.go` - CobaltDB B+Tree engine
- `internal/storage/judgments.go` - Judgment time-series storage
- `internal/storage/timeseries.go` - Time-series data support

### Findings

#### No SQL Query Construction Found

No usage of:
- `db.Query()`, `db.Exec()`, `db.QueryRow()` (Go database/sql)
- String concatenation into SQL statements
- Raw SQL strings with user input interpolation
- ORM raw query methods (`$queryRawUnsafe`, `.rawQuery()`, etc.)

**Note:** `internal/api/rest.go:1714` contains `containsInjectionPatterns()` which is a **security validation function** (defense mechanism), not a vulnerability. It checks for SQL injection patterns in user input.

### Test Coverage

The file `internal/api/handlers_extra_test.go:985-989` contains SQL injection test vectors (e.g., `"SELECT * FROM users"`, `"UNION SELECT password FROM admin"`). These are **test data only**, not actual SQL execution.

### Verdict

**NO SQL INJECTION VULNERABILITIES FOUND**

The project uses no SQL database. All storage is through a custom B+Tree engine with method-based access. No code path exists where user-controlled input could be concatenated into a SQL query string.

---

## Common False Positives Eliminated

1. **B+Tree storage (CobaltDB)** - not SQL, no injection possible
2. **Test vectors** - test data, not runtime code
3. **Security validation functions** - `containsInjectionPatterns()` is a defense, not a vulnerability

## References

- CWE-89: SQL Injection
- https://owasp.org/Top10/A03_2021-Injection/