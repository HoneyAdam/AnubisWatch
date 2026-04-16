# Server-Side Template Injection (SSTI) Scan Results (sc-ssti)

**Scanner:** sc-ssti  
**Date:** 2026-04-16  
**Target:** D:\CODEBOX\PROJECTS\AnubisWatch

---

## Summary

| Category | Result |
|----------|--------|
| Files Scanned | cmd/, internal/, web/ |
| Vulnerabilities Found | 0 |
| Informational Findings | 1 |
| Confidence | High |

---

## Analysis

### Template Engine Usage

The project does NOT use standard template engines:
- **No Jinja2** (`render_template_string`, `jinja2.Template`)
- **No Twig** (`$twig->createTemplate`)
- **No Freemarker/Velocity**
- **No ERB** (`ERB.new`)
- **No Handlebars** (server-side)
- **No Pug/EJS/Nunjucks**

### Custom Template Replacement Pattern

**File:** `internal/alert/dispatchers.go:876-881`

The alert dispatchers use a **simple string replacement pattern**, not a real template engine:

```go
// Simple template replacement (NOT a template engine)
message := template
message = strings.ReplaceAll(message, "{{.SoulName}}", event.SoulName)
message = strings.ReplaceAll(message, "{{.Status}}", string(event.Status))
message = strings.ReplaceAll(message, "{{.Message}}", event.Message)
message = strings.ReplaceAll(message, "{{.Severity}}", string(event.Severity))
message = strings.ReplaceAll(message, "{{.Time}}", event.Timestamp.Format(time.RFC3339))
```

### Informational Finding: Template Variable Interpolation in Journey Executor

**File:** `internal/journey/executor.go:381-387`

The journey executor interpolates `${variable}` placeholders from response data:

```go
// interpolateVariables replaces ${variable} placeholders with values
func (e *Executor) interpolateVariables(s string, variables map[string]string) string {
    for k, v := range variables {
        s = strings.ReplaceAll(s, "${"+k+"}", v)
    }
    return s
}
```

**Assessment:** This is a **false positive** for SSTI. The `variables` map is populated from structured Go types (`core.Judgment`, `core.ExtractionRule`), not from raw user input that could contain template syntax. The values come from:
- HTTP response bodies (structured JSON, not template code)
- HTTP response headers
- Cookie values
- Regex extractions

These are data values, not executable template code. An attacker cannot inject `{{7*7}}` or `${os.popen('id')}` through this path because the data flows through structured extraction rules, not raw template rendering.

### No `text/template` with User Input

**File:** `internal/api/rest.go:2146`
```go
w.StartTime, _ = time.Parse(time.RFC3339, str)
```

This uses `time.Parse()`, not `text/template.Parse()`.

### Findings

**NO SERVER-SIDE TEMPLATE INJECTION VULNERABILITIES FOUND**

The alert templates use simple string replacement (`strings.ReplaceAll`), not a template engine. There is no path where user input becomes part of template source code before rendering.

### Verdict

**CLEAN - No SSTI attack surface**

The project uses `strings.ReplaceAll` for template-like substitution, not template engines that compile user input as code. String replacement cannot evaluate template expressions like `{{7*7}}`.

---

## Common False Positives Eliminated

1. **Simple string replacement** - `strings.ReplaceAll` is NOT a template engine, no code evaluation
2. **Journey variable interpolation** - Data from structured extraction rules, not raw user input
3. **`time.Parse`** - Time parsing function, not template parsing
4. **No actual template compilation** - No use of `template.Parse()`, `Template()`, etc.

## References

- CWE-1336: Improper Neutralization of Special Elements Used in a Template Engine
- https://portswigger.net/web-security/server-side-template-injection