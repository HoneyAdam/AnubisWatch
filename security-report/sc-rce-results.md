# sc-rce: Remote Code Execution — Results

**Skill:** sc-rce | **Scan Date:** 2026-04-16
**Files Scanned:** internal/probe/, internal/, cmd/anubis/

---

## Summary: NO RCE VULNERABILITIES FOUND

No dynamic code execution vectors (eval, exec, Function, ScriptEngine, plugin.Open, yaegi, ClassLoader.loadClass) were detected in the codebase.

---

## Detailed Scan Results

### eval/exec/exec variants
- **Pattern:** `eval(`, `exec(`, `DynamicMethod`, `yaegi`, `plugin.Open`, `ClassLoader.loadClass`, `scriptEngine`, `ScriptEngine`
- **Result:** No matches found across all Go files

### Probe Engine (internal/probe/)
The probe engine uses struct-based protocol checkers (HTTP, TCP, DNS, etc.) that perform defined network operations. No dynamic code evaluation:
- `internal/probe/http.go` — HTTP check via net/http
- `internal/probe/tcp.go` — TCP socket check
- `internal/probe/dns.go` — DNS lookup
- `internal/probe/grpc.go` — gRPC health check
- `internal/probe/websocket.go` — WebSocket handshake
- `internal/probe/tls.go` — TLS certificate check
- `internal/probe/smtp.go`, `internal/probe/imap.go` — Email protocol checks
- `internal/probe/icmp.go` — ICMP ping

### Journey Executor (internal/journey/executor.go)
Uses `json.Unmarshal` to parse HTTP response bodies for JSONPath extraction. This is safe JSON parsing, not code execution.

### Auth Modules (internal/auth/)
- `local.go` — Uses `bcrypt.GenerateFromPassword`, `crypto/rand`, `json.Marshal/Unmarshal`. No dynamic evaluation.
- `oidc.go` — OIDC JWT verification uses `crypto/ecdsa`, `crypto/rsa` signature verification.
- `ldap.go` — LDAP bind via `ldap.DialURL` and `conn.Bind()`.

---

## Verdict: Clean

**No RCE attack surface detected.** The codebase uses zero-dependency Go with no dynamic code evaluation, template injection, or unsafe deserialization for code execution.
