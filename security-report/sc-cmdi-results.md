# Command Injection Scan Results (sc-cmdi)

**Scanner:** sc-cmdi  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/ |
| Vulnerabilities Found | 0 |
| Informational Findings | 0 |
| Confidence | High |

---

## Analysis

### Command Execution Patterns

All `exec.Command` usages found in the codebase are in **test files** and use **static arguments**:

**Test files (all safe):**
- `cmd/anubis/cluster_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/backup_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/judge_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/init_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/main_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/system_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`
- `cmd/anubis/soul_test.go` - Uses `exec.Command(os.Args[0], "-test.run=...")`

All test usages call the Go test binary with no user-controlled arguments.

### No shell=True Usage

No instances of:
- `shell=True` (Python subprocess)
- `exec.Command("sh", "-c", ...)` (shell invocation)
- String interpolation into command strings passed to `exec.Command()`

### Probe Engine

The probe engine (`internal/probe/`) performs health checks but does NOT use command execution:
- HTTP checks use `net/http`
- TCP/UDP checks use `net.Dial`
- DNS checks use `net.Resolver`
- ICMP checks use `syscall` or external libraries
- SMTP/IMAP use `net.TextProtoClient`
- gRPC uses `google.golang.org/grpc`
- WebSocket uses `nhooyr.io/websocket` or `github.com/coder/websocket`
- TLS checks use `crypto/tls`

### Findings

**NO COMMAND INJECTION VULNERABILITIES FOUND**

All `exec.Command` calls use static arguments from constants or test infrastructure. No user-controlled input reaches any command execution function.

### Verdict

**CLEAN - No command injection attack surface**

---

## Common False Positives Eliminated

1. **Test infrastructure** - `os.Args[0]` with `-test.run=` is test runner, not user input
2. **Static arguments** - No `exec.Command` with string concatenation from user input
3. **No shell=True** - Go's `exec.Command` does not invoke shell unless explicitly called with `sh -c`
4. **Probe checks** - Network checks use standard Go libraries, not shell commands

## References

- CWE-78: OS Command Injection
- https://cwe.mitre.org/data/definitions/78.html