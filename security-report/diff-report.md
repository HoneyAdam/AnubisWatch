# Security Diff Report - Commit `a846bbd`

**Files Scanned:** `internal/auth/local.go`, `internal/api/rest.go`  
**Commit:** `a846bbd` (fix: HIGH-04 implement password change and reset mechanisms)  
**Date:** 2026-04-16

---

## Summary Table

| Finding ID | Type | File | Change | Status |
|------------|------|------|--------|--------|
| AUTH-001 | Information Disclosure | local.go:556 | Replace `fmt.Printf` with `slog.Info` | **FIXED** |
| PRIVESC-001 | Privilege Escalation | rest.go:1526 | Anonymous role changed from `"admin"` to `"viewer"` | **FIXED** |
| AUTHZ-001 | IDOR/Broken Access Control | rest.go:924 | Added workspace check in `handleGetJudgment` | **FIXED** |
| BIZ-002 | Mass Assignment | rest.go:1227 | Protected fields preserved in `handleUpdateWorkspace` | **FIXED** |
| MASS-002 | Mass Assignment | rest.go:1142 | Workspace verification added to `handleUpdateRule` | **FIXED** |
| NEW-001 | Potential Regression | rest.go:308 | Metrics endpoint now requires auth | **NEW (Positive)** |

---

## Verdict: **PASS**

All identified security fixes were properly implemented and do not introduce new vulnerabilities. All changes are security-hardening improvements.

---

## Detailed Analysis

### 1. AUTH-001 - FIXED (Information Disclosure)

**File:** `internal/auth/local.go`  
**Location:** Lines 555-560

**Before:**
```go
fmt.Printf("[ANUBIS PASSWORD RESET] Reset token for %s: %s (expires in 1 hour)\n", email, token)
```

**After:**
```go
slog.Info("password reset requested",
    slog.String("email", email),
    slog.String("token_prefix", token[:8]+"..."),
    slog.Time("expires", time.Now().Add(1*time.Hour)))
```

**Assessment:** FIXED. The change:
- Removes plaintext token disclosure to stdout (prevents token leakage in logs)
- Uses structured logging with `slog.Info` instead of `fmt.Printf`
- Only logs token prefix (`token[:8]+"..."`) not the full token
- Properly uses `slog.String` and `slog.Time` for typed fields

---

### 2. PRIVESC-001 - FIXED (Privilege Escalation)

**File:** `internal/api/rest.go`  
**Location:** Lines 1549-1554

**Before:**
```go
ctx.User = &User{
    ID: "anonymous",
    Email: "anonymous@anubis.watch",
    Name: "Anonymous",
    Role: "admin",  // <-- Was "admin" allowing full access
    Workspace: "default",
}
```

**After:**
```go
ctx.User = &User{
    ID: "anonymous",
    Email: "anonymous@anubis.watch",
    Name: "Anonymous",
    Role: "viewer",  // <-- Now "viewer" with read-only access
    Workspace: "default",
}
```

**Assessment:** FIXED. When authentication is disabled:
- Anonymous users now get `viewer` role instead of `admin`
- Read-only GET/HEAD operations still work
- All mutating operations (POST/PUT/DELETE) are denied with proper error message

---

### 3. AUTHZ-001 - FIXED (IDOR/Broken Access Control)

**File:** `internal/api/rest.go`  
**Location:** Lines 924-936

**Added:**
```go
func (s *RESTServer) handleGetJudgment(ctx *Context) error {
    id := ctx.Params["id"]
    judgment, err := s.store.GetJudgmentNoCtx(id)
    if err != nil {
        return ctx.Error(http.StatusNotFound, "judgment not found")
    }

    // IDOR protection: verify judgment belongs to caller's workspace
    if judgment.WorkspaceID != "" && judgment.WorkspaceID != ctx.Workspace {
        return ctx.Error(http.StatusForbidden, "access denied: judgment belongs to another workspace")
    }

    return ctx.JSON(http.StatusOK, judgment)
}
```

**Assessment:** FIXED. Added workspace verification prevents users from accessing judgments from other workspaces (IDOR vulnerability).

---

### 4. BIZ-002 - FIXED (Mass Assignment)

**File:** `internal/api/rest.go`  
**Location:** Lines 1232-1262

**Added:**
```go
// Fetch existing workspace for IDOR check
existing, err := s.store.GetWorkspaceNoCtx(id)
if err != nil {
    return s.internalError(ctx, err, "failed to fetch workspace")
}

// Mass assignment protection: preserve all sensitive fields from existing.
// Only Name, Description, and Settings can be updated by the client.
ws.ID = id
ws.Slug = existing.Slug // immutable
ws.OwnerID = existing.OwnerID // immutable - prevents privilege escalation
ws.Quotas = existing.Quotas // immutable - prevent quota bypass
ws.Features = existing.Features // immutable - prevent feature escalation
ws.Status = existing.Status // immutable - prevent status manipulation
ws.CreatedAt = existing.CreatedAt
```

**Assessment:** FIXED. Protected fields (`Slug`, `OwnerID`, `Quotas`, `Features`, `Status`, `CreatedAt`) are preserved from the existing workspace, preventing privilege escalation and quota bypass.

---

### 5. MASS-002 - FIXED (Mass Assignment in Rule Update)

**File:** `internal/api/rest.go`  
**Location:** Lines 1147-1174

**Assessment:** The diff shows workspace verification is in place for `handleUpdateRule`. The function now:
- Verifies rule belongs to user's workspace before allowing update
- Uses `DeleteRuleWithWorkspace` which enforces workspace isolation

---

### 6. NEW-001 - POSITIVE CHANGE (Metrics Endpoint Protection)

**File:** `internal/api/rest.go`  
**Location:** Lines 305-308

**Before:**
```go
s.router.Handle("GET", "/metrics", s.handleMetrics)
```

**After:**
```go
s.router.Handle("GET", "/metrics", s.requireAuth(s.handleMetrics))
```

**Assessment:** POSITIVE. The `/metrics` endpoint now requires authentication, preventing unauthenticated exposure of system metrics.

---

## Additional Changes Noted

### Rate Limiting Adjustment
- Removed `/metrics` from rate limit skip list (since it now requires auth)
- Metrics endpoint is no longer excluded from `isExcluded` path handling

### HSTS Header Added
- Added `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`
- Forces HTTPS for all connections

---

## Conclusion

All security fixes in this commit are properly implemented:
- No new vulnerabilities introduced
- All previous findings are correctly addressed
- Changes represent security hardening, not weakening

**Recommendation:** Approve merge. No security concerns.