# Security Check: Cryptography Misuse (sc-crypto)

## Summary
**Result: PASS**

No misuse of cryptography found. The codebase uses strong, modern cryptographic primitives throughout. TLS 1.2+ is enforced, AES-256-GCM is used for storage encryption with HKDF key derivation, bcrypt is used for password hashing, and HMAC-SHA256 is used for OIDC state signing. No use of weak algorithms (MD5, SHA1, DES, RC4) was found in production code.

---

## Scan Scope
- Directories: `internal/auth/`, `internal/api/`, `internal/storage/encryption.go`, `cmd/anubis/`, `internal/probe/tls.go`
- Patterns: `crypto/md5`, `crypto/sha1`, `des`, `rc4`, `InsecureSkipVerify`, `bcrypt`, `crypto/hmac`, `crypto/rand`, TLS version constraints

---

## Cryptographic Implementations

### Password Hashing — PASS

**File:** `internal/auth/local.go`

- Uses `golang.org/x/crypto/bcrypt` for password hashing (lines: `17`, `76`, `296`, `484`, `518`, `540`, `582`)
- `bcryptCost` is configurable (not hardcoded to `bcrypt.DefaultCost` in production)
- Constant-time comparison via `bcrypt.CompareHashAndPassword` for login (line ~490)
- **Timing attack protection:** Uses a dummy hash with `bcrypt.DefaultCost` on authentication failure to prevent timing oracles (lines `296`, `540`)

```go
// internal/auth/local.go:296
dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
```

This is a strong pattern — the dummy hash prevents an attacker from distinguishing "user not found" from "wrong password" based on response time.

### Storage Encryption — PASS (EXCELLENT)

**File:** `internal/storage/encryption.go`

Uses **AES-256-GCM** (strongest symmetric encryption available in Go's stdlib):
- `aes.NewCipher(derivedKey)` → `cipher.NewGCM(block)` (lines `49-54`)
- Key derivation uses **HKDF-SHA256** with a random salt per encryption operation (lines `38-44`, `64-67`)
- Each encryption generates a fresh random 12-byte nonce via `crypto/rand` (lines `85-88`)
- Format: `[32-byte salt][12-byte nonce][ciphertext + 16-byte GCM auth tag]`
- Salt, nonce, and ciphertext are all cryptographically random

```go
// internal/storage/encryption.go:40
hkdfReader := hkdf.New(sha256.New, e.masterKey, salt, []byte(encryptionKDFInfo))
```

The HKDF domain separation constant `"anubiswatch-storage-encryption-v1"` prevents cross-context key reuse.

### OIDC State Signing — PASS

**File:** `internal/auth/oidc.go:108-150`

- Generates 32-byte random HMAC key via `crypto/rand` on startup (line `110`)
- State signed with **HMAC-SHA256** (`hmac.New(sha256.New, o.stateHMAC)`) (lines `128`, `142`)
- Signature verification uses `hmac.Equal()` for constant-time comparison (line `146`)
- Algorithm restriction: only `RS256` and `ES256` JWT algorithms allowed (line `556` — rejects `alg: "none"`)

### TLS Configuration — PASS

**Files:**
- `internal/probe/tls.go:85` — `MinVersion: tls.VersionTLS12`
- `cmd/anubis/server.go:487` — `MinVersion: tls.VersionTLS12`
- `internal/cluster/manager.go:32` — `MinVersion: tls.VersionTLS12`

TLS 1.2 is enforced as the minimum across all TLS configurations. No TLS 1.0 or 1.1 usage.

### TLS Diagnostic with InsecureSkipVerify — GOOD PATTERN

**File:** `internal/probe/tls.go:302-317`

```go
tlsConfig := &tls.Config{
    InsecureSkipVerify: true, // Required to not fail handshake; we verify ourselves below
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        certs := make([]*x509.Certificate, len(rawCerts))
        for i, rawCert := range rawCerts {
            cert, err := x509.ParseCertificate(rawCert)
            if err != nil {
                return nil // Don't fail the handshake
            }
            certs[i] = cert
        }
        capturedCerts = certs
        return nil // Never fail the handshake
    },
}
```

This is the **correct pattern** for diagnostic TLS inspection. The `InsecureSkipVerify` is only set to capture the raw certificate chain; the code then parses and inspects the certificates manually. This is the same approach recommended in the Go standard library documentation for diagnostic use cases and is marked with a comment referencing HIGH-06.

### Random Number Generation — PASS

**Files:** `internal/auth/local.go`, `internal/auth/oidc.go`, `internal/storage/encryption.go`, `internal/secrets/manager.go`, `cmd/anubis/init.go`, `cmd/anubis/server.go`

All use `crypto/rand` (not `math/rand`) for cryptographic operations:
- Session token generation (local.go:357, 368)
- OIDC state and nonce generation (oidc.go:110)
- Encryption salt and nonce (encryption.go:65, 86)
- Admin password generation (server.go:401)
- Secure password generation for init (init.go:437)

No weak random number generation found.

### InsecureSkipVerify in Test Files — OK (Test Only)

Multiple test files use `InsecureSkipVerify: true` for testing against self-signed certificates:
- `internal/probe/grpc_test.go:661,721,988,1008,1048,1079,1112,1145`
- `internal/probe/http_test.go:555,590`
- `internal/probe/smtp_test.go:122,494,591,615`
- `internal/probe/websocket_test.go:180,219,1171`
- `internal/probe/tls_test.go:511,998`

**All are in `*_test.go` files**, which is acceptable — tests against self-signed certs require this. No production code uses `InsecureSkipVerify` for anything other than the diagnostic pattern described above.

### JWT Algorithm Validation — PASS

**File:** `internal/auth/oidc.go:556`

```go
return nil, fmt.Errorf("JWT signing algorithm %q not allowed: asymmetric signature required", alg)
```

The code rejects the `alg: "none"` attack (confirmed by test at `oidc_test.go:1594`) and requires asymmetric algorithms (RS256, ES256). HMAC (symmetric) algorithms are also rejected, which is correct for an OIDC context where the public key is fetched from the IdP.

### No Weak Hash Algorithms — PASS

No use of MD5, SHA1, or other weak hash functions for security purposes was found in production code. SHA256 is used throughout (HMAC-SHA256 for OIDC state, SHA256 in HKDF for key derivation).

---

## Minor Observations

### Bcrypt Cost Configurable (Good)

**File:** `internal/auth/local.go`

The bcrypt cost is configurable via `bcryptCost` which defaults to `bcrypt.DefaultCost` (10) but can be increased. No issues found, but increasing to 12 or 13 would be more conservative.

### LDAP Filter Injection Protection (Good)

**File:** `internal/auth/ldap.go:97`

```go
filter = strings.ReplaceAll(filter, "{{mail}}", ldap.EscapeFilter(email))
```

Uses `ldap.EscapeFilter` to prevent LDAP injection. Good practice.

### No TLS 1.3 Specific Config (Info)

TLS 1.3 is supported by default in Go's `crypto/tls` package, but no explicit `MaxVersion: tls.VersionTLS13` is set. This is fine — it allows the server to negotiate TLS 1.3 automatically. No action needed.

---

## Conclusion

**Excellent cryptography hygiene.** No weak algorithms, no misuse of crypto primitives, strong defaults throughout. The only recommendation is to consider increasing bcrypt cost above the default if the deployment has headroom for the CPU cost.