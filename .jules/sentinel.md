# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2025-02-18 - Prevent Browser Caching of Sensitive REST API Data
**Vulnerability:** REST API responses containing sensitive user session history, execution traces, and artifacts could be cached by web browsers or intermediate proxies if explicit anti-caching HTTP headers are missing.
**Learning:** In the absence of explicit `Cache-Control`, `Pragma`, and `Expires` headers, browsers and intermediate proxies may heuristically cache JSON responses. This exposes confidential session state and execution telemetry to unauthorized users sharing a terminal or accessing proxy logs.
**Prevention:** Always enforce global anti-caching security headers (`Cache-Control: no-store, no-cache, must-revalidate, proxy-revalidate`, `Pragma: no-cache`, `Expires: 0`) in the REST API's global middleware or routers serving sensitive/telemetry data, and verify them in automated integration or unit tests.
