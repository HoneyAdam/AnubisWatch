# LDAP Injection Scan Results (sc-ldap)

**Scanner:** sc-ldap  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/ |
| Vulnerabilities Found | 1 |
| Medium Severity | 1 |
| Confidence | 75 |
| False Positives | 1 |

---

## Findings

### Finding LDAP-001: Unescaped Email in LDAP BindDN Construction

**Title:** LDAP Injection via Unescaped Email in BindDN Pattern  
**Severity:** Medium  
**Confidence:** 75  
**File:** `internal/auth/ldap.go:166`  

**Vulnerability Type:** CWE-90 (LDAP Injection)  

**Description:**
When the LDAP BindDN configuration contains the pattern `{{mail}}`, the email is substituted directly without proper LDAP DN escaping. The `ldap.EscapeFilter()` function is correctly used for the search filter (line 97), but NOT for the BindDN construction.

**Code:**
```go
// internal/auth/ldap.go:164-167
if strings.Contains(l.cfg.BindDN, "{{mail}}") {
    return strings.ReplaceAll(l.cfg.BindDN, "{{mail}}", email)
}
```

**Attack Vector:**
If an attacker can control their email address and the BindDN uses `{{mail}}`, they could inject LDAP special characters:
- Input: `admin@example.com,CN=Anyone,OU=Admins,DC=company,DC=com`
- This could modify the DN structure if not properly escaped

**Impact:**
- Authentication bypass if LDAP bind succeeds with crafted DN
- Unauthorized access to LDAP directory

**Remediation:**
```go
import "github.com/go-ldap/ldap/v3"

if strings.Contains(l.cfg.BindDN, "{{mail}}") {
    // Escape special DN characters: , + " \ < > ; and NUL
    escaped := ldap.EscapeDN(email)
    return strings.ReplaceAll(l.cfg.BindDN, "{{mail}}", escaped)
}
```

**Note:** The Go ldap/v3 library provides `ldap.EscapeFilter()` but NOT `ldap.EscapeDN()`. You may need to implement DN escaping manually or use a regex-based approach.

---

### Positive Security Finding: Search Filter Properly Escaped

**File:** `internal/auth/ldap.go:97`

```go
filter = strings.ReplaceAll(filter, "{{mail}}", ldap.EscapeFilter(email))
```

The search filter correctly uses `ldap.EscapeFilter()` to escape LDAP special characters: `* ( ) \ NUL`

---

### False Positive: buildUserDN uses safe pattern

**File:** `internal/auth/ldap.go:169`

```go
return fmt.Sprintf("CN=%s,%s", username, l.cfg.BaseDN)
```

This is a **false positive**. The `username` is extracted by splitting the email at `@` (line 163), so it cannot contain LDAP special characters like `,` that would break the DN structure. The `BaseDN` is configuration, not user input.

---

## Verdict

**1 MEDIUM VULNERABILITY FOUND**

The BindDN construction at line 166 does not escape the email before substitution when using the `{{mail}}` pattern. While `ldap.EscapeFilter()` is correctly used for search filters, the BindDN uses raw string replacement without DN escaping.

---

## LDAP Configuration Reference

The LDAP authenticator is configured via `internal/core/feather.go:220-225`:
```go
Type string `json:"type" yaml:"type"` // local, oidc, ldap
LDAP LDAPAuth `json:"ldap" yaml:"ldap"`
```

The BindDN pattern is read from `core.LDAPAuth.BindDN` configuration.

---

## References

- CWE-90: LDAP Injection
- https://cwe.mitre.org/data/definitions/90.html