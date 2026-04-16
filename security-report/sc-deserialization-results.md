# sc-deserialization: Insecure Deserialization — Results

**Skill:** sc-deserialization | **Scan Date:** 2026-04-16
**Files Scanned:** internal/storage/, internal/, cmd/anubis/

---

## Summary: NO INSECURE DESERIALIZATION FOUND

All serialization uses safe formats (JSON). No Python pickle, Java ObjectInputStream, PHP unserialize, .NET BinaryFormatter, Ruby Marshal, or unsafe YAML loading.

---

## Safe Format Usage Confirmed

| Location | Format | Purpose |
|----------|--------|---------|
| `internal/auth/local.go:134` | json.Unmarshal | sessions.json load |
| `internal/auth/local.go:171` | json.Marshal | sessions.json save |
| `internal/auth/oidc.go:546` | json.Unmarshal | JWT header parsing |
| `internal/auth/oidc.go:610` | json.Unmarshal | JWT claims parsing |
| `internal/api/rest.go:1378` | json.NewDecoder | config update |
| `internal/api/rest.go:2066` | json.NewDecoder + maxDepthReader | request binding |
| `internal/backup/manager.go:492` | json.Marshal | backup checksum |
| `internal/backup/manager.go:657` | json.MarshalIndent | backup archive |
| `internal/backup/manager.go:705` | json.Unmarshal | backup restore |
| `internal/journey/executor.go:471` | json.Unmarshal | HTTP response parsing |
| `internal/alert/dispatchers.go:106` | json.Marshal | alert payload |
| `internal/api/mcp.go:316` | json.Unmarshal | MCP params |

### Depth Limiting on JSON (DoS Protection)
- `internal/api/rest.go:21-24` — maxJSONDepth = 32 limits nesting depth
- `internal/api/rest.go:26-48` — maxDepthReader struct enforces depth limit during decoding

### Storage Engine (internal/storage/engine.go)
The CobaltDB B+Tree engine uses raw binary page format (fixed-size integer/byte array records). Not a serialization format that supports object instantiation.

### No Unsafe Formats Detected
- No pickle, gob, binary.Read/Write with interface types
- No yaml.Load or yaml.Unmarshal with unsafe loading
- No encoding/gob usage
- No encoding/xml unmarshaled into interface{}
- No protobuf with google.protobuf.Any or unknown types

---

## Verdict: Clean

**No deserialization vulnerabilities found.** JSON is used exclusively for all data serialization including session persistence, JWT parsing, backup archives, and API request bodies. The B+Tree storage uses raw binary format inherently safe against object injection attacks.
