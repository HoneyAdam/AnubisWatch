# Business Logic Flaws — Security Check Results

**Scanner:** sc-business-logic v1.0.0  
**Target:** `internal/alert/`, `internal/cluster/`  
**Findings:** 5

---

## Finding: BIZ-001

- **Title:** Mass assignment on maintenance window update allows setting any struct field
- **Severity:** High
- **Confidence:** 85
- **File:** `internal/api/rest.go:2101`
- **Vulnerability Type:** CWE-840 (Business Logic Errors)
- **Description:** `handleUpdateMaintenanceWindow` at line 2101 binds the entire request body into a `map[string]interface{}` and then performs type-assertion updates for named fields. While this is safer than direct struct binding, the use of `ctx.Request.Body` → `json.NewDecoder().Decode(&input)` without a field allowlist means any JSON field is accepted. Critically, if a client sends `"enabled": true` it overwrites the field without validation of the field's relationship to other state.
- **Impact:** An authenticated attacker with `settings:write` permission could set arbitrary struct fields on a maintenance window, potentially enabling maintenance mode for all souls unexpectedly.
- **Remediation:** Use an explicit struct or DTO with only the mutable fields, and validate all field values before applying updates. Consider using `omitempty` JSON tags to only accept explicitly provided fields.
- **References:** https://cwe.mitre.org/data/definitions/840.html

---

## Finding: BIZ-002

- **Title:** Workspace update allows overwriting workspace ID and timestamps
- **Severity:** High
- **Confidence:** 80
- **File:** `internal/api/rest.go:1227`
- **Vulnerability Type:** CWE-840 (Business Logic Errors)
- **Description:** `handleUpdateWorkspace` at line 1227 binds the full `core.Workspace` struct via `ctx.Bind(&ws)`, then sets `ws.ID = id`. The `Workspace` struct likely contains fields like `CreatedAt`, `UpdatedAt`, `Plan`, `Settings`, or `OwnerID` that should not be user-modifiable. A malicious user could supply these fields in the JSON body and have them overwritten.
- **Impact:** Privilege escalation — a user with workspace update permissions could modify internal workspace metadata (e.g., plan tier, ownership, billing settings) that should require higher privileges.
- **Remediation:** Use a dedicated `UpdateWorkspaceRequest` DTO that only includes mutable fields (name, settings). Never bind directly to the domain struct for updates.
- **References:** https://cwe.mitre.org/data/definitions/840.html

---

## Finding: BIZ-003

- **Title:** Alert rule condition evaluation allows bypass via compound "at_least" logic
- **Severity:** Medium
- **Confidence:** 75
- **File:** `internal/alert/manager.go:798`
- **Vulnerability Type:** CWE-840 (Business Logic Errors)
- **Description:** `checkCompound` at line 798 uses `matchedCount >= threshold` for the `at_least` logic type. However, `threshold` defaults to `1` if `cond.Threshold <= 0`. This means a rule configured with `at_least` logic but no explicit threshold defaults to triggering if just 1 sub-condition matches, potentially bypassing the intended logic that requires multiple conditions to fire.
- **Impact:** Alert rules that are intended to require multiple conditions (e.g., AND logic or majority vote) could fire with only a single matching sub-condition, leading to alert fatigue or missed critical alerts depending on configuration.
- **Remediation:** Default `threshold` to the total number of sub-conditions for `at_least` (requiring all), or require explicit threshold configuration. Validate threshold against sub-condition count at rule registration time.
- **References:** https://cwe.mitre.org/data/definitions/840.html

---

## Finding: BIZ-004

- **Title:** Rate limit window reset is not atomic — entry is modified before window check completes
- **Severity:** Medium
- **Confidence:** 70
- **File:** `internal/alert/manager.go:565`
- **Vulnerability Type:** CWE-362 (Race Condition) / CWE-840 (Business Logic Errors)
- **Description:** `isRateLimited` at line 565 checks `if time.Since(entry.FirstSent) > channel.RateLimit.Window.Duration` and then resets the entry. However, between reading `entry` and acquiring the mutex, another goroutine could have already reset or updated the same entry. The check-then-act pattern with a non-atomic read under `defer m.history.Mu.Unlock()` creates a race where two concurrent alerts in the same window could both see the expired window and both increment without proper limiting.
- **Impact:** Under high concurrency, the rate limit per channel/soul may not be enforced correctly, allowing more alerts than intended through the rate limiter.
- **Remediation:** Move the entire rate limit check-and-update logic inside the `m.history.Mu.Lock()` section to make it atomic. The read-modify-write pattern must be protected by a single lock acquisition.
- **References:** https://cwe.mitre.org/data/definitions/362.html

---

## Finding: BIZ-005

- **Title:** Anomaly detection standard deviation computation is incorrect — uses wrong divisor
- **Severity:** Low
- **Confidence:** 85
- **File:** `internal/alert/manager.go:743`
- **Vulnerability Type:** CWE-840 (Business Logic Errors)
- **Description:** `checkAnomaly` at line 743 computes standard deviation with three sequential calculations that overwrite each other, with the final correct formula `stdDev = sqrt(sqSum/n)`. However, there are two intermediate incorrect calculations at lines 741-744 that are dead code and confusing. The final formula divides by `len(values)` (population std dev) rather than `len(values)-1` (sample std dev), which slightly underestimates variability with small sample sizes. More critically, the code comment at line 743 shows `// Fix: variance = sqSum/(n-1)` but then uses `sqSum/len(values)` (n, not n-1), indicating the intended sample std dev was not applied.
- **Impact:** With small historical datasets (the code requires only 3 events), the anomaly detection threshold bounds may be slightly off. This is a low-severity issue because the anomaly threshold also has a fallback to simple threshold checking.
- **Remediation:** Use the correct sample standard deviation formula consistently: `stdDev = sqrt(sqSum / float64(len(values)-1))` with proper handling for when `len(values) <= 1`. Or use the built-in `math` package's `sqrt` for clarity.
- **References:** https://cwe.mitre.org/data/definitions/840.html