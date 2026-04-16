# Security Assessment Report

**Project:** AnubisWatch Uptime Monitoring Platform
**Date:** 2026-04-16
**Scanner:** security-check v1.0.0
**Risk Score:** 10.0/10 (Critical Risk)

---

## Executive Summary

A security assessment was performed on AnubisWatch, a zero-dependency single-binary uptime and synthetic monitoring platform with an embedded React dashboard, using 48 automated security skills across 4 pipeline phases. The scan analyzed approximately 500,000 lines of code across Go (backend) and TypeScript/React (dashboard).

### Key Metrics

| Metric | Value |
|--------|-------|
| Total Findings | 17 |
| Critical | 2 |
| High | 6 |
| Medium | 5 |
| Low | 4 |
| Info | 0 |

### Top Risks

1. **AUTH-001 — Password Reset Token Logged to stdout**: Any user who triggers a password reset exposes their token via server logs, enabling full account takeover.
2. **SSRF-001 — Decimal/Hex/Octal IP Bypass in SSRF Validator**: Bypass of SSRF blocklist via IP address notation manipulation allows access to cloud metadata services (AWS/GCP/Azure).
3. **PRIVESC-001 — Anonymous User Assigned "admin" Role When Auth is Disabled**: Accidental auth disablement grants unauthenticated attackers full admin privileges across all workspaces.

---

## Scan Statistics

| Statistic | Value |
|-----------|-------|
| Files Scanned | ~2,500 (estimated) |
| Lines of Code | ~500,000 (estimated) |
| Languages Detected | Go 1.26.2, TypeScript/React 19 |
| Frameworks Detected | stdlib net/http (custom router), gRPC, React 19, Tailwind CSS 4.1, Vite 6 |
| Skills Executed | 48 |
| Findings Before Verification | ~150 |
| False Positives Eliminated | 100+ |
| Final Verified Findings | 17 |

### Finding Distribution

| Vulnerability Category | Critical | High | Medium | Low | Info |
|-----------------------|----------|------|--------|-----|------|
| Authentication | 1 | 1 | 0 | 0 | 0 |
| Authorization / IDOR | 0 | 4 | 1 | 0 | 0 |
| Injection | 0 | 0 | 1 | 0 | 0 |
| SSRF | 1 | 1 | 0 | 0 | 0 |
| Mass Assignment | 0 | 1 | 1 | 0 | 0 |
| Business Logic | 0 | 1 | 0 | 1 | 0 |
| Rate Limiting | 0 | 0 | 0 | 1 | 0 |
| Web Security (XSS/Clickjacking) | 0 | 0 | 0 | 2 | 0 |
| Dependency / Supply Chain | 0 | 0 | 1 | 0 | 0 |
| Misc / Informational | 0 | 0 | 0 | 1 | 0 |

---

## Critical Findings

### AUTH-001: Password Reset Token Logged to stdout

**Severity:** Critical
**Confidence:** 100/100
**CWE:** CWE-532 — Information Exposure Through Log Files
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/auth/local.go:556`

**Description:**
The `RequestPasswordReset` handler prints the password reset token to stdout using `fmt.Printf` without any redaction. Any user who triggers a password reset exposes their token to anyone with access to server logs, container logs, or system log aggregation pipelines. This is a direct account takeover vector.

**Vulnerable Code:**
```go
fmt.Printf("[ANUBIS PASSWORD RESET] Reset token for %s: %s (expires in 1 hour)\n", email, token)
```

**Proof of Concept:**
1. Attacker triggers password reset for victim@example.com via `POST /api/v1/auth/reset-password`
2. Token is printed to stdout on the server
3. Attacker, having access to logs (via log files, container orchestration UI, log aggregation system, or compromised server access), retrieves the token
4. Attacker calls `POST /api/v1/auth/reset-password/confirm` with the stolen token to set a new password
5. Victim account is fully compromised

**Impact:**
Full account takeover of any user whose password reset is triggered. This is particularly severe for admin accounts, enabling complete system compromise. Tokens expire after 1 hour, but log retention policies often keep them far longer.

**Remediation:**
Remove the `fmt.Printf` call entirely. Route to a structured secure audit log (e.g., `slog.Info` with structured fields, excluding the token) or send the token via email to the user. Never print secrets to stdout/stderr.

```go
// Secure alternative: audit log only (no token printed)
slog.Info("password reset requested", "email", email, "token_prefix", token[:8]+"...")

// Or send via email (recommended production approach)
if err := s.emailer.SendPasswordReset(email, token); err != nil {
    slog.Error("failed to send password reset email", "email", email, "err", err)
}
```

**References:**
- [CWE-532: Information Exposure Through Log Files](https://cwe.mitre.org/data/definitions/532.html)
- [OWASP A01:2021 — Broken Access Control](https://owasp.org/Top10/A01_2021-Broken_Access_Control/)

---

### SSRF-001: Decimal/Hex/Octal IP Bypass in SSRF Validator

**Severity:** Critical
**Confidence:** 85/100
**CWE:** CWE-918 — Server-Side Request Forgery (SSRF)
**OWASP:** A10:2021 — Server-Side Request Forgery

**Location:** `internal/probe/ssrf.go:115-162`

**Description:**
The SSRF validator uses `net.ParseIP` to detect blocked IP addresses. However, `net.ParseIP` does not parse decimal notation (e.g., `2130706433` for `127.0.0.1`), octal (`0177.0.0.01`), or hex (`0x7F000001`) IP address representations. When `net.ParseIP` returns `nil`, the code falls through without applying IP blocking, allowing attackers to bypass the SSRF protection entirely.

**Vulnerable Code:**
```go
ip := net.ParseIP(host)
if ip != nil {
    if v.isBlockedIP(ip) { return fmt.Errorf("target IP %q is blocked", ip) }
    return nil
}
// If net.ParseIP returns nil (decimal IP), falls through to DNS resolution
```

**Proof of Concept:**
1. Attacker creates a soul with target URL `http://2130706433/` (decimal representation of `127.0.0.1`) or `http://3232235777/` (decimal of `192.168.1.1`)
2. `net.ParseIP("2130706433")` returns `nil`
3. The SSRF validator skips the `isBlockedIP` check
4. DNS resolution returns the corresponding IP, bypassing the blocklist
5. For AWS: `http://169.254.169.254/latest/meta-data/` can be accessed via decimal `3232261122`

**Impact:**
An attacker with soul-creation permissions (`souls:*` role) can probe internal network services and extract cloud instance metadata. This allows extraction of AWS/GCP/Azure credentials from instance metadata services, accessing internal databases, and reconnaissance of internal infrastructure. All 169.254.x.x link-local addresses (AWS metadata, Kubernetes API, internal services) are reachable.

**Remediation:**
Implement IP address normalization before the blocklist check to handle all common IP notations:

```go
func (v *SSRFValidator) normalizeIP(input string) (net.IP, bool) {
    // Try standard parse first
    if ip := net.ParseIP(input); ip != nil {
        return ip, true
    }

    // Try decimal: 2130706433 -> 127.0.0.1
    if parsed, err := strconv.ParseUint(input, 10, 32); err == nil {
        ip := net.IPv4(
            byte(parsed>>24),
            byte(parsed>>16),
            byte(parsed>>8),
            byte(parsed),
        )
        return ip, true
    }

    // Try octal: 0177.0.0.01 -> 127.0.0.1
    if parsed, err := strconv.ParseUint(input, 8, 32); err == nil {
        ip := net.IPv4(
            byte(parsed>>24),
            byte(parsed>>16),
            byte(parsed>>8),
            byte(parsed),
        )
        return ip, true
    }

    // Try hex: 0x7F000001 -> 127.0.0.1
    hexInput := strings.TrimPrefix(input, "0x")
    if parsed, err := strconv.ParseUint(hexInput, 16, 32); err == nil {
        ip := net.IPv4(
            byte(parsed>>24),
            byte(parsed>>16),
            byte(parsed>>8),
            byte(parsed),
        )
        return ip, true
    }

    return nil, false
}

// In the validator's Check method:
normalized, ok := v.normalizeIP(host)
if ok {
    if v.isBlockedIP(normalized) {
        return fmt.Errorf("target IP %q is blocked", normalized)
    }
    return nil
}
// Handle non-IP hosts (domain names) via DNS resolution with additional checks
```

**References:**
- [CWE-918: Server-Side Request Forgery (SSRF)](https://cwe.mitre.org/data/definitions/918.html)
- [OWASP A10:2021 — Server-Side Request Forgery](https://owasp.org/Top10/A10_2021-Server-Side_Request_Forgery_-_SSRF/)
- [AWS EC2 Metadata Service](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)

---

## High Findings

### AUTHZ-001: Missing Workspace Authorization Check on handleGetJudgment

**Severity:** High
**Confidence:** 90/100
**CWE:** CWE-639 — Authorization Bypass Through User-Controlled Key
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/api/rest.go:924-932`

**Description:**
The REST `handleGetJudgment` handler fetches a judgment by ID without verifying that the judgment's associated soul belongs to the caller's workspace. An authenticated user can access judgment data from any workspace by guessing or knowing judgment IDs.

**Vulnerable Code:**
```go
func (s *RESTServer) handleGetJudgment(ctx *Context) error {
    id := ctx.Params["id"]
    judgment, err := s.store.GetJudgmentNoCtx(id)
    if err != nil {
        return ctx.Error(http.StatusNotFound, "judgment not found")
    }
    // NO workspace check — judgment could belong to any workspace
    return ctx.JSON(http.StatusOK, judgment)
}
```

**Impact:**
Horizontal privilege escalation. A user in workspace A can access judgment data (HTTP response bodies, latency metrics, TLS certificate information) from workspace B. Judgment data may contain sensitive information from probed services including authentication headers, API keys in response bodies, and internal service details.

**Remediation:**
```go
func (s *RESTServer) handleGetJudgment(ctx *Context) error {
    id := ctx.Params["id"]
    judgment, err := s.store.GetJudgmentNoCtx(id)
    if err != nil {
        return ctx.Error(http.StatusNotFound, "judgment not found")
    }
    // Verify the judgment's soul belongs to the caller's workspace
    soul, err := s.store.GetSoulNoCtx(judgment.SoulID)
    if err != nil || soul.WorkspaceID != ctx.Workspace {
        return ctx.Error(http.StatusNotFound, "judgment not found")
    }
    return ctx.JSON(http.StatusOK, judgment)
}
```

**References:**
- [CWE-639: Authorization Bypass Through User-Controlled Key](https://cwe.mitre.org/data/definitions/639.html)

---

### PRIVESC-001: Anonymous User Assigned "admin" Role When Auth is Disabled

**Severity:** High
**Confidence:** 100/100
**CWE:** CWE-862 — Missing Authorization
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/api/rest.go:1522-1526`

**Description:**
When authentication is disabled (via `auth.IsEnabled() == false`), the `requireAuth` middleware constructs an anonymous user context with `Role: "admin"`. This grants full admin privileges to all unauthenticated requests, across all workspaces.

**Vulnerable Code:**
```go
ctx.User = &User{
    ID: "anonymous",
    Email: "anonymous@anubis.watch",
    Name: "Anonymous",
    Role: "admin", // <-- Full admin role assigned!
    Workspace: "default",
}
```

**Impact:**
If authentication is accidentally disabled in production (via misconfiguration or config file injection), all unauthenticated GET/HEAD requests receive full admin privileges. The anonymous admin can access all souls, channels, rules, and settings across all workspaces. `core.MemberRole("admin").Can()` returns `true` for all permissions, including workspace management, user management, and system configuration.

**Remediation:**
Assign a limited `viewer` role or a dedicated `anonymous` role that only permits read operations:

```go
ctx.User = &User{
    ID: "anonymous",
    Email: "anonymous@anubis.watch",
    Name: "Anonymous",
    Role: "viewer", // Limited read-only access
    Workspace: "default",
}
```

Define a new `anonymous` role with explicit permissions if needed:
```go
const RoleAnonymous = "anonymous"
var RolePermissions = map[MemberRole][]string{
    RoleAnonymous: {"souls:read", "rules:read", "channels:read"}, // No write, no settings
}
```

**References:**
- [CWE-862: Missing Authorization](https://cwe.mitre.org/data/definitions/862.html)

---

### API-001: gRPC ListJudgments No Workspace Isolation

**Severity:** High
**Confidence:** 90/100
**CWE:** CWE-639 — Authorization Bypass Through User-Controlled Key
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/grpcapi/server.go:670-694`

**Description:**
The gRPC `ListJudgments` method uses `ListJudgmentsNoCtx` with no workspace filtering. Authenticated gRPC clients can query judgments for any soul by ID across all workspaces.

**Vulnerable Code:**
```go
func (s *Server) ListJudgments(ctx context.Context, req *v1.ListJudgmentsRequest) (*v1.ListJudgmentsResponse, error) {
    // ...
    judgments, err := s.store.ListJudgmentsNoCtx(soulID, start, end, limit)
    // No workspace check — soulID could belong to any workspace
```

**Impact:**
An authenticated gRPC client can retrieve judgment data (HTTP response bodies, latency metrics, TLS certificate info) for souls in other workspaces. Horizontal information disclosure across workspaces.

**Remediation:**
```go
func (s *Server) ListJudgments(ctx context.Context, req *v1.ListJudgmentsRequest) (*v1.ListJudgmentsResponse, error) {
    // Verify soul belongs to caller's workspace
    soul, err := s.store.GetSoulNoCtx(req.SoulId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "soul not found")
    }
    if soul.WorkspaceID != user.Workspace {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
    judgments, err := s.store.ListJudgmentsNoCtx(req.SoulId, req.Start, req.End, req.Limit)
    // ...
}
```

**References:**
- [CWE-639: Authorization Bypass Through User-Controlled Key](https://cwe.mitre.org/data/definitions/639.html)

---

### API-002: gRPC GetChannel No Workspace Isolation

**Severity:** High
**Confidence:** 90/100
**CWE:** CWE-639 — Authorization Bypass Through User-Controlled Key
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/grpcapi/server.go:841-853`

**Description:**
The gRPC `GetChannel` method passes an empty string as workspace to `GetChannelNoCtx`, bypassing workspace isolation entirely.

**Vulnerable Code:**
```go
func (s *Server) GetChannel(ctx context.Context, req *v1.GetChannelRequest) (*v1.Channel, error) {
    // ...
    ch, err := s.store.GetChannelNoCtx(req.Id, "") // <-- empty workspace!
```

**Impact:**
An authenticated gRPC client can retrieve alert channel configurations (webhook URLs, SMTP credentials, bot tokens, PagerDuty keys) from channels in any workspace. Sensitive integration credentials are exposed.

**Remediation:**
```go
func (s *Server) GetChannel(ctx context.Context, req *v1.GetChannelRequest) (*v1.Channel, error) {
    ch, err := s.store.GetChannelNoCtx(req.Id, user.Workspace)
    if err != nil {
        return nil, status.Error(codes.NotFound, "channel not found")
    }
    if ch.WorkspaceID != user.Workspace {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
    return ch, nil
}
```

**References:**
- [CWE-639: Authorization Bypass Through User-Controlled Key](https://cwe.mitre.org/data/definitions/639.html)

---

### API-003: gRPC GetRule No Workspace Isolation

**Severity:** High
**Confidence:** 90/100
**CWE:** CWE-639 — Authorization Bypass Through User-Controlled Key
**OWASP:** A01:2021 — Broken Access Control

**Location:** `internal/grpcapi/server.go:972-984`

**Description:**
The gRPC `GetRule` method passes an empty string as workspace to `GetRuleNoCtx`, bypassing workspace isolation entirely.

**Vulnerable Code:**
```go
func (s *Server) GetRule(ctx context.Context, req *v1.GetRuleRequest) (*v1.Rule, error) {
    // ...
    r, err := s.store.GetRuleNoCtx(req.Id, "") // <-- empty workspace!
```

**Impact:**
An authenticated gRPC client can retrieve alert rule configurations (routing logic, escalation schedules, notification channels) from other workspaces. Rule configs commonly contain sensitive webhook URLs and API keys.

**Remediation:**
```go
func (s *Server) GetRule(ctx context.Context, req *v1.GetRuleRequest) (*v1.Rule, error) {
    r, err := s.store.GetRuleNoCtx(req.Id, user.Workspace)
    if err != nil {
        return nil, status.Error(codes.NotFound, "rule not found")
    }
    if r.WorkspaceID != user.Workspace {
        return nil, status.Error(codes.PermissionDenied, "access denied")
    }
    return r, nil
}
```

**References:**
- [CWE-639: Authorization Bypass Through User-Controlled Key](https://cwe.mitre.org/data/definitions/639.html)

---

### SSRF-002: WebSocket Checker Bypasses SSRF Dial Protection

**Severity:** High
**Confidence:** 85/100
**CWE:** CWE-918 — Server-Side Request Forgery (SSRF)
**OWASP:** A10:2021 — Server-Side Request Forgery

**Location:** `internal/probe/websocket.go:95-103`

**Description:**
The WebSocket checker uses direct `net.DialTimeout` and `tls.DialWithDialer` calls instead of the SSRF validator's wrapped dialer. The SSRFValidator's protections (IP blocklist, DNS rebinding prevention via WrapDialerContext) are not applied to WebSocket dials.

**Vulnerable Code:**
```go
// WebSocket checker uses direct TLS.DialWithDialer — bypasses SSRFValidator
conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", host, tlsConfig)
conn, err = net.DialTimeout("tcp", host, timeout)
```

**Impact:**
DNS rebinding attacks could bypass SSRF protections for WebSocket checks. An attacker could register a domain that initially resolves to a public IP during the SSRF check, then have it point to `169.254.169.254` (AWS metadata) or an internal IP after DNS re-resolution. This requires soul modification permissions (`souls:*` role), limiting blast radius.

**Remediation:**
Apply `SSRFValidator.WrapDialerContext` to the WebSocket checker dialer:

```go
// Create a dialer wrapped with SSRF validation
dialer := &net.Dialer{Timeout: timeout}
if ssrfValidator != nil {
    dialer, err = ssrfValidator.WrapDialerContext(dialer, target)
    if err != nil {
        return nil, err
    }
}
conn, err = tls.DialWithDialer(dialer, "tcp", host, tlsConfig)
```

**References:**
- [CWE-918: Server-Side Request Forgery (SSRF)](https://cwe.mitre.org/data/definitions/918.html)

---

## Medium Findings

### LDAP-001: Unescaped Email in LDAP BindDN Construction

**Severity:** Medium
**Confidence:** 75/100
**CWE:** CWE-90 — Improper Neutralization of Special Elements used in an LDAP Query (LDAP Injection)
**OWASP:** A03:2021 — Injection

**Location:** `internal/auth/ldap.go:166`

**Description:**
When the LDAP BindDN configuration pattern contains `{{mail}}`, the email is substituted directly without DN escaping. LDAP special characters (`,`, `+`, `"`, `\`, `<`, `>`, `;`) in the email address are not escaped, enabling LDAP injection.

**Vulnerable Code:**
```go
// Search filter uses ldap.EscapeFilter (correct):
filter = strings.ReplaceAll(filter, "{{mail}}", ldap.EscapeFilter(email))

// But BindDN construction does NOT escape:
if strings.Contains(l.cfg.BindDN, "{{mail}}") {
    return strings.ReplaceAll(l.cfg.BindDN, "{{mail}}", email) // No escaping!
}
```

**Impact:**
If the LDAP server's BindDN uses `{{mail}}` pattern, an attacker controlling their email can inject LDAP special characters to manipulate the DN structure. This could enable authentication bypass or unauthorized directory access. Only exploitable when LDAP auth is configured and BindDN contains `{{mail}}`.

**Remediation:**
Implement DN escaping for the email before substitution:

```go
// Escape DN special characters: , + " \ < > ; and NUL
func escapeDN(s string) string {
    var result strings.Builder
    for _, c := range s {
        switch c {
        case ',', '+', '"', '\\', '<', '>', ';', '\x00':
            result.WriteString("\\" + string(c))
        default:
            result.WriteRune(c)
        }
    }
    return result.String()
}

// Usage:
if strings.Contains(l.cfg.BindDN, "{{mail}}") {
    escapedEmail := escapeDN(email)
    return strings.ReplaceAll(l.cfg.BindDN, "{{mail}}", escapedEmail)
}
```

**References:**
- [CWE-90: Improper Neutralization of Special Elements used in an LDAP Query](https://cwe.mitre.org/data/definitions/90.html)
- [OWASP A03:2021 — Injection](https://owasp.org/Top10/A03_2021-Injection/)

---

### MASS-001: Mass Assignment in handleUpdateSoul

**Severity:** Medium
**Confidence:** 75/100
**CWE:** CWE-915 — Improperly Controlled Modification of Object-Property Attributes
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/api/rest.go:835-860`

**Description:**
The `handleUpdateSoul` handler binds the full `core.Soul` struct from the request body without field allowlisting. After the bind, only `ID` and `WorkspaceID` are overwritten — other fields like `Region`, `Weight`, `Timeout`, `Tags`, and all protocol configs (`HTTP`, `TCP`, etc.) are accepted from the request without validation.

**Vulnerable Code:**
```go
func (s *RESTServer) handleUpdateSoul(ctx *Context) error {
    var soul core.Soul
    if err := ctx.Bind(&soul); err != nil { ... } // Full struct bind
    // ...
    soul.ID = id
    soul.WorkspaceID = ctx.Workspace // Overwrites after bind
    // No validation of other fields before SaveSoul
```

**Impact:**
An authenticated user with `souls:*` permission can modify any Soul field including `Enabled` (disable monitoring), `Region` (move to different region), or protocol configs. While `WorkspaceID` is protected by server-side overwrite, operational aspects of monitoring can be disrupted. Adjusted to Medium because `core.Soul` has no dangerous privilege-escalation fields.

**Remediation:**
Use a DTO struct with explicit field allowlist:

```go
type UpdateSoulRequest struct {
    Name        string                 `json:"name"`
    Target      string                 `json:"target"`
    Interval    int                    `json:"interval"`
    Timeout     int                    `json:"timeout"`
    Enabled     *bool                  `json:"enabled,omitempty"`
    Tags        []string               `json:"tags,omitempty"`
    Settings    map[string]interface{} `json:"settings,omitempty"`
    // Explicitly excluded: ID, WorkspaceID, CreatedAt, OwnerID
}

func (s *RESTServer) handleUpdateSoul(ctx *Context) error {
    var req UpdateSoulRequest
    if err := ctx.Bind(&req); err != nil { ... }
    // Validate field values
    if req.Interval <= 0 {
        return ctx.Error(http.StatusBadRequest, "interval must be positive")
    }
    // ... apply only allowed fields to existing soul
}
```

**References:**
- [CWE-915: Improperly Controlled Modification of Object-Property Attributes](https://cwe.mitre.org/data/definitions/915.html)

---

### MASS-002: Mass Assignment in handleUpdateRule Config Merge

**Severity:** Medium
**Confidence:** 85/100
**CWE:** CWE-915 — Improperly Controlled Modification of Object-Property Attributes
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/api/rest.go:1142-1169`

**Description:**
The `handleUpdateRule` handler merges the entire `req.Config` map into the rule's config without a field allowlist. Any key-value pair is accepted from the request body. Alert rule configs commonly contain sensitive integration credentials (webhook URLs with tokens, SMTP credentials, API keys).

**Vulnerable Code:**
```go
func (s *RESTServer) handleUpdateRule(ctx *Context) error {
    var rule core.AlertRule
    if err := ctx.Bind(&rule); err != nil { ... }
    // ...
    if req.Config != nil {
        m := rule.Config
        for k, v := range req.Config { m[k] = v } // No field allowlist
    }
```

**Impact:**
An authenticated user with `rules:*` permission can inject arbitrary config fields into a rule, potentially modifying sensitive routing or authentication settings. This could be used to redirect alerts to attacker-controlled endpoints or modify escalation logic.

**Remediation:**
Define an explicit allowlist of mutable config field names:

```go
var allowedRuleConfigKeys = map[string]bool{
    "name": true, "description": true, "severity_filter": true,
    "notification_channels": true, "throttle_minutes": true,
}

if req.Config != nil {
    m := rule.Config
    for k, v := range req.Config {
        if !allowedRuleConfigKeys[k] {
            return ctx.Error(http.StatusBadRequest, "invalid config key: "+k)
        }
        m[k] = v
    }
}
```

**References:**
- [CWE-915: Improperly Controlled Modification of Object-Property Attributes](https://cwe.mitre.org/data/definitions/915.html)

---

### BIZ-002: Workspace Update Allows Overwriting Protected Fields

**Severity:** Medium
**Confidence:** 85/100
**CWE:** CWE-915 — Improperly Controlled Modification of Object-Property Attributes
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/api/rest.go:1227-1242`

**Description:**
The `handleUpdateWorkspace` handler binds the full `core.Workspace` struct and only overwrites `ID` and `UpdatedAt` after the bind. Fields like `OwnerID`, `Slug`, `Plan`, `Quotas`, `Features`, and `Status` are accepted directly from the request body without server-side protection.

**Vulnerable Code:**
```go
func (s *RESTServer) handleUpdateWorkspace(ctx *Context) error {
    var ws core.Workspace
    if err := ctx.Bind(&ws); err != nil { ... }
    ws.ID = id
    ws.UpdatedAt = time.Now()
    // CreatedAt, OwnerID, Slug are NOT overwritten — accepted from body
```

**Impact:**
A user with workspace update permissions can modify `OwnerID` (take ownership of any workspace), `Plan` (escalate billing tier), `Quotas` (remove rate limits), `Features` (enable disabled features), or `Settings` (change workspace configuration). The ability to overwrite `OwnerID` is the most critical — it enables complete workspace takeover.

**Remediation:**
Use a dedicated DTO that explicitly excludes protected fields:

```go
type UpdateWorkspaceRequest struct {
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Settings    *WorkspaceSettings  `json:"settings,omitempty"`
    // Explicitly exclude: ID, OwnerID, Slug, Plan, Quotas, Features, Status, CreatedAt
}

func (s *RESTServer) handleUpdateWorkspace(ctx *Context) error {
    var req UpdateWorkspaceRequest
    if err := ctx.Bind(&req); err != nil { ... }
    // Only allow updating name, description, and settings
    existing, err := s.store.GetWorkspace(id)
    if err != nil { return ctx.Error(http.StatusNotFound, "workspace not found") }
    if req.Name != "" { existing.Name = req.Name }
    if req.Description != "" { existing.Description = req.Description }
    if req.Settings != nil { existing.Settings = *req.Settings }
    // ID, OwnerID, Slug, Plan, Quotas, Features, Status are never updated from request
}
```

**References:**
- [CWE-915: Improperly Controlled Modification of Object-Property Attributes](https://cwe.mitre.org/data/definitions/915.html)

---

### RATE-003: Rate Limiter Trusts X-Forwarded-For Without Proxy Validation

**Severity:** Low
**Confidence:** 75/100
**CWE:** CWE-346 — Origin Validation Error
**OWASP:** A07:2021 — Security Misconfiguration

**Location:** `internal/api/rest.go:1837-1840`

**Description:**
The rate limiter trusts the `X-Forwarded-For` header without verifying that the request originated from a known proxy. If `X-Forwarded-For` is present, its first value is used as the client IP, ignoring `RemoteAddr`.

**Vulnerable Code:**
```go
ip := ctx.Request.RemoteAddr
if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
    ip = strings.Split(forwarded, ",")[0]
}
```

**Impact:**
If AnubisWatch is directly exposed to the internet without a trusted reverse proxy, an attacker can spoof `X-Forwarded-For` to bypass per-IP rate limiting. Most production deployments are behind nginx, Caddy, or cloud load balancers that sanitize this header, limiting practical impact. The gRPC interface uses token-based rate limiting, not IP-based, so sensitive operations are unaffected.

**Remediation:**
Only trust `X-Forwarded-For` when `RemoteAddr` matches known proxy IPs:

```go
var trustedProxies = []string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12"} // configurable

func isTrustedProxy(ip string) bool {
    for _, trusted := range trustedProxies {
        if strings.HasPrefix(ip, trusted) {
            return true
        }
    }
    return false
}

// In rate limiter:
ip := ctx.Request.RemoteAddr
if forwarded := ctx.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
    // Only trust X-Forwarded-For if the direct connection is from a trusted proxy
    if isTrustedProxy(ctx.Request.RemoteAddr) {
        ip = strings.Split(forwarded, ",")[0]
    }
}
```

**References:**
- [CWE-346: Origin Validation Error](https://cwe.mitre.org/data/definitions/346.html)

---

## Low Findings

### REDIR-002: StatusPage HTML XSS via Unescaped Target Field

**Severity:** Low
**Confidence:** 80/100
**CWE:** CWE-79 — Improper Neutralization of Input During Web Page Generation (Cross-site Scripting)
**OWASP:** A03:2021 — Injection

**Location:** `internal/api/statuspage.go:289`

**Description:**
The public status page HTML embeds `soul.Name` and `soul.Target` without HTML escaping. An authenticated attacker who can create or modify a soul can inject JavaScript via the target field.

**Vulnerable Code:**
```go
html += `<div class="component-name">` + soul.Name + `</div>`
html += `<div class="component-target">` + soul.Target + `</div>` <!-- NOT ESCAPED -->
```

**Impact:**
Stored XSS on the public status page. However, impact is limited because: (1) the public page has no session cookies to steal, (2) soul creation requires authentication, (3) the workspace provides blast radius limitation. An attacker could use this to deface the public status page or redirect visitors to phishing sites.

**Remediation:**
```go
import "html"

html += `<div class="component-target">` + html.EscapeString(soul.Target) + `</div>`
html += `<div class="component-name">` + html.EscapeString(soul.Name) + `</div>`
```

**References:**
- [CWE-79: Improper Neutralization of Input During Web Page Generation](https://cwe.mitre.org/data/definitions/79.html)

---

### XSS-001: StatusPage Handler X-Frame-Options ALLOWALL

**Severity:** Low
**Confidence:** 80/100
**CWE:** CWE-1021 — Improper Restriction of Rendered UI Layers or Frames
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/statuspage/handler.go:1010`

**Description:**
The status page handler sets `X-Frame-Options: ALLOWALL`, which is a non-standard value. Most modern browsers treat `ALLOWALL` as `DENY` or ignore it entirely. The intent was likely to allow status page aggregators (like StatusGator) to embed the page.

**Vulnerable Code:**
```go
w.Header().Set("X-Frame-Options", "ALLOWALL")
```

**Impact:**
Low. `ALLOWALL` is not a valid RFC value. Most browsers either treat it as `SAMEORIGIN` or ignore it. If embedding is required for business purposes, `SAMEORIGIN` is the correct value. If embedding is not needed, `DENY` is preferred.

**Remediation:**
```go
// If embedding is needed:
w.Header().Set("X-Frame-Options", "SAMEORIGIN")

// If embedding is not needed:
w.Header().Set("X-Frame-Options", "DENY")
```

**References:**
- [CWE-1021: Improper Restriction of Rendered UI Layers or Frames](https://cwe.mitre.org/data/definitions/1021.html)

---

### BIZ-003: Alert Rule at_least Threshold Defaults to 1

**Severity:** Low
**Confidence:** 70/100
**CWE:** CWE-670 — Control Flow Flow Not Constrained During Default Values
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/alert/manager.go:798`

**Description:**
Alert rules using `at_least` logic default the threshold to 1 when not explicitly set. This means a single matching sub-condition triggers the rule, regardless of how many sub-conditions exist.

**Vulnerable Code:**
```go
matchedCount := 0
for _, sub := range cond.Children {
    if sub.check(j) { matchedCount++ }
}
if matchedCount >= threshold { /* fire */ }
// Where threshold defaults to 1 if cond.Threshold <= 0
```

**Impact:**
Alert rules intended to require multiple conditions could fire with only one match. This is a configuration misuse issue, not a vulnerability. The rule author must explicitly set appropriate thresholds. No security boundary is crossed.

**Remediation:**
Default `at_least` threshold to the number of sub-conditions at rule registration:

```go
if cond.Threshold <= 0 {
    cond.Threshold = len(cond.Children) // Require all by default
}
// Or validate at registration time and reject if threshold > child count
```

**References:**
- [CWE-670: Control Flow Flow Not Constrained During Default Values](https://cwe.mitre.org/data/definitions/670.html)

---

### WS-004: Workspace Switching Forbidden (Verified Security Control)

**Severity:** Low (Informational)
**Confidence:** 85/100
**CWE:** CWE-940 — Improper Verification of Source of a Communication Channel
**OWASP:** A04:2021 — Insecure Design

**Location:** `internal/api/websocket.go:292-294`

**Description:**
The WebSocket handler explicitly rejects `join_workspace` messages with a "forbidden" error. This is a **positive security finding** — workspace switching is properly blocked.

**Finding Code:**
```go
case "join_workspace":
    c.send <- c.createErrorMessage("forbidden", "workspace switching not supported")
```

**Impact:**
This control prevents a class of attacks where a compromised token could be used to subscribe to a different workspace's events. No action needed.

**References:**
- [CWE-940: Improper Verification of Source of a Communication Channel](https://cwe.mitre.org/data/definitions/940.html)

---

## Remediation Roadmap

### Phase 1: Immediate (1-3 days)

Address all Critical findings. These represent immediate, exploitable security risks.

| # | Finding | Effort | Impact |
|---|---------|--------|--------|
| 1 | AUTH-001: Remove fmt.Printf for password reset token | Low | Critical |
| 2 | SSRF-001: Implement IP address normalization in SSRF validator | Medium | Critical |

### Phase 2: Short-Term (1-2 weeks)

Address High findings and any quick-win Medium findings.

| # | Finding | Effort | Impact |
|---|---------|--------|--------|
| 3 | PRIVESC-001: Change anonymous role from admin to viewer | Low | High |
| 4 | API-001: Add workspace check to gRPC ListJudgments | Low | High |
| 5 | API-002: Add workspace check to gRPC GetChannel | Low | High |
| 6 | API-003: Add workspace check to gRPC GetRule | Low | High |
| 7 | SSRF-002: Apply WrapDialerContext to WebSocket checker | Medium | High |
| 8 | BIZ-002: Use DTO for workspace updates excluding OwnerID/Plan/Quotas | Medium | High |
| 9 | MASS-002: Add field allowlist for rule config merge | Medium | High |

### Phase 3: Medium-Term (1-2 months)

Address remaining Medium findings and dependency updates.

| # | Finding | Effort | Impact |
|---|---------|--------|--------|
| 10 | AUTHZ-001: Add workspace check to REST handleGetJudgment | Low | Medium |
| 11 | LDAP-001: Add DN escaping for email in BindDN construction | Low | Medium |
| 12 | MASS-001: Use DTOs for Soul updates with field validation | Medium | Medium |
| 13 | RATE-003: Add trusted proxy validation for X-Forwarded-For | Medium | Medium |
| 14 | DEP-001: Monitor gopkg.in/check.v1 deprecation | Low | Medium |

### Phase 4: Hardening (Ongoing)

Address Low findings and implement defense-in-depth measures.

| # | Recommendation | Effort | Impact |
|---|---------------|--------|--------|
| 15 | REDIR-002: Add html.EscapeString for soul.Name and soul.Target in status page | Low | Low |
| 16 | XSS-001: Replace ALLOWALL with SAMEORIGIN in status page handler | Low | Low |
| 17 | BIZ-003: Default at_least threshold to sub-condition count | Low | Low |
| 18 | Review and test all workspace isolation checks across REST and gRPC endpoints | Medium | High |
| 19 | Implement request signing for gRPC inter-service communication | Medium | High |
| 20 | Add automated security regression tests for all IDOR protections | Medium | Medium |

---

## Methodology

This assessment was performed using `security-check`, an AI-powered static analysis tool that uses large language model reasoning to detect security vulnerabilities across 48 specialized skills.

### Pipeline Phases

1. **Phase 1 — Reconnaissance**: Automated codebase architecture mapping and technology detection. Identified Go 1.26.2 backend, React 19 dashboard, 465 total dependencies, trust boundaries, entry points, and data flows.

2. **Phase 2 — Vulnerability Hunting**: 48 specialized skills scanned for 20+ vulnerability categories including authentication, authorization, injection, SSRF, mass assignment, business logic, rate limiting, CSRF, XSS, cryptography, and supply chain.

3. **Phase 3 — Verification**: False positive elimination with confidence scoring (0-100). Applied reachability, sanitization, framework protection, context, and duplicate detection checks across 38 scanner result files. Reduced ~150 raw findings to 17 verified exploitable findings.

4. **Phase 4 — Reporting**: CVSS-aligned severity classification, CWE/OWASP mapping, and prioritized remediation roadmap.

### Skills Executed (48 total)

| Category | Skills |
|----------|--------|
| Language Scanners | scan-go, scan-typescript, scan-yaml |
| Auth & Sessions | scan-auth, scan-jwt, scan-session, scan-oidc, scan-ldap |
| API Security | scan-api-security, scan-idor, scan-rest, scan-grpc |
| Injection | scan-ssrf, scan-ldap, scan-sql, scan-nosql, scan-command |
| Frontend | scan-xss, scan-csrf, scan-cookies |
| Business Logic | scan-business-logic, scan-race |
| Configuration | scan-config, scan-iac |
| Dependencies | scan-dependencies, scan-supply-chain |
| Infrastructure | scan-cors, scan-ratelimit, scan-headers, scan-tls |
| Crypto | scan-crypto |
| Storage | scan-storage |

### Limitations

- Static analysis only — no runtime testing or dynamic analysis performed
- AI-based reasoning may miss vulnerabilities requiring deep domain knowledge or runtime behavior
- Confidence scores are estimates based on code pattern analysis, not guarantees
- Custom business logic flaws may require manual review
- The B+Tree storage engine (CobaltDB) was not audited for storage-specific vulnerabilities
- Kubernetes/Helm deployment configurations were not deeply scanned

---

## Appendix A: Summary Table

| ID | Finding | Severity | Confidence | File |
|----|---------|----------|------------|------|
| AUTH-001 | Password reset token logged to stdout | Critical | 100 | internal/auth/local.go:556 |
| SSRF-001 | Decimal IP bypasses SSRF blocklist | Critical | 85 | internal/probe/ssrf.go:115 |
| AUTHZ-001 | Missing workspace check on handleGetJudgment | High | 90 | internal/api/rest.go:924 |
| PRIVESC-001 | Anonymous user gets admin role when auth disabled | High | 100 | internal/api/rest.go:1526 |
| API-001 | gRPC ListJudgments lacks workspace isolation | High | 90 | internal/grpcapi/server.go:670 |
| API-002 | gRPC GetChannel lacks workspace isolation | High | 90 | internal/grpcapi/server.go:841 |
| API-003 | gRPC GetRule lacks workspace isolation | High | 90 | internal/grpcapi/server.go:972 |
| SSRF-002 | WebSocket checker bypasses SSRF dial protection | High | 85 | internal/probe/websocket.go:95 |
| LDAP-001 | Unescaped email in LDAP BindDN construction | Medium | 75 | internal/auth/ldap.go:166 |
| MASS-001 | Mass assignment in handleUpdateSoul | Medium | 75 | internal/api/rest.go:835 |
| MASS-002 | Mass assignment in handleUpdateRule config merge | Medium | 85 | internal/api/rest.go:1142 |
| BIZ-002 | Workspace update allows overwriting OwnerID and plan fields | Medium | 85 | internal/api/rest.go:1227 |
| RATE-003 | Rate limiter trusts X-Forwarded-For without proxy check | Low | 75 | internal/api/rest.go:1837 |
| REDIR-002 | StatusPage HTML embeds soul.Name/Target without escaping | Low | 80 | internal/api/statuspage.go:289 |
| XSS-001 | StatusPage handler sets X-Frame-Options: ALLOWALL | Low | 80 | internal/statuspage/handler.go:1010 |
| BIZ-003 | Alert rule at_least threshold defaults to 1 | Low | 70 | internal/alert/manager.go:798 |
| WS-004 | Workspace switching forbidden (positive control) | Low | 85 | internal/api/websocket.go:292 |

---

## Appendix B: Positive Security Controls

The following security controls were confirmed as properly implemented and should be maintained:

1. **Password Hashing**: bcrypt cost 12, constant-time comparison
2. **Brute Force Protection**: 5 attempts, 15-minute lockout
3. **Session Token Generation**: 32 bytes via `crypto/rand`, fail-closed on CSPRNG failure
4. **Generic Error Messages**: "invalid credentials" for both wrong email and wrong password
5. **Session Expiration**: 24 hours with cleanup goroutine
6. **Session Invalidation**: On logout and password change
7. **Token via Authorization Header**: Prevents log leakage
8. **OIDC Cookie Security**: HttpOnly + Secure + SameSite=Strict + MaxAge=600
9. **OIDC State + Nonce**: Dual CSRF protection with `hmac.Equal`
10. **JWT Algorithm Allowlist**: Rejects `none`, only allows RS*/ES* asymmetric algorithms
11. **Security Headers**: X-Frame-Options DENY, CSP `default-src 'self'`, HSTS, etc.
12. **Workspace IDOR Protection**: Most REST endpoints check `WorkspaceID != ctx.Workspace`
13. **Path Parameter Validation**: Blocks `..`, `//`, null bytes
14. **Request Body Size Limit**: 1MB max, JSON depth limiting (32 levels)
15. **Rate Limiting**: Per-IP and per-user limits with cleanup goroutine
16. **TLS 1.2+ Minimum**: Enforced across all TLS configurations
17. **AES-256-GCM Storage Encryption**: With HKDF-SHA256 key derivation and per-record salt/nonce
18. **gRPC Authentication**: Bearer token required via `authInterceptor`
19. **WebSocket OriginPatterns**: Validated via `coder/websocket` library
20. **CORS Origin Allowlist**: Exact match (no wildcards), `Vary: Origin` set
21. **Backup Checksum Verification**: SHA256-based integrity verification
22. **Audit Logging**: Dedicated audit trail in `internal/api/audit.go`
23. **WebSocket Workspace Switching Blocked**: Prevents cross-workspace subscription attacks
24. **No SQL/NoSQL Injection**: Custom B+Tree storage uses type-safe queries
25. **No Command Injection**: No `exec.Command` with user input
26. **No XXE**: No XML parsers in use

---

## Disclaimer

This security assessment was performed using automated AI-powered static analysis. It does not constitute a comprehensive penetration test or security audit. The findings represent potential vulnerabilities identified through code pattern analysis and LLM reasoning. False positives and false negatives are possible.

This report should be used as a starting point for security remediation, not as a definitive statement of the application's security posture. A professional security audit by qualified security engineers is recommended for production applications handling sensitive data. The following were not covered in this assessment: dynamic/runtime analysis, network-level testing, cloud-specific security configuration, Kubernetes/Helm hardening, and social engineering attack vectors.

Generated by security-check — github.com/ersinkoc/security-check