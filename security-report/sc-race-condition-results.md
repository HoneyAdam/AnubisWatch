# Race Conditions / TOCTOU — Security Check Results

**Scanner:** sc-race-condition v1.0.0  
**Target:** `internal/storage/`, `internal/cluster/`  
**Findings:** 4

---

## Finding: RACE-001

- **Title:** Concurrent read-modify-write on `CobaltDB.GetSoulNoCtx` soul iteration
- **Severity:** High
- **Confidence:** 75
- **File:** `internal/storage/engine.go:887`
- **Vulnerability Type:** CWE-362 (Race Condition)
- **Description:** `GetSoulNoCtx` at line 887 iterates over hardcoded workspace prefixes (`["default/souls/"]`) with `db.data.mu.RLock()` held. However, `db.ListSouls` at line 731 acquires souls first via `db.PrefixScan` (which uses `db.data.mu.RLock()`), then iterates over results — the map returned from `PrefixScan` is a shared reference to `result` that could be concurrently modified if another goroutine inserts/deletes souls during iteration. The iteration pattern `for key, data := range results` iterates over the map returned by `PrefixScan` which is protected by `db.data.mu.RLock()` but the map itself could have been created from a point-in-time snapshot — not necessarily thread-safe for iteration if modified.
- **Impact:** Concurrent soul creation/deletion during a `ListSouls` call could cause the iteration to see partially modified data or miss souls, leading to inconsistent query results or potential panics from concurrent map access.
- **Remediation:** Ensure that any iteration over maps returned from `PrefixScan` or `RangeScan` is done on a defensive copy. Alternatively, ensure all reads that return maps perform a deep copy of keys/values before releasing the read lock.
- **References:** https://cwe.mitre.org/data/definitions/362.html

---

## Finding: RACE-002

- **Title:** TOCTOU in `GetSoulNoCtx` — workspace prefix list is hardcoded and incomplete
- **Severity:** Medium
- **Confidence:** 70
- **File:** `internal/storage/engine.go:888`
- **Vulnerability Type:** CWE-367 (TOCTOU)
- **Description:** `GetSoulNoCtx` at line 888 iterates over hardcoded prefixes `["default/souls/"]` to find a soul by ID. If a soul is migrated between workspaces (or a new workspace is created), the hardcoded prefix list would not include the new workspace path. This creates a window where a soul exists but is not findable via `GetSoulNoCtx` — this is effectively a TOCTOU race where the existence check depends on a static list of prefixes rather than the actual state of storage.
- **Impact:** Soul lookup failures for souls in non-default workspaces via the REST API's `GetSoulNoCtx` path. This could cause 404 errors for valid resources or inconsistent behavior depending on which storage path is used.
- **Remediation:** Use a consistent key scheme where soul keys include workspace in the key itself (already done for `SaveSoul`), and always construct the key directly rather than scanning prefixes. Alternatively, maintain an index of soul IDs to their workspace.
- **References:** https://cwe.mitre.org/data/definitions/367.html

---

## Finding: RACE-003

- **Title:** B+Tree insert is not fully atomic — split operation releases and re-acquires lock
- **Severity:** Medium
- **Confidence:** 70
- **File:** `internal/storage/engine.go:446`
- **Vulnerability Type:** CWE-362 (Race Condition)
- **Description:** `btreeIndex.insert` at line 396 acquires `idx.root.insertNonFull` within `db.data.mu.Lock()`. However, `insertNonFull` at line 446 can call `splitChild` which itself manipulates the tree. The split operation at line 462 checks `if len(n.children[idx].keys) >= order-1` and then calls `splitChild` — if two concurrent inserts both hit the same full node simultaneously, both could pass the condition and split, causing tree corruption.
- **Impact:** Under high write concurrency to the same B+Tree node, concurrent splits could corrupt the tree structure, leading to data loss, missed keys, or incorrect reads.
- **Remediation:** The `splitChild` operation must be atomic with the check, or a write-write conflict resolution mechanism should be in place. Consider using a per-node lock in addition to the global tree lock, or use lock coupling (hand-over-hand locking) during splits.
- **References:** https://cwe.mitre.org/data/definitions/362.html

---

## Finding: RACE-004

- **Title:** Session file write is not fully atomic — `os.Chmod` between write and rename
- **Severity:** Low
- **Confidence:** 80
- **File:** `internal/auth/local.go:187`
- **Vulnerability Type:** CWE-362 (Race Condition)
- **Description:** `saveSessionsLocked` at line 187 writes JSON to a temp file, calls `os.Chmod(tmpPath, 0600)`, then calls `os.Rename(tmpPath, a.sessionPath)`. On POSIX systems, there is a window between `Chmod` and `Rename` where another process could theoretically open the temp file. Additionally, on some systems, `Rename` is not atomic with respect to the target's existing permissions. The code comment explicitly notes "defense in depth" for this pattern.
- **Impact:** Low — requires an attacker with filesystem access to the session directory to exploit. The `os.WriteFile` with 0600 permissions already sets restrictive permissions immediately; the additional `Chmod` call is a hardening measure. However, on Windows the permission model differs, and this code path (lines 183-197) is largely Unix-specific.
- **Remediation:** The current implementation is adequate for the threat model. Consider using `os.Link` or `os.Rename` with an atomic overlay filesystem operation on platforms that support it. On Windows, ensure the data directory is created with restrictive ACLs (0700) at initialization.
- **References:** https://cwe.mitre.org/data/definitions/362.html