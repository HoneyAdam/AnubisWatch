# gRPC API Security Diff Report

**Files Scanned:** `internal/grpcapi/server.go`, `internal/grpcapi/server_extra_test.go`
**Commit:** HEAD~1 to HEAD
**Date:** 2026-04-16

---

## Summary of Changes

### 1. ListJudgments (API-001 Fix)
**Change:** Added user extraction from context and soul ownership verification before listing judgments.

```go
user, ok := GetUserFromContext(ctx)
if !ok {
    return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}
// ...
if soulID != "" {
    soul, err := s.store.GetSoulNoCtx(soulID)
    if err == nil && soul != nil {
        if s, ok := soul.(*core.Soul); ok && s.WorkspaceID != "" && s.WorkspaceID != user.Workspace {
            return nil, status.Error(codes.PermissionDenied, "access denied: soul belongs to another workspace")
        }
    }
}
```

**Verification:** PASS - Workspace check correctly uses `user.Workspace` from authenticated context.

---

### 2. GetChannel (API-002 Fix)
**Change:** Added user extraction and now passes `user.Workspace` to storage layer.

```go
user, ok := GetUserFromContext(ctx)
if !ok {
    return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}
ch, err := s.store.GetChannelNoCtx(req.Id, user.Workspace)
```

**Verification:** PASS - Now correctly uses `user.Workspace` instead of empty string.

---

### 3. GetRule (API-003 Fix)
**Change:** Added user extraction, passes `user.Workspace` to storage, and adds explicit workspace IDOR check.

```go
user, ok := GetUserFromContext(ctx)
if !ok {
    return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}
r, err := s.store.GetRuleNoCtx(req.Id, user.Workspace)
// ...
if rule, ok := r.(*core.AlertRule); ok && rule.WorkspaceID != "" && rule.WorkspaceID != user.Workspace {
    return nil, status.Error(codes.PermissionDenied, "access denied: rule belongs to another workspace")
}
```

**Verification:** PASS - Double-layer protection: (1) storage filtered by workspace, (2) explicit ownership check.

---

### 4. UpdateRule (MASS-002 Fix)
**Change:** Added workspace IDOR check and config field allowlist.

```go
user, ok := GetUserFromContext(ctx)
if !ok {
    return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}
existing, err := s.store.GetRuleNoCtx(req.Id, user.Workspace)
// ...
if existingWS, ok := m["workspace_id"].(string); ok && existingWS != "" && existingWS != user.Workspace {
    return nil, status.Error(codes.PermissionDenied, "access denied: rule belongs to another workspace")
}
// Config allowlist
allowedConfigKeys := map[string]bool{
    "channel_ids": true,
    "cooldown": true,
    "severity": true,
    "notification_delay": true,
    "recovery_delay": true,
    "aggregation_window": true,
}
```

**Verification:** PASS - Sensitive fields (`api_key`, `token`, `webhook_url`) correctly excluded from allowlist.

---

### 5. TLS Support (Non-security relevant)
Added TLS configuration support to gRPC server.

---

## NEW Issues Found

### None identified

All changes are security fixes - no new vulnerabilities introduced.

---

## Pre-existing Issues (Not from this diff)

1. **GetChannel explicit ownership check missing**: Unlike GetRule, GetChannel does not have an explicit workspace ownership check after retrieval. It relies solely on the storage layer's filtering via `user.Workspace`. This is not a regression from this diff.

2. **UpdateChannel/DeleteChannel missing user extraction**: These methods still do not extract user from context or check workspace ownership. This is a pre-existing gap, not introduced by this diff.

3. **UpdateRule re-fetch uses empty workspace**: Line 1085 calls `s.store.GetRuleNoCtx(req.Id, "")`. This is the same pattern used by UpdateSoul (line 639) and appears to be pre-existing design, not a new issue.

---

## Verdict

**PASS** - All security-relevant changes correctly implement workspace isolation and IDOR protection. No new vulnerabilities introduced by this diff. The four identified fixes (API-001, API-002, API-003, MASS-002) are properly implemented.

---

## Changes Verified

| Fix ID | Description | Status |
|--------|-------------|--------|
| API-001 | ListJudgments soul ownership check | VERIFIED |
| API-002 | GetChannel workspace passing | VERIFIED |
| API-003 | GetRule workspace check + ownership | VERIFIED |
| MASS-002 | UpdateRule config field allowlist | VERIFIED |