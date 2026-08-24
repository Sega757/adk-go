# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-07-27 - [Dynamic Scheme-Enforced CORS Origin Validation]
**Vulnerability:** Unconditional static CORS `Access-Control-Allow-Origin` mapping or naive substring validation can allow malicious origins (e.g. `localhost.attacker.com` or mismatched HTTP/HTTPS schemes) to bypass cross-origin restrictions or corrupt CORS caches.
**Learning:** Setting CORS `Access-Control-Allow-Origin` statically without dynamically validating request `Origin` headers violates secure origin rules. Furthermore, not providing the `Vary: Origin` header can lead to intermediate proxy or browser cache poisoning.
**Prevention:** Always dynamically validate the incoming request `Origin` header against an exact-match list or strict schemed/scheme-less configurations, and enforce identical schemes if specified. Always include the `Vary: Origin` header to safeguard caches.
