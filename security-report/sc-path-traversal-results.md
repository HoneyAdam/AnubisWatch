# Path Traversal Security Scan Results

**Scanner:** sc-path-traversal
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch
**Focus:** internal/storage/, internal/backup/
**Severity Classification:** Critical | High | Medium | Low

---

## Summary

| Finding ID | Title | Severity | Confidence |
|------------|-------|----------|------------|
| PATH-001 | isWithinDirectory uses string prefix without path canonicalization | Low | 70% |
| PATH-002 | Potential double-extraction in tar archive handling | Low | 60% |

**Risk Rating:** LOW - The backup and storage system is well-designed with path containment checks. The identified issues are minor and unlikely to be exploitable in practice.

---

## Finding PATH-001: isWithinDirectory Uses String Prefix Without Path Canonicalization

**Severity:** Low
**Confidence:** 70%
**File:** internal/backup/manager.go:608
**CWE:** CWE-22 (Path Traversal)
**File References:**
- `internal/backup/manager.go:608` (isWithinDirectory)
- `internal/backup/manager.go:457-470` (Delete method)
- `internal/backup/manager.go:473-483` (Get method)

### Description

The `isWithinDirectory()` function uses a simple string prefix check to verify that a file path is within a directory:

```go
// internal/backup/manager.go:608
func isWithinDirectory(path, dir string) bool {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return false
    }
    absDir, err := filepath.Abs(dir)
    if err != nil {
        return false
    }
    return len(absPath) > len(absDir) && absPath[:len(absDir)] == absDir
}
```

This approach is **insecure** because:

1. **No path cleaning:** The function does not call `filepath.Clean()` or resolve `..` components before comparison
2. **Symlink attack:** If an attacker can create a symlink within the backups directory (e.g., `backups/evil -> /etc`), the check would fail to detect this

### Attack Scenario

```bash
# If attacker can create files in backups directory:
cd /var/lib/anubis/backups
ln -s /etc/passwd malicious_backup

# Now isWithinDirectory("backups/malicious_backup", "backups/") returns true
# because the string prefix matches, even though it points outside
```

### Code Called By

This function is used in two security-critical paths:

1. **Delete operation** (`internal/backup/manager.go:457-470`):
   ```go
   func (m *Manager) Delete(filename string) error {
       path := filepath.Join(m.backupsDir, filename)
       if !isWithinDirectory(path, m.backupsDir) {  // Security check
           return fmt.Errorf("invalid backup path")
       }
       if err := os.Remove(path); err != nil {
           return fmt.Errorf("failed to delete backup: %w", err)
       }
       // ...
   }
   ```

2. **Get operation** (`internal/backup/manager.go:473-483`):
   ```go
   func (m *Manager) Get(filename string) (*Backup, error) {
       path := filepath.Join(m.backupsDir, filename)
       if !isWithinDirectory(path, m.backupsDir) {  // Security check
           return nil, fmt.Errorf("invalid backup path")
       }
       return m.readBackupFile(path)
   }
   ```

### Why This Is Low Risk

1. **Windows is less susceptible:** Windows does not typically allow creating symlinks without elevated privileges
2. **Backups directory ownership:** The backups directory (`dataDir/backups`) is created with `0700` permissions and owned by the application user (line 142)
3. **No user input to path:** Filenames are not user-controlled; they're generated server-side as `anubis_backup_YYYYMMDD_HHMMSS.json[.gz]`
4. **Defensive programming:** Even if bypassed, the actual file read operations go through `readBackupFile()` which validates JSON structure

### Remediation

Replace string comparison with canonical path comparison:

```go
func isWithinDirectory(path, dir string) bool {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return false
    }
    absDir, err := filepath.Abs(dir)
    if err != nil {
        return false
    }
    
    // Clean both paths to resolve .. components and symlinks
    cleanPath := filepath.Clean(absPath)
    cleanDir := filepath.Clean(absDir)
    
    // Use filepath.HasDir or check with separator
    if cleanPath == cleanDir {
        return true
    }
    
    // Ensure path starts with directory + separator
    sep := string(filepath.Separator)
    return strings.HasPrefix(cleanPath, cleanDir+sep)
}
```

For even stronger security, use `os.ReadDir` to enumerate and verify the file actually exists within the directory.

### References

- https://cwe.mitre.org/data/definitions/22.html
- https://owasp.org/www-community/attacks/Path_Traversal

---

## Finding PATH-002: Potential Double-Extraction in Tar Archive Handling

**Severity:** Low
**Confidence:** 60%
**File:** internal/backup/manager.go:682-724
**CWE:** CWE-22 (Path Traversal)
**File References:**
- `internal/backup/manager.go:682` (ImportFromTar)
- `internal/backup/manager.go:686-696` (tar entry reading)

### Description

The `ImportFromTar()` function reads entries from a tar archive:

```go
// internal/backup/manager.go:682-696
func (m *Manager) ImportFromTar(storage RestoreStorage, r io.Reader, opts RestoreOptions) error {
    tr := tar.NewReader(r)

    for {
        header, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        if filepath.Ext(header.Name) != ".json" {  // Only checks extension
            continue
        }
        // ... reads and processes entry
    }
}
```

### Potential Issues

1. **Path traversal in tar entries:** If a malicious tar file contains entries with `../` in the name (e.g., `../../../etc/passwd`), the current check only verifies the extension is `.json`, not that the path is safe.

2. **Double-extension bypass:** An entry named `evil.json/../../../etc/passwd` would pass the extension check.

### Why This Is Low Risk

1. **Internal use only:** `ImportFromTar` is not exposed via the REST API - it's used for internal restore operations
2. **Backup source validation:** Backups are created internally and signed with a checksum (verified at line 710)
3. **The tar is not extracted to filesystem:** The tar entries are read into memory and unmarshaled as JSON, not written to disk

### Code Flow Analysis

```go
// internal/backup/manager.go:699-707
data := make([]byte, header.Size)
if _, err := io.ReadFull(tr, data); err != nil {
    return err
}

var backup Backup
if err := json.Unmarshal(data, &backup); err != nil {  // Parsed as JSON, not file
    return err
}

// Verify checksum
if err := m.verifyChecksum(&backup); err != nil {
    return err
}
```

The tar data is read as raw bytes and parsed as JSON, not written to the filesystem. This means path traversal in the tar entry name doesn't lead to file overwrite.

### Remediation

Add path validation for tar entries:

```go
// Validate tar entry name doesn't contain path traversal
if strings.Contains(header.Name, "..") {
    slog.Warn("rejecting tar entry with path traversal", "name", header.Name)
    continue
}
```

Or use `archive/tar` with `FormatGNU` or `FormatPAX` and validate the Clean() path.

### References

- https://cwe.mitre.org/data/definitions/22.html
- CVE-2021-43297 (related tar path traversal)

---

## Positive Security Findings

The backup and storage system has several security strengths:

### Strengths

1. **Namespace isolation via key prefixes** (`internal/storage/engine.go:696-708`):
   ```go
   // Souls are stored with workspace-prefixed keys
   key := fmt.Sprintf("%s/souls/%s", soul.WorkspaceID, soul.ID)
   ```
   This prevents cross-workspace data access even if key iteration were possible.

2. **Checksum verification** (`internal/backup/manager.go:501-512`):
   ```go
   func (m *Manager) verifyChecksum(backup *Backup) error {
       expected := backup.Checksum
       actual, err := m.calculateChecksum(backup)
       // ...
       if expected != actual {
           return fmt.Errorf("checksum mismatch")
       }
   }
   ```

3. **Atomic file writes** (`internal/backup/manager.go:515-551`):
   ```go
   // Write to temp file first, then atomic rename
   tmpPath := path + ".tmp"
   // ... write to tmp ...
   return os.Rename(tmpPath, path)  // Atomic on POSIX, better on Win
   ```

4. **Strict file extension validation** (`internal/backup/manager.go:597-599`):
   ```go
   func isBackupFile(name string) bool {
       return len(name) > 7 && name[:7] == "anubis_" && 
              (filepath.Ext(name) == ".json" || filepath.Ext(name) == ".gz")
   }
   ```

5. **Gzip magic byte detection** (`internal/backup/manager.go:601-606`):
   ```go
   func isGzipped(file *os.File) bool {
       buf := make([]byte, 2)
       file.Read(buf)
       file.Seek(0, 0)
       return buf[0] == 0x1f && buf[1] == 0x8b  // Validates gzip magic
   }
   ```

6. **Backup scope limitations** (`internal/backup/manager.go:64-81`): The backup explicitly excludes sensitive data:
   - `secrets.enc` (encrypted keys)
   - WAL files
   - Dashboard assets

7. **Request body size limits** (`internal/api/rest.go:21`): 1MB limit prevents memory exhaustion.

---

## Recommendations

1. **Low Priority:** Update `isWithinDirectory` to use `filepath.Clean()` before comparison (PATH-001)
2. **Low Priority:** Add `..` validation in `ImportFromTar` as defense-in-depth (PATH-002)

---

*Generated by sc-path-traversal security scanner*