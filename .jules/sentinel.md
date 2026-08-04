# Sentinel Security Journal

This journal tracks critical security learnings, vulnerability discoveries, and custom prevention guidelines for the ADK Go framework.

## 2026-07-26 - [Strict META-CORE Input Validation]
**Vulnerability:** Numerical out-of-bounds, NaN (Not-a-Number) values, or nil dereferences in decision packets could bypass reasoning and safety validation steps. Specifically, invalid/NaN `Confidence` or `VulnerabilityScore` metrics could cause undefined validation behavior, possibly allowing unsafe executions to proceed undetected.
**Learning:** Checking for standard bounds (`0.0 <= score <= 1.0`) is not sufficient because standard float comparison operators (`<`, `>`) evaluate to false when one operand is `NaN`. Thus, `NaN` values bypass threshold checks unless explicitly handled.
**Prevention:** Use `math.IsNaN()` to explicitly check and reject any `NaN` values in floating-point security metrics before performing range comparisons, and always validate input structures for `nil` pointers before field dereferencing.

## 2026-07-27 - [Secure Dynamic CORS Validation & Cache Poisoning Mitigation]
**Vulnerability:** Permissive or basic CORS configurations (e.g., static Allowed-Origin reflection or scheme-less/cross-scheme matches) can expose APIs to cross-scheme origin bypass and HTTP cache poisoning.
**Learning:** Hardcoding a single origin can fail when multiple/different schemes are used, whereas dynamically reflecting the Origin header without strict origin validation leads to CSRF or unauthorized cross-origin access. Furthermore, not setting `Vary: Origin` can cause cache poisoning in downstream CDN/proxies.
**Prevention:** Use a dynamic CORS validator that derives acceptable origins from configuration, strictly checks schemes (if configured), applies `Vary: Origin` to responses, and rejects unmatched origins with 403 Forbidden.
