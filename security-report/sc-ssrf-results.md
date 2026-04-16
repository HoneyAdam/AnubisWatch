# SSRF Security Scan Results

**Scanner:** sc-ssrf
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch
**Focus:** internal/probe/, internal/api/
**Severity Classification:** Critical | High | Medium | Low

---

## Summary

| Finding ID | Title | Severity | Confidence |
|------------|-------|----------|------------|
| SSRF-001 | Decimal/hex/octal IP bypass in SSRF validator | Medium | 70% |
| SSRF-002 | WebSocket checker bypasses SSRF dial protection | Low | 85% |
| SSRF-003 | Port parsing failure silently ignored in ValidateAddress | Low | 75% |

**Risk Rating:** LOW - The codebase has strong SSRF protection with blocklists for cloud metadata and private IP ranges. However, there are bypass techniques and gaps that should be addressed.

---

## Finding SSRF-001: Decimal/Hex/Octal IP Bypass in SSRF Validator

**Severity:** Medium
**Confidence:** 70%
**File:** internal/probe/ssrf.go:162
**CWE:** CWE-918 (Server-Side Request Forgery)
**File References:**
- `internal/probe/ssrf.go:162` (isBlockedIP function)
- `internal/probe/ssrf.go:115` (net.ParseIP call)

### Description

The `isBlockedIP()` function uses `net.ParseIP()` to parse IP addresses. While `net.ParseIP()` correctly handles IPv4 dotted-decimal notation (e.g., `127.0.0.1`) and IPv6 hex notation, it does **not** handle alternative IP representations such as:

- **Decimal:** `2130706433` (127.0.0.1)
- **Hex:** `0x7F000001` (127.0.0.1)
- **Octal:** `0177.0.0.01` (127.0.0.1)
- **Dword:** `3232235521` (192.168.0.1)
- **IPv6-mapped IPv4:** `::ffff:127.0.0.1`

An attacker who can control the target URL could bypass the private IP blocklist by using decimal notation for `169.254.169.254` (AWS metadata) or private IP ranges.

### Code Analysis

```go
// internal/probe/ssrf.go:115-121
ip := net.ParseIP(host)
if ip != nil {
    if v.isBlockedIP(ip) {  // Only called if net.ParseIP succeeds
        return fmt.Errorf("target IP %q is blocked", ip)
    }
    return nil
}
// If net.ParseIP returns nil (unparseable), hostname resolution is attempted
// An attacker could use decimal IP like "2130706433" which net.ParseIP rejects
```

The flow is:
1. `ValidateTarget()` parses URL at line 91
2. Extracts hostname at line 104
3. Attempts `net.ParseIP(host)` at line 115
4. If `net.ParseIP` returns `nil` (for alternative formats like decimal), falls through to DNS resolution at line 124
5. DNS resolution succeeds for attacker-controlled domain

### Impact

An attacker who can configure a monitoring target (e.g., via soul creation API) could use decimal IP notation to bypass SSRF protections and probe:
- AWS/GCP/Azure cloud metadata endpoints
- Internal network services
- Port scanning internal infrastructure

### Remediation

Implement IP address normalization before blocking:

```go
func (v *SSRFValidator) normalizeIP(input string) (net.IP, bool) {
    // Try standard parse first
    if ip := net.ParseIP(input); ip != nil {
        return ip, true
    }
    
    // Try decimal (dword) notation
    if parsed, err := strconv.ParseUint(input, 10, 32); err == nil {
        ip := net.IPv4(byte(parsed>>24), byte(parsed>>16), byte(parsed>>8), byte(parsed))
        return ip, true
    }
    
    // Try hex notation
    if parsed, err := strconv.ParseUint(strings.TrimPrefix(input, "0x"), 16, 32); err == nil {
        ip := net.IPv4(byte(parsed>>24), byte(parsed>>16), byte(parsed>>8), byte(parsed))
        return ip, true
    }
    
    return nil, false
}
```

Then replace `net.ParseIP(host)` with `normalizeIP()`.

### References

- https://cwe.mitre.org/data/definitions/918.html
- https://blog.safebuff.com/2016/07/06/SSRF-Notes/#bypass-techniques

---

## Finding SSRF-002: WebSocket Checker Bypasses SSRF Dial Protection

**Severity:** Low
**Confidence:** 85%
**File:** internal/probe/websocket.go:95-103
**CWE:** CWE-918 (Server-Side Request Forgery)
**File References:**
- `internal/probe/websocket.go:95-103` (direct TLS.Dial)
- `internal/probe/ssrf.go:255` (WrapDialer - unused by WebSocket)

### Description

The WebSocket checker (`internal/probe/websocket.go`) uses `tls.DialWithDialer` and `net.DialTimeout` directly for establishing connections, bypassing the SSRF protection provided by `SSRFValidator.WrapDialer()` and `WrapDialerContext()`.

The HTTP checker uses `WrapDialerContext` indirectly through the HTTP transport (via `getTransport()`), which applies DNS re-resolution protection on every connection attempt. However, the WebSocket checker makes direct dial calls at lines 100 and 102:

```go
// internal/probe/websocket.go:100
conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", host, tlsConfig)
// internal/probe/websocket.go:102
conn, err = net.DialTimeout("tcp", host, timeout)
```

These direct dials do not benefit from:
1. DNS rebinding protection (re-resolving hostname before connect)
2. Blocked IP validation on the resolved addresses
3. The `SSRFValidator` checks applied by `WrapDialer`

### Impact

A DNS rebinding attack could potentially bypass SSRF protections for WebSocket checks. An attacker could:
1. Register a domain that resolves to a public IP initially
2. After validation passes, have the DNS point to a private/internal IP (e.g., `169.254.169.254`)
3. When the WebSocket checker connects, it uses the already-resolved IP from the hostname

HTTP/WebSocket targets are typically user-configured (souls), so the attacker needs to modify an existing soul target to exploit this.

### Remediation

Apply the same `WrapDialerContext` protection to the WebSocket checker:

```go
// In WebSocketChecker.Judge(), create dialer with SSRF protection
validator := DefaultValidator  // or create instance
dialer := &net.Dialer{Timeout: timeout}
wrappedDial := validator.WrapDialerContext(dialer.DialContext)

if u.Scheme == "wss" {
    conn, err = tls.DialWithDialer(wrappedDial, "tcp", host, tlsConfig)
} else {
    conn, err = wrappedDial(ctx, "tcp", host)
}
```

### References

- `internal/probe/ssrf.go:255-284` (WrapDialer implementation)
- `internal/probe/ssrf.go:287-314` (WrapDialerContext implementation)

---

## Finding SSRF-003: Port Parsing Failure Silently Ignored

**Severity:** Low
**Confidence:** 75%
**File:** internal/probe/ssrf.go:192-198
**CWE:** CWE-918 (Server-Side Request Forgery)
**File References:**
- `internal/probe/ssrf.go:192` (SplitHostPort)
- `internal/probe/ssrf.go:198` (ignored port)

### Description

In `ValidateAddress()`, when `net.SplitHostPort()` fails, the code falls back to treating the entire input as a hostname, but **silently discards the port**:

```go
// internal/probe/ssrf.go:186-198
func (v *SSRFValidator) ValidateAddress(address string) error {
    if address == "" {
        return fmt.Errorf("address is empty")
    }

    host, port, err := net.SplitHostPort(address)
    if err != nil {
        // Try without port
        host = address
    }

    _ = port // Port validation can be added here if needed  <-- LINE 198
    // ... continues with host-only validation
```

### Analysis

This creates a minor security gap:
1. If an address like `10.0.0.1:INVALID_PORT` is passed, `SplitHostPort` fails
2. The code falls back to `host = address` (the entire string including port)
3. Only `host` is validated (blocked IPs, etc.)
4. Port is never validated (range check, well-known port check, etc.)

While the current impact is low (the host is still validated), this could become a problem if additional port-based validation is added later without addressing this code path.

### Remediation

Add logging when port parsing fails to detect potential attack attempts:

```go
if err != nil {
    slog.Debug("port parsing failed, treating entire address as host", 
        "address", address, "error", err)
    host = address
}
```

Or implement proper port validation:

```go
if err != nil {
    host = address
} else {
    // Validate port if needed
    portNum, _ := strconv.Atoi(port)
    if portNum < 1 || portNum > 65535 {
        return fmt.Errorf("invalid port number: %d", portNum)
    }
}
```

---

## Positive Security Findings

The SSRF protection in this codebase is **strong** overall:

### Strengths

1. **Comprehensive blocklist** (`internal/probe/ssrf.go:14-64`):
   - AWS metadata endpoints (169.254.169.254)
   - GCP metadata (metadata.google.internal)
   - Azure, DigitalOcean, Alibaba Cloud, Oracle Cloud metadata
   - All private IP ranges (RFC 1918, RFC 3927, etc.)
   - Multicast, broadcast, reserved ranges

2. **DNS rebinding protection** (`internal/probe/ssrf.go:271-280`):
   ```go
   // Re-resolve hostname and validate all IPs (prevents DNS rebinding)
   addrs, err := net.LookupHost(host)
   ```

3. **Redirect validation** (`internal/probe/http.go:199-204`):
   ```go
   client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
       // SSRF check on every redirect target
       redirectURL := req.URL.String()
       if err := ValidateTarget(redirectURL); err != nil {
           return fmt.Errorf("redirect target blocked by SSRF: %w", err)
       }
   }
   ```

4. **Scheme allowlist** (`internal/probe/ssrf.go:97-102`):
   ```go
   switch u.Scheme {
   case "http", "https", "ws", "wss", "grpc", "tcp", "udp":
       // Allowed
   default:
       return fmt.Errorf("URL scheme %q is not allowed", u.Scheme)
   }
   ```

5. **Unresolvable hostname blocking** (`internal/probe/ssrf.go:126-128`):
   ```go
   if err != nil {
       // If we can't resolve, we can't validate - block it
       return fmt.Errorf("cannot resolve hostname %q: %w", host, err)
   }
   ```

6. **Test coverage** (`internal/probe/ssrf_test.go`): Comprehensive tests for validator behavior.

---

## Recommendations

1. **High Priority:** Add IP address normalization to handle decimal/hex/octal representations (SSRF-001)
2. **Medium Priority:** Apply `WrapDialerContext` to WebSocket checker (SSRF-002)
3. **Low Priority:** Add logging for port parsing failures (SSRF-003)

---

*Generated by sc-ssrf security scanner*