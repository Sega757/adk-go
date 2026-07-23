# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-23 - Anti-Caching HTTP Security Headers for Sensitive Agent Traces & Session Data
**Vulnerability:** HTTP responses containing sensitive AI agent execution traces, session history, and generated artifacts did not restrict downstream caching, leaving them vulnerable to intermediate proxy caching or local browser history exposure.
**Learning:** Standard security headers (such as `X-Frame-Options` or `X-Content-Type-Options`) alone do not govern browser/proxy caching behaviors. Explicit anti-caching directives are necessary for REST endpoints containing dynamically-generated trace logs, session states, and artifact payloads.
**Prevention:** Always enforce global middleware injecting anti-caching headers (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) in all agent REST API servers and web UI interfaces, and verify them via automated integration tests.
