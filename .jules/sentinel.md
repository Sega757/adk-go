# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-08-05 - [Secure CORS Origin Validation & Cache Poisoning Mitigation]
**Vulnerability:** Permissive or statically reflected CORS configurations can expose APIs to unauthorized cross-origin requests, cross-scheme origin bypass, or HTTP cache poisoning.
**Learning:** Hardcoding a dynamic origin reflections without exact string verification is vulnerable to host-suffix/prefix bypass. Additionally, omission of the `Vary: Origin` header can cause intermediate caching proxies to cache CORS response headers for one origin and mistakenly serve them to another. Finally, responding with `403 Forbidden` on origin mismatch breaks standard HTTP access patterns for same-origin or non-browser clients; standard compliant servers should instead simply omit the CORS headers and proceed.
**Prevention:** Build a dynamic CORS validator that matches origins against a whitelist derived from allowed configuration hosts (supporting both schemed and scheme-less options), applies `Vary: Origin` to responses whenever an `Origin` header is present, and safely lets the request proceed without CORS headers on mismatch instead of returning an active block.
