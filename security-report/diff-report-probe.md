# Probe Security Diff Report

**Files Scanned:** `internal/probe/ssrf.go`, `internal/probe/websocket.go`
**Commit:** HEAD~1 to HEAD
**Date:** 2026-04-16

---

## Summary of Changes

### 1. SSRF-001 Fix: IP Normalization in ssrf.go
**Change:** Added `parseIP()` function to handle decimal IP notation (e.g., `2130706433` -> `127.0.0.1`) and updated all IP parsing call sites to use it instead of `net.ParseIP()`.

**Locations updated:**
- `ValidateTarget()` (line 116): `net.ParseIP(host)` -> `v.parseIP(host)`
- `WrapDialer()` (line 278): `net.ParseIP(host)` -> `v.parseIP(host)`
- `WrapDialer()` (line 291): `net.ParseIP(resolved)` -> `v.parseIP(resolved)`
- `WrapDialerContext()` (line 309): `net.ParseIP(host)` -> `v.parseIP(host)`
- `WrapDialerContext()` (line 321): `net.ParseIP(resolved)` -> `v.parseIP(resolved)`

**Verification:** PARTIAL

### 2. SSRF-002 Fix: WebSocket SSRF Protection in websocket.go
**Change:** Added `ValidateTarget()` call before connecting (line 92), and wrapped dialer with `DefaultValidator.WrapDialerContext()` for DNS rebinding protection (line 98).

**Locations updated:**
- Added `ValidateTarget(soul.Target)` validation check before connection
- Changed `net.DialTimeout("tcp", host, timeout)` to `dialCtx(context.Background(), "tcp", host)`
- Changed `tls.DialWithDialer(&net.Dialer{Timeout: timeout}, ...)` to `tls.DialWithDialer(dialer, ...)`

**Verification:** PARTIAL

---

## SSRF-001 Fix Verification

| Requirement | Status |
|-------------|--------|
| parseIP() function added | VERIFIED |
| parseIP() handles decimal IP (2130706433) | VERIFIED |
| parseIP() used in ValidateTarget | VERIFIED |
| parseIP() used in WrapDialer | VERIFIED |
| parseIP() used in WrapDialerContext | VERIFIED |

---

## NEW Issues Found

### Issue 1: Incomplete IP Parsing - Hex/Octal Not Implemented

**File:** `internal/probe/ssrf.go:140-153`
**Severity:** Medium
**Introduced by:** SSRF-001 fix

The `parseIP()` function comment states it supports three notations:

```go
// parseIP parses an IP address string, supporting decimal (2130706433),
// hex (0x7F000001), and octal (0177.0.0.01) notations in addition to
// standard dotted-decimal. Returns nil if the input cannot be parsed.
```

However, the implementation **only handles decimal**:

```go
func (v *SSRFValidator) parseIP(host string) net.IP {
    // Try standard parsing first
    if ip := net.ParseIP(host); ip != nil {
        return ip
    }
    // Try decimal: 2130706433 -> 127.0.0.1
    if parsed, err := strconv.ParseUint(host, 10, 32); err == nil {
        return net.IPv4(byte(parsed>>24), byte(parsed>>16), byte(parsed>>8), byte(parsed))
    }
    return nil  // <-- Hex and octal fall through to here and return nil
}
```

**Impact:** An attacker can still bypass SSRF protection using hex notation:
- `0xA9FEA9FE` = `169.254.169.254` (AWS metadata)
- `0x7F000001` = `127.0.0.1` (localhost)
- `0xC0A80101` = `192.168.1.1` (private)

**Example bypass:**
```
http://0xA9FEA9FE/latest/meta-data/  -> Decodes to 169.254.169.254
```

**Fix needed:** Implement hex and octal parsing to match documented behavior:
```go
// Try hex: 0x7F000001 -> 127.0.0.1
if strings.HasPrefix(host, "0x") || strings.HasPrefix(host, "0X") {
    if parsed, err := strconv.ParseUint(strings.TrimPrefix(host, "0x"), 16, 32); err == nil {
        return net.IPv4(byte(parsed>>24), byte(parsed>>16), byte(parsed>>8), byte(parsed))
    }
}
// Try octal: 0177.0.0.01 -> 127.0.0.1
// Note: Go's strconv doesn't auto-parse octal dotted notation, needs custom parsing
```

---

### Issue 2: wss:// Connections Do Not Use DNS Rebinding Protection

**File:** `internal/probe/websocket.go:104-109`
**Severity:** Low
**Introduced by:** SSRF-002 fix

The SSRF-002 fix correctly wraps the `ws://` dial with `WrapDialerContext`, but the `wss://` path still uses the bare `dialer`:

```go
dialer := &net.Dialer{Timeout: timeout}
dialCtx := DefaultValidator.WrapDialerContext(dialer.DialContext)

if u.Scheme == "wss" {
    // BUG: Uses bare dialer, NOT dialCtx (no DNS rebinding protection)
    conn, err = tls.DialWithDialer(dialer, "tcp", host, tlsConfig)
} else {
    // CORRECT: Uses wrapped dial with DNS rebinding protection
    conn, err = dialCtx(context.Background(), "tcp", host)
}
```

The `tls.DialWithDialer` is called with the raw `dialer`, not with `dialCtx`. This means:
- `ws://` targets: Protected by `WrapDialerContext` (DNS re-resolved before each connect)
- `wss://` targets: No DNS rebinding protection applied

**Impact:** A DNS rebinding attack could bypass SSRF for `wss://` WebSocket checks. However, `ValidateTarget()` is still called at line 92 before connecting, which validates the initial DNS resolution. The risk is lower but not zero.

**Fix needed:**
```go
if u.Scheme == "wss" {
    conn, err = tls.DialWithDialer(dialCtx, "tcp", host, tlsConfig)
} else {
    conn, err = dialCtx(context.Background(), "tcp", host)
}
```

---

## Pre-existing Issues (Not from this diff)

These issues existed before this commit and were not introduced by these changes:

1. **SSRF-003 (Port parsing failure ignored):** `ValidateAddress()` silently ignores port parsing failures and falls back to treating entire input as hostname.

2. **Hex IP notation bypass:** The `net.ParseIP()` function does handle some hex formats (`::ffff:127.0.0.1`), but not the `0x` prefix format commonly used in bypass attempts.

---

## Verdict

**FAIL** - SSRF-001 and SSRF-002 fixes are partially implemented but introduce one new issue and leave an existing gap unfixed.

| Fix | Status | Notes |
|-----|--------|-------|
| SSRF-001 | PARTIAL | Decimal IP normalization works, but hex/octal not implemented despite documentation claiming support |
| SSRF-002 | PARTIAL | ws:// protected, but wss:// bypasses DNS rebinding protection |

---

## Changes Verified

| Fix ID | Description | Status |
|--------|-------------|--------|
| SSRF-001 | parseIP() function added | VERIFIED |
| SSRF-001 | Decimal IP normalization in ValidateTarget | VERIFIED |
| SSRF-001 | Decimal IP normalization in WrapDialer | VERIFIED |
| SSRF-001 | Decimal IP normalization in WrapDialerContext | VERIFIED |
| SSRF-001 | Hex/octal support | NOT IMPLEMENTED (doc claims it exists) |
| SSRF-002 | ValidateTarget called before connect | VERIFIED |
| SSRF-002 | WrapDialerContext for ws:// | VERIFIED |
| SSRF-002 | WrapDialerContext for wss:// | NOT APPLIED |

---

## Recommendations

1. **High Priority:** Implement hex and octal IP parsing in `parseIP()` to match documented behavior
2. **Medium Priority:** Apply `dialCtx` to `wss://` connections for consistent DNS rebinding protection
3. **Low Priority:** Add SSRF-003 fix for port validation logging

---

*Generated by incremental security diff scan*