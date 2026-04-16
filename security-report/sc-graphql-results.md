# GraphQL Injection Scan Results (sc-graphql)

**Scanner:** sc-graphql  
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

### GraphQL Usage

No GraphQL library or server is used in this project.

**Not found:**
- `typeDefs`, `resolvers`, `ApolloServer`, `graphql-yoga`
- `buildSchema`, `makeExecutableSchema`
- `introspection`, `depthLimit`, `costAnalysis`
- `__schema`, `__type`
- `.graphql`, `.gql` schema files

### API Architecture

The project uses:
- **REST API** (`internal/api/rest.go`) - Custom router implementation
- **gRPC API** (`internal/grpcapi/server.go`) - For internal/probe communication
- **WebSocket** (`internal/api/websocket.go`) - For real-time updates (Duat)
- **MCP Server** (`internal/api/mcp.go`) - Model Context Protocol for AI integration

### Findings

**NO GRAPHQL INJECTION VULNERABILITIES FOUND**

The project has no GraphQL endpoint or server. The API is purely REST/gRPC.

### Verdict

**CLEAN - No GraphQL attack surface exists**

---

## Common False Positives Eliminated

1. **REST API** - Not GraphQL, different attack surface
2. **gRPC** - Uses protobuf, not GraphQL query language
3. **No schema files** - No `.graphql` or `.gql` files in codebase

## References

- CWE-89: SQL Injection (for GraphQL injection variants)
- CWE-200: Information Disclosure (for introspection)
- https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html