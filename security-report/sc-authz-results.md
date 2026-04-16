# sc-authz: Authorization Flaws / IDOR — Results

**Skill:** sc-authz | **Severity:** Mixed | **Confidence:** High
**Scan Date:** 2026-04-16 | **Files Scanned:** internal/api/rest.go, internal/grpcapi/, internal/
**Focus:** internal/api/rest.go for access control checks

---

## Summary: 1 MEDIUM FINDING, 1 LOW FINDING, BROAD IDOR PROTECTION VERIFIED

### Findings

#### AUTHZ-001: Missing Workspace Authorization Check on handleGetJudgment
- **Severity:** Medium
- **Confidence:** 90
- **File:** `internal/api/rest.go:924-932`
- **Vulnerability Type:** CWE-639 (Insecure Direct Object Reference)
- **Description:** The `handleGetJudgment` handler fetches a judgment by ID without verifying that the judgment belongs to a soul within the user's workspace. An authenticated user could potentially access judgments from other workspaces if they know or guess the judgment ID.
- **Impact:** Horizontal privilege escalation — a user in workspace A could access judgment data from workspace B if they guess/enumerate judgment IDs. While judgment data is time-series monitoring data (not directly sensitive), this represents broken access control.
- **Code:**
  ```go
  func (s *RESTServer) handleGetJudgment(ctx *Context) error {
      id := ctx.Params["id"]
      judgment, err := s.store.GetJudgmentNoCtx(id)
      if err != nil {
          return ctx.Error(http.StatusNotFound, "judgment not found")
      }
      // NO workspace check here — judgment could belong to any workspace
      return ctx.JSON(http.StatusOK, judgment)
  }
  ```
- **Remediation:** Add workspace verification by fetching the associated soul and checking `soul.WorkspaceID == ctx.Workspace` before returning the judgment.
- **References:** https://cwe.mitre.org/data/definitions/639.html

#### AUTHZ-002: handleListAllJudgments Returns Empty Array (Stub)
- **Severity:** Low
- **Confidence:** 100
- **File:** `internal/api/rest.go:934-937`
- **Vulnerability Type:** CWE-863 (Incorrect Authorization)
- **Description:** The `handleListAllJudgments` handler is a stub that returns an empty array with no filtering logic. If implemented in the future without proper workspace filtering, it could expose all judgments across all workspaces.
- **Code:**
  ```go
  func (s *RESTServer) handleListAllJudgments(ctx *Context) error {
      // List recent judgments across all souls
      return ctx.JSON(http.StatusOK, []interface{}{})
  }
  ```
- **Impact:** Currently returns empty data so no exposure, but the placeholder comment "List recent judgments across all souls" is a red flag for future implementation that could skip workspace filtering.
- **Remediation:** Implement with explicit workspace-scoped query: `s.store.ListJudgmentsNoCtx(workspace, start, end, limit)` using `ctx.Workspace`.
- **References:** https://cwe.mitre.org/data/definitions/863.html

---

## Positive Security Practices Verified

### Workspace Isolation (HIGH-09 compliance)
The codebase demonstrates comprehensive workspace-based IDOR protection across most handlers:

1. **handleGetSoul** — `internal/api/rest.go:820-833`
   - Checks `soul.WorkspaceID != ctx.Workspace` ✅

2. **handleUpdateSoul** — `internal/api/rest.go:835-861`
   - Checks `existing.WorkspaceID != ctx.Workspace` ✅

3. **handleDeleteSoul** — `internal/api/rest.go:863-881`
   - Checks `soul.WorkspaceID != ctx.Workspace` ✅

4. **handleForceCheck** — `internal/api/rest.go:883-899`
   - Checks `soul.WorkspaceID != ctx.Workspace` ✅

5. **handleListJudgments** — `internal/api/rest.go:901-922`
   - Verifies soul belongs to user's workspace before listing judgments ✅

6. **handleGetChannel** — `internal/api/rest.go:996-1012`
   - Checks `ch.WorkspaceID != ctx.Workspace` ✅

7. **handleGetRule** — `internal/api/rest.go:1124-1140`
   - Checks `rule.WorkspaceID != ctx.Workspace` ✅

8. **handleListChannels** — `internal/api/rest.go:945`
   - Filters by workspace via `ListChannelsByWorkspace(ctx.Workspace)` ✅

9. **handleListRules** — `internal/api/rest.go:1074`
   - Filters by workspace via `ListRulesByWorkspace(ctx.Workspace)` ✅

### Role-Based Access Control (requireRole middleware)
`internal/api/rest.go:1551-1564` — `requireRole` middleware properly checks role permissions using `core.MemberRole.Can(permission)`:
- All mutating endpoints (POST, PUT, DELETE) require specific permissions
- Permissions follow pattern `resource:*` (e.g., `souls:*`, `channels:*`, `rules:*`)
- Admin-only endpoints use `settings:write` or `members:*`

### Auth Disabled Path
`internal/api/rest.go:1516-1531` — When auth is disabled, mutating requests are blocked:
```go
if !s.authConfig.IsEnabled() {
    method := ctx.Request.Method
    if method != http.MethodGet && method != http.MethodHead {
        return ctx.Error(http.StatusForbidden, "authentication is required...")
    }
}
```

### Path Traversal Prevention
`internal/api/rest.go:1754-1777` — `validatePathParams` checks for `..`, `//`, null bytes in path parameters.

---

## Verdict

The authorization system is well-implemented with comprehensive workspace-level isolation on most endpoints. The two findings (handleGetJudgment missing workspace check, handleListAllJudgments stub) should be addressed to achieve complete IDOR protection.

---