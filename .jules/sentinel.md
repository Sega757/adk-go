# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-17 - Sensitive REST API Data Cache Control Enforcements
**Vulnerability:** Default HTTP responses from the REST API did not specify anti-caching headers. This allowed browsers or intermediate proxies to potentially cache sensitive session histories, telemetry traces, and execution artifacts.
**Learning:** Standard security headers (X-Frame-Options, X-Content-Type-Options, etc.) were applied in custom HTTP middleware, but anti-caching directives were omitted. Telemetry traces and agent states should never be cached as they contain live session details and potentially sensitive parameters/data.
**Prevention:** Always enforce global anti-caching headers (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) in the REST API handler middleware. Validate their presence across all routes using automated integration tests.
