# Insecure File Upload Security Scan Results

**Scanner:** sc-file-upload
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch
**Focus:** internal/api/rest.go, internal/backup/
**Severity Classification:** Critical | High | Medium | Low

---

## Summary

| Finding ID | Title | Severity | Confidence |
|------------|-------|----------|------------|
| UPLOAD-001 | No user-facing file upload endpoint found | N/A | N/A |
| UPLOAD-002 | Backup import lacks filename sanitization (defense-in-depth) | Low | 60% |

**Risk Rating:** N/A - **No file upload vulnerability.** The application does not expose file upload functionality to users. This is a positive security finding.

---

## Finding UPLOAD-001: No User-Facing File Upload Endpoint

**Status:** NOT A VULNERABILITY
**File References:**
- `internal/api/rest.go` (full file - no upload handlers)
- `internal/backup/manager.go` (backup operations are server-initiated)

### Analysis

After thorough analysis of the codebase, **no user-facing file upload functionality exists**:

1. **No multipart/form-data handlers:** The REST API (`internal/api/rest.go`) has no handlers accepting file uploads. Common upload patterns like multer, busboy, or manual `multipart.Reader` usage were searched and not found.

2. **Backup/restore is server-managed:** The backup system (`internal/backup/manager.go`) operates entirely server-side:
   - `Create()` generates backups internally with timestamps (line 249)
   - `Restore()` reads from internal `backupPath` parameter (line 273), not user input
   - `ImportFromTar()` / `ExportToTar()` work with io.Reader/io.Writer, not HTTP uploads

3. **Configuration import:** There is no endpoint to import configuration or soul definitions from uploaded files.

4. **Static assets:** Dashboard assets are embedded at compile time, not served from uploaded files.

### REST API Routes Analyzed

All routes in `setupRoutes()` (lines 295-450) were reviewed:
- Health/metrics endpoints
- Authentication endpoints
- CRUD for souls, channels, rules, workspaces, status pages, journeys, dashboards
- Cluster management
- MCP tools
- WebSocket/SSE

**No upload routes found.**

### Positive Security Controls

The absence of file upload functionality is a **strong security control** because it eliminates an entire vulnerability class (CWE-434 - Unrestricted Upload of Dangerous File Types).

---

## Finding UPLOAD-002: Backup Import Filename Validation (Defense-in-Depth)

**Severity:** Low
**Confidence:** 60%
**File:** internal/backup/manager.go:694
**CWE:** CWE-22 (Path Traversal)
**File References:**
- `internal/backup/manager.go:682` (ImportFromTar)
- `internal/backup/manager.go:694` (extension check only)

### Description

The `ImportFromTar()` function validates tar entry names by checking only the file extension:

```go
// internal/backup/manager.go:694
if filepath.Ext(header.Name) != ".json" {
    continue
}
```

This does not sanitize the full path of the tar entry. However, as noted in the analysis below, this is **not exploitable** because:

1. The tar entry data is read as raw bytes and parsed as JSON
2. It is not written to the filesystem
3. The import is used internally, not exposed to users

### Why This Is Not Exploitable

**The tar entries are not extracted to the filesystem:**

```go
// internal/backup/manager.go:699-707
data := make([]byte, header.Size)
if _, err := io.ReadFull(tr, data); err != nil {
    return err
}

var backup Backup
if err := json.Unmarshal(data, &backup); err != nil {
    return err  // Parsed as JSON, not written to file
}
```

Even if a malicious tar contained `../../../etc/passwd` as an entry name, the data is:
1. Read into memory (`data := make([]byte, header.Size)`)
2. Parsed as JSON (`json.Unmarshal(data, &backup)`)
3. Never written anywhere

**No RCE, LFI, or path traversal is possible.**

### Remediation (Optional Defense-in-Depth)

Add path traversal validation for belt-and-suspenders security:

```go
// internal/backup/manager.go:694
// Validate entry name doesn't contain path traversal
if strings.Contains(header.Name, "..") || strings.Contains(header.Name, "\\") {
    m.logger.Warn("rejecting tar entry with path traversal characters", 
        "entry", header.Name)
    continue
}
```

---

## Security Best Practices Observed

While no upload functionality exists, the codebase demonstrates good security practices relevant to file handling:

### 1. Request Body Size Limits

```go
// internal/api/rest.go:21
const (
    maxRequestBodySize = 1 << 20 // 1MB
)

// internal/api/rest.go:2063-2066
func (c *Context) Bind(v interface{}) error {
    c.Request.Body = http.MaxBytesReader(c.Response, c.Request.Body, maxRequestBodySize)
    return json.NewDecoder(&maxDepthReader{r: c.Request.Body}).Decode(v)
}
```

### 2. JSON Depth Limiting

```go
// internal/api/rest.go:22
maxJSONDepth = 32

// Prevents stack overflow from deeply nested JSON
type maxDepthReader struct {
    r io.Reader
    depth int
}
```

### 3. Path Parameter Validation

```go
// internal/api/rest.go:1754-1777
func (s *RESTServer) validatePathParams(handler Handler) Handler {
    return func(ctx *Context) error {
        for key, value := range ctx.Params {
            // Check for path traversal attempts
            if strings.Contains(value, "..") || strings.Contains(value, "//") {
                s.logger.Warn("Path traversal attempt detected", ...)
                return ctx.Error(http.StatusBadRequest, "Invalid path parameter")
            }
            // ...
        }
        return handler(ctx)
    }
}
```

### 4. Injection Pattern Detection

```go
// internal/api/rest.go:1701-1728
func containsInjectionPatterns(input string) bool {
    // Check for path traversal
    if strings.Contains(input, "../") || strings.Contains(input, "..\\") {
        return true
    }
    // Check for null bytes
    if strings.Contains(input, "\x00") {
        return true
    }
    // Check for SQL injection patterns
    // Check for XSS patterns (script tags, javascript:)
}
```

### 5. Backup File Validation

```go
// internal/backup/manager.go:597-599
func isBackupFile(name string) bool {
    // Must start with "anubis_"
    // Must have .json or .gz extension
    return len(name) > 7 && name[:7] == "anubis_" && 
           (filepath.Ext(name) == ".json" || filepath.Ext(name) == ".gz")
}

// Checksum verification prevents tampering
func (m *Manager) verifyChecksum(backup *Backup) error {
    // ...
    if expected != actual {
        return fmt.Errorf("checksum mismatch")
    }
}
```

---

## Recommendations

**No action required for file upload security.** The application wisely avoids user-facing file uploads entirely.

Optional improvements if file handling is added in the future:

1. **If adding upload support:**
   - Validate file type server-side (magic bytes, not just extension)
   - Store files outside web root
   - Randomize filenames
   - Set file size limits
   - Configure upload directory to disable script execution

2. **Defense-in-depth (low priority):**
   - Add path traversal check in `ImportFromTar` (UPLOAD-002)

---

*Generated by sc-file-upload security scanner*