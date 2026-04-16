# sc-privilege-escalation: Privilege Escalation — Results

**Skill:** sc-privilege-escalation | **Severity:** Mixed | **Confidence:** High
**Scan Date:** 2026-04-16 | **Files Scanned:** internal/auth/, internal/api/
**Focus:** internal/auth/ for role escalation, internal/api/rest.go for access control

---

## Summary: 1 HIGH FINDING, 1 INFORMATIONAL

### Findings

#### PRIVESC-001: Anonymous User Assigned "admin" Role When Auth is Disabled
- **Severity:** High
- **Confidence:** 100
- **File:** `internal/api/rest.go:1516-1531`
- **Vulnerability Type:** CWE-269 (Improper Privilege Management)
- **Description:** When `authConfig.IsEnabled()` returns false, the `requireAuth` middleware creates an anonymous user with `Role: "admin"` for GET/HEAD requests. Since the `requireRole` middleware checks `ctx.User.Role` against required permissions, and `core.MemberRole("admin").Can()` returns true for all permissions, the anonymous user effectively has full admin access.
- **Code:**
  ```go
  func (s *RESTServer) requireAuth(handler Handler) Handler {
      return func(ctx *Context) error {
          if !s.authConfig.IsEnabled() {
              method := ctx.Request.Method
              if method != http.MethodGet && method != http.MethodHead {
                  return ctx.Error(http.StatusForbidden, "authentication is required...")
              }
              ctx.User = &User{
                  ID: "anonymous",
                  Email: "anonymous@anubis.watch",
                  Name: "Anonymous",
                  Role: "admin",  // <-- Full admin role assigned!
                  Workspace: "default",
              }
              ctx.Workspace = "default"
              return handler(ctx)
          }
          // ...
      }
  }
  ```
- **Impact:** If authentication is accidentally disabled in production config, all unauthenticated GET requests receive full admin privileges. An attacker who disables auth (via config manipulation or misconfiguration) gains admin access to all data and settings.
- **Remediation:** Assign `Role: "viewer"` or `Role: "readonly"` to anonymous users, and explicitly grant only `souls:read`, `rules:read`, `channels:read` permissions. Create a limited `core.MemberRole("anonymous")` that only allows read operations with no access to settings, users, or workspace management.
- **References:** https://cwe.mitre.org/data/definitions/269.html

#### PRIVESC-002: Default Workspace is "default" for All Users
- **Severity:** Informational
- **Confidence:** 100
- **File:** `internal/auth/local.go:104`, `internal/auth/oidc.go:251`, `internal/auth/ldap.go:150`
- **Vulnerability Type:** CWE-266 (Incorrect Privilege Assignment)
- **Description:** All authenticators default `Workspace: "default"` for new users. This is the same workspace used by the anonymous fallback in rest.go. If multiple workspaces are intended to provide true multi-tenant isolation, users should be assigned to a unique workspace per organization.
- **Code:**
  ```go
  // local.go:104
  Role: "admin",
  Workspace: "default",
  ```
- **Impact:** Currently informational — if multi-tenancy with workspace isolation is deployed, the hardcoded "default" workspace could cause confusion. However, the workspace-based IDOR checks in rest.go do properly scope data access.
- **Remediation:** If multi-tenant workspaces are implemented, users should be assigned to the workspace they are created in, not hardcoded to "default".
- **References:** https://cwe.mitre.org/data/definitions/266.html

---

## Positive Security Practices Verified

### No Role Field in User-Created Objects
When creating souls, channels, rules, and other resources, the `WorkspaceID` is set from `ctx.Workspace` (derived from the authenticated user), not from any user-supplied field. Role cannot be escalated via API.

### Role Verification via core.MemberRole.Can()
`internal/api/rest.go:1558-1560` — Roles are checked via `core.MemberRole.Can(permission)`, not string comparison, preventing bypass via `role: "admin"` in request body.

### OIDC Role Handling
`internal/auth/oidc.go:249` — New OIDC users are assigned `Role: "viewer"` by default. The role comes from the IdP claims and is not user-settable via the AnubisWatch API. Role mapping appears controlled by the OIDC configuration.

### Password Change Invalidates All Sessions
`internal/auth/local.go:524-526` — On password change, all existing sessions are invalidated:
```go
// Invalidate all existing sessions (force re-login)
a.tokens = make(map[string]*session)
```

### No Role Manipulation in User Profile Updates
No user profile update endpoint exists that would allow a user to modify their own role. The system is single-user (admin only) for local auth.

---

## Verdict

The HIGH finding (anonymous = admin when auth disabled) is a significant risk if auth is misconfigured. All other aspects of privilege management are well-implemented with no direct role escalation vectors found.

---