# Mass Assignment / Over-Posting — Security Check Results

**Scanner:** sc-mass-assignment v1.0.0  
**Target:** `internal/api/rest.go`  
**Findings:** 6

---

## Finding: MASS-001

- **Title:** Mass assignment in REST API `handleUpdateSoul` — full Soul struct bound without field filtering
- **Severity:** High
- **Confidence:** 90
- **File:** `internal/api/rest.go:835`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `handleUpdateSoul` at line 835 binds the entire request body into a `core.Soul` struct via `ctx.Bind(&soul)`, then copies all fields including potentially dangerous ones. The struct is then saved directly. While the handler sets `soul.ID = id` and `soul.WorkspaceID = ctx.Workspace` after binding (partially mitigating ID overwriting), any fields present in the JSON body that map to the `core.Soul` struct will be saved as-is. Fields like `Role`, `Plan`, `BillingInfo`, or internal flags could be set via the JSON body.
- **Impact:** A user with `souls:*` permission could send a crafted JSON body with fields like `"role": "admin"` or `"workspace_id": "another-workspace"` that get silently accepted and persisted, enabling privilege escalation or cross-workspace data access.
- **Remediation:** Replace `ctx.Bind(&soul)` with a DTO that only contains updateable fields (name, target, interval, timeout, tags, labels). Explicitly ignore WorkspaceID and ID from the request body. Use `json.Unmarshal` into a struct with `omitempty` and validate all fields.
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Finding: MASS-002

- **Title:** Mass assignment in REST API `handleUpdateRule` — arbitrary config fields merged
- **Severity:** High
- **Confidence:** 85
- **File:** `internal/api/rest.go:1142`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `handleUpdateRule` at line 1142 binds the request body to a `core.AlertRule` struct via `ctx.Bind(&rule)`, then if `req.Config != nil`, it iterates and merges ALL config fields: `m[k] = v` at line 1022 with no field allowlist. The rule's Config map could contain sensitive fields like `api_key`, `webhook_secret`, or internal routing flags that should not be user-modifiable via the API.
- **Impact:** A user with `rules:*` permission could send a crafted config update that overwrites sensitive rule settings, potentially escalating privileges or changing alert routing behavior in unexpected ways.
- **Remediation:** Use a field allowlist for Config keys in rule updates, similar to the pattern used in `UpdateChannel` at line 891. Define which config fields are mutable via the API and reject any others.
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Finding: MASS-003

- **Title:** Mass assignment in REST API `handleCreateSoul` — all Soul fields accepted on creation
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/api/rest.go:797`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `handleCreateSoul` at line 797 binds the full request body to `core.Soul` via `ctx.Bind(&soul)`, then overwrites `ID`, `WorkspaceID`, `CreatedAt`, `UpdatedAt`. Any fields in the JSON body that the `core.Soul` struct accepts will be persisted, including potentially sensitive fields like `Region`, `Weight` (interval), or internal metadata fields.
- **Impact:** An authenticated user with soul creation rights could specify fields like a specific `Region` they don't have access to, or set internal state flags that affect scheduling behavior.
- **Remediation:** Use a `CreateSoulRequest` DTO that only contains the fields users should be able to set at creation time. Validate all field values server-side (e.g., interval must be positive, target must be a valid URL).
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Finding: MASS-004

- **Title:** gRPC `UpdateRule` merges arbitrary config fields without allowlist
- **Severity:** High
- **Confidence:** 85
- **File:** `internal/grpcapi/server.go:1020`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `UpdateRule` at line 1020 merges `req.Config` into the existing rule's config map with no field filtering: `for k, v := range req.Config { m[k] = v }`. This is identical to the REST API vulnerability but in the gRPC path. Any arbitrary config key-value pairs can be injected into a rule's config.
- **Impact:** A gRPC client with rule update permissions could inject arbitrary config fields into a rule, potentially modifying sensitive routing or authentication settings.
- **Remediation:** Apply the same field allowlist pattern used in `UpdateChannel` at line 891 to `UpdateRule` as well. Define a set of allowed config field names and reject anything outside that set.
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Finding: MASS-005

- **Title:** Maintenance window update accepts arbitrary JSON fields without validation
- **Severity:** Medium
- **Confidence:** 80
- **File:** `internal/api/rest.go:2107`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `handleUpdateMaintenanceWindow` at line 2107 decodes the request body into `map[string]interface{}` and then updates struct fields via type assertions. While this approach is safer than struct binding (only known fields are applied), there is no validation of the incoming values beyond type checking. An attacker could send values that cause logical issues (e.g., negative durations, very long recurring patterns that cause DoS, etc.).
- **Impact:** A user with `settings:write` permission could set maintenance windows with invalid or extreme values (e.g., very long durations, empty soul lists, malformed recurring schedules) that could cause unexpected behavior in the scheduling engine.
- **Remediation:** Add explicit validation for each field after type assertion: validate duration ranges, string lengths, array lengths, and format any fields that have specific constraints (e.g., RFC3339 timestamps).
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Finding: MASS-006

- **Title:** Dashboard create/update allows arbitrary map fields without allowlist
- **Severity:** Low
- **Confidence:** 70
- **File:** `internal/api/rest.go:432`
- **Vulnerability Type:** CWE-915 (Improperly Controlled Modification of Dynamically-Determined Object Attributes)
- **Description:** `handleCreateDashboard` at line 432 binds `core.CustomDashboard` directly from the request body with no field filtering. The `CustomDashboard` struct likely contains a config map that accepts arbitrary key-value pairs. While dashboard modifications may be restricted to admin-level users, any custom dashboard field could be set at creation time.
- **Impact:** Lower severity since this is typically an admin-only operation. However, if a lower-privileged user somehow gains dashboard creation access (via misconfigured role), they could inject arbitrary data into dashboard configurations.
- **Remediation:** Use a DTO for dashboard creation that validates and restricts the fields that can be set, especially for the configuration map.
- **References:** https://cwe.mitre.org/data/definitions/915.html

---

## Positive Findings (Not Vulnerable)

- `handleUpdateChannel` (rest.go:1014) — Uses `ctx.Bind(&channel)` then overwrites `channel.ID` and `channel.WorkspaceID`, but the gRPC `UpdateChannel` (grpcapi/server.go:889) uses a field allowlist (`allowedConfigFields`) for config updates, which is the correct pattern.
- `handleCreateChannel` (rest.go:978) — Uses `ctx.Bind(&channel)` which could accept extra fields, but since it's a creation endpoint, the ID is regenerated server-side, limiting the impact.