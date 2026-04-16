# REST/gRPC API Security — Security Check Results

**Scanner:** sc-api-security v1.0.0  
**Target:** `internal/api/rest.go`, `internal/grpcapi/server.go`  
**Findings:** 7

---

## Finding: API-001

- **Title:** gRPC `ListJudgments` does not enforce workspace isolation
- **Severity:** High
- **Confidence:** 90
- **File:** `internal/grpcapi/server.go:670`
- **Vulnerability Type:** CWE-284 (Improper Access Control)
- **Description:** `ListJudgments` at line 670 does not check the authenticated user's workspace. The function takes `soulID` as an optional filter but does not verify that the `soulID` belongs to the caller's workspace. Unlike `GetSoul` at line 534 which explicitly checks `soul.WorkspaceID != user.Workspace`, this endpoint allows listing judgments across all souls in the system (or for a specific soul that may belong to another workspace).
- **Impact:** An authenticated gRPC client could query judgments for any soul by ID, potentially exposing judgment data (which includes response bodies, TLS certificate info, and latency data) from souls in other workspaces they don't have access to.
- **Remediation:** Either filter `ListJudgmentsNoCtx` by the caller's workspace (ensuring the soulID belongs to the user's workspace), or require the caller to have cross-workspace access permissions explicitly.
- **References:** https://owasp.org/API-Security/

---

## Finding: API-002

- **Title:** gRPC `GetChannel` does not enforce workspace isolation
- **Severity:** High
- **Confidence:** 90
- **File:** `internal/grpcapi/server.go:841`
- **Vulnerability Type:** CWE-284 (Improper Access Control)
- **Description:** `GetChannel` at line 841 calls `s.store.GetChannelNoCtx(req.Id, "")` passing an empty workspace string. The `GetChannelNoCtx` method accepts a workspace parameter that is used for filtering — passing `""` means it may return a channel from any workspace. Unlike `ListChannels` which correctly uses `user.Workspace`, the `GetChannel` RPC does not enforce that the channel belongs to the caller's workspace.
- **Impact:** An authenticated gRPC client could retrieve alert channel configurations (which may contain webhook URLs, bot tokens, API keys, SMTP credentials) for channels in other workspaces.
- **Remediation:** Pass `user.Workspace` as the workspace parameter to `GetChannelNoCtx`, and verify the returned channel's WorkspaceID matches the caller's workspace before returning it.
- **References:** https://owasp.org/API-Security/

---

## Finding: API-003

- **Title:** gRPC `GetRule` does not enforce workspace isolation
- **Severity:** High
- **Confidence:** 90
- **File:** `internal/grpcapi/server.go:972`
- **Vulnerability Type:** CWE-284 (Improper Access Control)
- **Description:** `GetRule` at line 972 calls `s.store.GetRuleNoCtx(req.Id, "")` with an empty workspace string. The rule lookup ignores workspace membership, potentially exposing rule configurations (including alert routing, escalation policies, and notification channels) to users in other workspaces.
- **Impact:** An authenticated gRPC client could retrieve alert rule configurations from other workspaces, potentially revealing sensitive routing logic, escalation schedules, and integration details.
- **Remediation:** Pass `user.Workspace` to `GetRuleNoCtx` and verify the returned rule belongs to the caller's workspace before returning it.
- **References:** https://owasp.org/API-Security/

---

## Finding: API-004

- **Title:** gRPC `ListJourneyRuns` and `GetJourneyRun` lack workspace enforcement
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/grpcapi/server.go:1172`
- **Vulnerability Type:** CWE-284 (Improper Access Access)
- **Description:** `ListJourneyRuns` at line 1172 and `GetJourneyRun` at line 1196 do not verify that the journey belongs to the caller's workspace. `ListJourneyRuns` calls `ListJourneyRunsNoCtx(req.JourneyId, limit)` and `GetJourneyRun` calls `GetJourneyRunNoCtx(req.JourneyId, req.RunId)` — neither passes or verifies workspace membership. A user could query journey run data for journeys they don't own.
- **Impact:** An authenticated gRPC client could access journey execution history (including extracted variables, step results, and timing data) for journeys in other workspaces. Extracted variables could contain sensitive information like API keys or authentication tokens used in synthetic monitoring.
- **Remediation:** Before returning journey runs, verify that the journey belongs to the authenticated user's workspace. Look up the journey first and check workspace membership, or include workspace in the storage query.
- **References:** https://owasp.org/API-Security/

---

## Finding: API-005

- **Title:** gRPC reflection service (`reflection.Register`) is enabled in production
- **Severity:** Medium
- **Confidence:** 95
- **File:** `internal/grpcapi/server.go:117`
- **Vulnerability Type:** CWE-200 (Exposure of Sensitive Information to an Unauthorized Actor)
- **Description:** `reflection.Register(s.grpc)` at line 117 enables the gRPC reflection service in production. This allows any gRPC client to discover all service names, method names, and message types implemented by the server without authentication. While useful for development tools like `grpcurl`, it exposes the full internal API surface to unauthenticated clients (subject to gRPC authentication requirements).
- **Impact:** An attacker with a valid gRPC token could use reflection to discover all available RPCs including unary and streaming methods, then probe for additional vulnerabilities or enumerate internal service interfaces. This significantly reduces the attacker's effort in mapping the API surface.
- **Remediation:** Disable gRPC reflection in production builds. Use a build tag (`//go:build !production`) or configuration flag to conditionally enable `reflection.Register` only in non-production environments. Alternatively, remove it entirely and rely on the generated protobuf stubs for client tooling.
- **References:** https://owasp.org/API-Security/; https://cwe.mitre.org/data/definitions/200.html

---

## Finding: API-006

- **Title:** REST `handleGetJudgment` does not verify soul workspace membership
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/api/rest.go:924`
- **Vulnerability Type:** CWE-284 (Improper Access Control)
- **Description:** `handleGetJudgment` at line 924 retrieves a judgment by ID without verifying that the judgment belongs to a soul owned by the caller's workspace. It simply returns the judgment if found. Unlike `handleGetSoul` at line 820 which explicitly checks `soul.WorkspaceID != ctx.Workspace`, this endpoint allows accessing any judgment by ID.
- **Impact:** An authenticated user could access judgment records for souls in other workspaces by knowing or guessing judgment IDs, potentially exposing monitoring data and response details from cross-workspace resources.
- **Remediation:** Look up the associated soul for the judgment and verify its workspace matches the caller's workspace before returning the judgment.
- **References:** https://owasp.org/API-Security/

---

## Finding: API-007

- **Title:** REST `handleListAllJudgments` is a stub returning empty array
- **Severity:** Low
- **Confidence:** 95
- **File:** `internal/api/rest.go:934`
- **Vulnerability Type:** CWE-1000 (Code Not Reachable / Dead Endpoint)
- **Description:** `handleListAllJudgments` at line 934 immediately returns `ctx.JSON(http.StatusOK, []interface{}{})` — an empty array with no data. This endpoint exists in the route table at line 345 (`s.router.Handle("GET", "/api/v1/judgments", ...)`) but serves no purpose. More problematically, if this endpoint were ever implemented, it could be a BOLA vulnerability since it has no soul ID filter and would need workspace-level filtering.
- **Impact:** Informational — this endpoint always returns an empty list, so it does not expose data. However, the route exists and could be mistakenly implemented in the future without proper workspace isolation.
- **Remediation:** Either implement the endpoint with proper workspace-scoped pagination and soul filtering, or remove the route entirely. Document the intent if the endpoint is planned for future use.
- **References:** https://owasp.org/API-Security/

---

## Positive Findings (Secure Patterns)

- `handleGetSoul` (rest.go:820) — Correctly checks `soul.WorkspaceID != ctx.Workspace` before returning
- `handleGetChannel` (rest.go:996) — Checks workspace membership with fallback to storage
- `handleGetRule` (rest.go:1124) — Checks workspace membership
- `GetSoul` (grpcapi/server.go:534) — Verifies `s.WorkspaceID != user.Workspace` before returning
- `CreateSoul` (grpcapi/server.go:557) — Sets workspace from authenticated user, not from request
- `UpdateSoul` (grpcapi/server.go:586) — Verifies workspace before updating
- REST rate limit middleware (rest.go:1779) — Per-IP and per-user rate limiting with cleanup goroutine
- REST security headers middleware (rest.go:1730) — Adds HSTS, X-Frame-Options, CSP headers
- REST CORS (rest.go:1582) — Validates origin against configurable allowlist